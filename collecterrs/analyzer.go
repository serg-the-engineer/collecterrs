package collecterrs

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/tools/go/packages"
)

type UsecaseAnalysis struct {
	returnedProviders map[string][]ProviderCall
}

func NewUsecaseAnalysis() *UsecaseAnalysis {
	return &UsecaseAnalysis{
		returnedProviders: make(map[string][]ProviderCall),
	}
}

func (ua *UsecaseAnalysis) Analyze(servicesPath string, moduleName string, verbose bool) (map[string]map[string][]string, error) {
	var services []string

	entries, err := os.ReadDir(servicesPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			services = append(services, toCamelCase(entry.Name()))
		}
	}

	errs := map[string]map[string][]string{}
	handledErrs := map[string]map[string]map[string]bool{}

	// first collect errors for each service separately, save references to providers except storage
	for _, service := range services {
		// Use the full module path for the package
		storageErrs, _, err := ua.AnalyzePkg(fmt.Sprintf("%s/services/%s/storage", moduleName, service), nil, verbose)
		if err != nil {
			return nil, err
		}
		serviceErrs, handleds, err := ua.AnalyzePkg(fmt.Sprintf("%s/services/%s/usecase", moduleName, service), storageErrs, verbose)
		errs[service] = serviceErrs
		handledErrs[service] = handleds
	}

	ua.LinkProviderErrors(errs, handledErrs, verbose)

	return errs, nil
}

func (ua *UsecaseAnalysis) LinkProviderErrors(errs map[string]map[string][]string, handledErrs map[string]map[string]map[string]bool, verbose bool) {
	// insert errors from called providers
	re := regexp.MustCompile(`^\[([a-zA-Z0-9]+)\]\.([a-zA-Z0-9]+)$`)
	linkExists := true
	for linkExists {
		linkExists = false
		for serviceName, serviceErrs := range errs {
			for usecaseName, usecaseErrs := range serviceErrs {
				newErrs := []string{}
				for _, usecaseErr := range usecaseErrs {
					matches := re.FindStringSubmatch(usecaseErr)
					if len(matches) < 1 { // not a provider - save the error
						newErrs = append(newErrs, usecaseErr)
						continue
					}
					nestedServiceName := toCamelCase(matches[1])
					s, ok := errs[nestedServiceName]
					if !ok { // external provider, not our service - remove it, it definitely won't return named errors
						if verbose {
							fmt.Printf("[DEBUG] Remove external provider %s from %s.%s\n", usecaseErr, serviceName, usecaseName)
						}
						continue
					}
					nestedErrs, ok := s[matches[2]]
					if !ok {
						if verbose {
							fmt.Printf("[DEBUG] Remove service provider %s from %s.%s [no nested errs]\n", usecaseErr, serviceName, usecaseName)
						}
						continue
					}
					for _, nestedErr := range nestedErrs {
						errWithoutDetails := strings.Split(nestedErr, " ")[0]
						if handledErrs[serviceName][usecaseName][errWithoutDetails] {
							if verbose {
								fmt.Printf("[DEBUG] Skip handled error %s.%s.%s \n", serviceName, usecaseName, nestedErr)
							}
							continue
						}
						nestedMatches := re.FindStringSubmatch(nestedErr)
						if len(nestedMatches) > 1 {
							linkExists = true // handle nested provider call
							newErrs = append(newErrs, nestedErr)
						} else {
							newErrs = append(newErrs, nestedServiceName+"."+nestedErr)
						}
					}
				}
				errs[serviceName][usecaseName] = newErrs
			}
		}
	}
}

