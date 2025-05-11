package collecterrs

import (
	"fmt"
	"go/ast"
	"log"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"golang.org/x/tools/go/packages"
)

type ServiceError struct {
	ServiceName string
	ErrorMap    map[string]string // Mapping: BaseName -> MethodName
}

func CollectServiceErrors(serviceName string, projectDir string) (*ServiceError, error) {
	pkgPath := "./..."
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedSyntax | packages.NeedFiles,
		Dir:  projectDir,
		Logf: log.Printf,
	}

	pkgs, err := packages.Load(cfg, pkgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load package: %w", err)
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found for path: %s", pkgPath)
	}
	pkg := pkgs[0]
	if len(pkg.Syntax) == 0 {
		return nil, fmt.Errorf("no syntax trees found in package: %s", pkg.PkgPath)
	}

	se := &ServiceError{ServiceName: serviceName, ErrorMap: make(map[string]string)}

	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if !ok || ts.Name.Name != cases.Title(language.English).String(serviceName)+"Errors" {
				return true
			}

			it, ok := ts.Type.(*ast.InterfaceType)
			if !ok {
				return true
			}

			for _, method := range it.Methods.List {
				methodName := method.Names[0].Name
				if strings.HasSuffix(methodName, "Error") {
					code := strings.TrimSuffix(methodName, "Error")
					se.ErrorMap[code] = methodName
				}
			}
			return true
		})
	}
	return se, nil
}