func (ua *UsecaseAnalysis) AnalyzePkg(pkgPath string, extraErrs map[string][]string, verbose bool) (map[string][]string, map[string]map[string]bool, error) {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Extract the module name from the package path
	parts := strings.Split(pkgPath, "/")
	var projectDir string
	if len(parts) > 0 {
		// Find the directory that contains the go.mod file for this module
		moduleDir := findModuleDir(cwd, parts[0])
		if moduleDir != "" {
			projectDir = moduleDir
		} else {
			projectDir = cwd
		}
	} else {
		projectDir = cwd
	}

	if verbose {
		fmt.Printf("[DEBUG] Using project directory: %s for package: %s\n", projectDir, pkgPath)
	}

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedSyntax,
		Dir:  projectDir,
	}
	isUsecase := strings.HasSuffix(pkgPath, "usecase")
	isStorage := strings.HasSuffix(pkgPath, "storage")

	pkgs, err := packages.Load(cfg, pkgPath)
	if verbose {
		for _, pkg := range pkgs {
			fmt.Printf("Package: %s\n", pkg.ID)
			fmt.Printf("Files: %v\n", pkg.GoFiles)
			fmt.Printf("Syntax trees: %d\n", len(pkg.Syntax))
		}
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load package: %w", err)
	}

	results := make(map[string][]string)
	handledErrors := make(map[string]map[string]bool)

	for _, pkg := range pkgs {
		pf := collectPackageFunctions(pkgs[0])
		if verbose {
			fmt.Printf("[DEBUG] Package functions: %+v\n", pf)
		}
		for _, file := range pkg.Syntax {
			filename := getFilename(file, pkg.Fset)
			ast.Inspect(file, func(n ast.Node) bool {
				fn, ok := n.(*ast.FuncDecl)
				if !ok || (isUsecase && !isUsecaseMethod(filename, fn)) {
					return true
				}

				var errors []string
				errtracker := NewErrorVarTracker()
				errhandler := NewErrorHandler()
				providerTracker := NewProviderTracker()

				if verbose {
					fmt.Println("Start analyze " + fn.Name.Name)
				}
				ua.analyzeFunction(fn, pf, make(map[string]bool), errtracker, errhandler, providerTracker, &errors)

				if verbose {
					fmt.Printf("[DEBUG] Function: %s\n", fn.Name.Name)
					fmt.Printf("[DEBUG] Tracked errors: %v\n", errtracker.errorVars)
					fmt.Printf("[DEBUG] Handled errors: %v\n", errhandler.handledErrors)
					fmt.Printf("[DEBUG] Provider calls: %v\n", providerTracker.Calls)
					fmt.Printf("[DEBUG] Result errors: %v\n", errors)
				}

				name := fn.Name.Name
				if isStorage {
					name = fmt.Sprintf("[Storage].%s", name)
				}
				for _, e := range unique(errors) {
					if nestedErrs, ok := extraErrs[e]; ok {
						for _, er := range nestedErrs {
							if !errhandler.handledErrors[er] {
								results[name] = append(results[name], er)
							}
						}
					} else if !strings.HasPrefix(e, "[Storage]") {
						results[name] = append(results[name], e)
					}
				}
				handledErrors[name] = errhandler.handledErrors
				return true
			})
		}
	}
	return results, handledErrors, nil
}

func isUsecaseMethod(filename string, fn *ast.FuncDecl) bool {
	return fn.Recv != nil &&
		strings.EqualFold(fn.Name.Name, filename)
}

func getFilename(file *ast.File, fset *token.FileSet) string {
	name := fset.File(file.Pos()).Name()
	return strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))
}

type PackageFunctions struct {
	FuncDecls map[string]*ast.FuncDecl // Function name -> AST-node
	Methods   map[string]*ast.FuncDecl // Struct methods
}

// collect functions to analyze them when we encounter a call in the code
// currently only functions of the passed package
func collectPackageFunctions(pkg *packages.Package) *PackageFunctions {
	pf := &PackageFunctions{
		FuncDecls: make(map[string]*ast.FuncDecl),
		Methods:   make(map[string]*ast.FuncDecl),
	}

	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			switch fn := n.(type) {
			case *ast.FuncDecl:
				if fn.Recv != nil {
					// store by the scheme VariableName.Method to find during ast analysis
					//typeName := exprToString(fn.Recv.List[0].Type)
					typeVal := fn.Recv.List[0].Names[0]
					key := fmt.Sprintf("%s.%s", typeVal.Name, fn.Name.Name)
					pf.Methods[key] = fn
				} else {
					// usual function
					pf.FuncDecls[fn.Name.Name] = fn
				}
			}
			return true
		})
	}
	return pf
}

// toCamelCase converts a string to camelCase format
func toCamelCase(s string) string {
	// Split the string by delimiters (hyphen, underscore, space)
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_' || unicode.IsSpace(r)
	})

	if len(parts) == 0 {
		return ""
	}

	// First word starts with a lowercase letter
	result := strings.ToLower(parts[0])

	// Other words start with a capital letter
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += string(unicode.ToUpper(rune(parts[i][0]))) + strings.ToLower(parts[i][1:])
		}
	}

	return result
}

func exprToString(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return exprToString(t.X)
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	default:
		return fmt.Sprintf("%T", t)
	}
}

func (ua *UsecaseAnalysis) analyzeFunction(
	fn *ast.FuncDecl,
	pf *PackageFunctions,
	visited map[string]bool,
	errtracker *ErrorVarTracker,
	errhandler *ErrorHandler,
	providerTracker *ProviderTracker,
	errors *[]string,
) {
	if visited[fn.Name.Name] {
		return
	}
	visited[fn.Name.Name] = true

	// Collect information about errors
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		providerTracker.Track(n, ua)
		errtracker.Track(n)
		errhandler.Inspect(n, errtracker)
		return true
	})

	var funcProviders []ProviderCall
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.ReturnStmt:
			ua.checkReturnStatement(node, errtracker, providerTracker, errors)
			for _, expr := range node.Results {
				if ident, ok := expr.(*ast.Ident); ok {
					if calls, exists := providerTracker.Calls[ident.Name]; exists {
						funcProviders = append(funcProviders, calls...)
					}
				} else if call, ok := expr.(*ast.CallExpr); ok {
					if provider, method, ok := extractProviderMethod(call); ok {
						funcProviders = append(funcProviders, ProviderCall{provider, method})
					}
				}
			}
		case *ast.CallExpr:
			ua.analyzeCallExpression(node, pf, visited, errtracker, errhandler, providerTracker, errors)
		}
		return true
	})
	// Save function providers to cache
	ua.returnedProviders[fn.Name.Name] = funcProviders
}

func (ua *UsecaseAnalysis) analyzeCallExpression(
	call *ast.CallExpr,
	pf *PackageFunctions,
	visited map[string]bool,
	errtracker *ErrorVarTracker,
	errhandler *ErrorHandler,
	providerTracker *ProviderTracker,
	errors *[]string,
) {
	switch fn := call.Fun.(type) {
	case *ast.SelectorExpr:
		typeName := exprToString(fn.X)
		key := fmt.Sprintf("%s.%s", typeName, fn.Sel.Name)
		if decl, exists := pf.Methods[key]; exists {
			ua.analyzeFunction(decl, pf, visited, errtracker, errhandler, providerTracker, errors)
		}

	case *ast.Ident:
		// function call
		if decl, exists := pf.FuncDecls[fn.Name]; exists {
			ua.analyzeFunction(decl, pf, visited, errtracker, errhandler, providerTracker, errors)
		}
	}
	// Save provider calls
	if provider, method, ok := extractProviderMethod(call); ok {
		*errors = append(*errors, fmt.Sprintf("[%s].%s", provider, method))
	}
}

func (ua *UsecaseAnalysis) checkReturnStatement(
	ret *ast.ReturnStmt,
	errtracker *ErrorVarTracker,
	providerTracker *ProviderTracker,
	errors *[]string,
) {
	for _, expr := range ret.Results {
		switch e := expr.(type) {
		case *ast.SelectorExpr:
			if strings.HasSuffix(e.Sel.Name, "Error") {
				if ident, ok := e.X.(*ast.Ident); ok && strings.HasPrefix(ident.Name, "errs") {
					*errors = append(*errors, strings.TrimSuffix(e.Sel.Name, "Error"))
				}
			}
		case *ast.CallExpr:
			// Direct provider calls
			if provider, method, ok := extractProviderMethod(e); ok {
				*errors = append(*errors, fmt.Sprintf("[%s].%s", provider, method))
			}
			if code := errtracker.getErrorCode(e); code != "" {
				*errors = append(*errors, code)
			}

		case *ast.Ident:
			if code, exists := errtracker.errorVars[e.Name]; exists {
				*errors = append(*errors, code)
			}
			if calls, exists := providerTracker.Calls[e.Name]; exists {
				for _, call := range calls {
					*errors = append(*errors, call.String())
				}
			}
		}
	}
}

func unique(input []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range input {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// findModuleDir searches for a go.mod file that contains the specified module name
// starting from the given directory and going up to parent directories
func findModuleDir(startDir, moduleName string) string {
	// First check if the module is in the current directory
	if isModuleDir(startDir, moduleName) {
		return startDir
	}

	// Check if there's a subdirectory with the module name
	moduleSubDir := filepath.Join(startDir, moduleName)
	if isModuleDir(moduleSubDir, moduleName) {
		return moduleSubDir
	}

	// Check if there's a "project" subdirectory (for backward compatibility)
	projectDir := filepath.Join(startDir, "project")
	if isModuleDir(projectDir, moduleName) {
		return projectDir
	}

	// If not found, return empty string
	return ""
}

// isModuleDir checks if the directory contains a go.mod file with the specified module name
func isModuleDir(dir string, moduleName string) bool {
	goModPath := filepath.Join(dir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return false
	}

	// Check if the module name is in the go.mod file
	// This is a simple check, a more robust solution would parse the go.mod file
	return strings.Contains(string(content), "module "+moduleName) ||
		strings.Contains(string(content), "module github.com/"+moduleName) ||
		strings.Contains(string(content), "module "+strings.ReplaceAll(moduleName, "-", ""))
}
