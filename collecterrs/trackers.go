package collecterrs

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

// ErrorVarTracker is responsible for finding returned errors that were saved in a variable
type ErrorVarTracker struct {
	errorVars map[string]string // Variable name → error code
}

func NewErrorVarTracker() *ErrorVarTracker {
	return &ErrorVarTracker{
		errorVars: make(map[string]string),
	}
}

func (et *ErrorVarTracker) Track(n ast.Node) {
	switch stmt := n.(type) {
	case *ast.AssignStmt:
		for _, expr := range stmt.Rhs {
			code := et.getErrorCode(expr)
			if code != "" {
				for _, lhs := range stmt.Lhs {
					if ident, ok := lhs.(*ast.Ident); ok {
						et.errorVars[ident.Name] = code
					}
				}
			}
		}
	}
}

func (et *ErrorVarTracker) getErrorCode(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.SelectorExpr:
		if strings.HasSuffix(e.Sel.Name, "Error") {
			if ident, ok := e.X.(*ast.Ident); ok && strings.HasPrefix(ident.Name, "errs") {
				return strings.TrimSuffix(e.Sel.Name, "Error")
			}
		}
		// Recursively check nested expressions
		return et.getErrorCode(e.X)
	case *ast.CallExpr:
		if sel, ok := e.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "WithDetails" {
			// Extract code from the base error before WithDetails
			code := et.getErrorCode(sel.X)
			if len(e.Args) > 0 {
				code = fmt.Sprintf("%s (%s)", code, strings.Join(extractMapKeys(e.Args[0]), ","))
			}
			return code
		}
		return ""
	case *ast.Ident:
		// Return code if the variable is already being tracked
		return et.errorVars[e.Name]
	default:
		return ""
	}
}

// Extracts keys from map[string]string literal for WithDetails errors
func extractMapKeys(expr ast.Expr) []string {
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil
	}

	// Extract the map value type (for example, "string" from map[string]string)
	var valueType string
	if mapType, ok := lit.Type.(*ast.MapType); ok {
		switch v := mapType.Value.(type) {
		case *ast.Ident:
			valueType = v.Name // Basic types: string, int, etc.
		case *ast.SelectorExpr:
			// For types from other packages (for example, time.Time)
			if pkg, ok := v.X.(*ast.Ident); ok {
				valueType = pkg.Name + "." + v.Sel.Name
			}
		default:
			valueType = "unknown"
		}
	} else {
		return nil
	}

	var params []string
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		// Extract the key
		var keyName string
		switch key := kv.Key.(type) {
		case *ast.BasicLit:
			if key.Kind == token.STRING {
				keyName = strings.Trim(key.Value, `"`)
			}
		case *ast.Ident:
			keyName = key.Name
		}

		if keyName != "" {
			params = append(params, fmt.Sprintf("%s:%s", keyName, valueType))
		}
	}
	return params
}

// ErrorHandler is responsible for checking returned errors
// If there is error handling in the code via errors.Is, it marks it
// so that the analyzer can later remove it from the returned errors
type ErrorHandler struct {
	handledErrors map[string]bool // Codes of handled errors
}

func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		handledErrors: make(map[string]bool),
	}
}

func (eh *ErrorHandler) Inspect(node ast.Node, errtracker *ErrorVarTracker) {
	if stmt, ok := node.(*ast.IfStmt); ok {
		// Handle combined conditions with logical operators
		if binExpr, ok := stmt.Cond.(*ast.BinaryExpr); ok {
			// Handle pattern: err != nil && !errors.Is(err, errsUsers.UserNotFoundError)
			if binExpr.Op == token.LAND {
				// Check if the right side is a unary expression with '!' operator
				if unaryExpr, ok := binExpr.Y.(*ast.UnaryExpr); ok && unaryExpr.Op == token.NOT {
					// Check if the operand is an errors.Is call
					if call, ok := unaryExpr.X.(*ast.CallExpr); ok && isErrorsIsCall(call) {
						if len(call.Args) < 2 {
							return
						}
						// Second argument of errors.Is is target
						if targetCall, ok := call.Args[1].(*ast.SelectorExpr); ok {
							if strings.HasSuffix(targetCall.Sel.Name, "Error") {
								if ident, ok := targetCall.X.(*ast.Ident); ok && strings.HasPrefix(ident.Name, "errs") {
									eh.handledErrors[strings.TrimSuffix(targetCall.Sel.Name, "Error")] = true
								}
							}
						}

						// If target is a variable (for example, err)
						if targetIdent, ok := call.Args[1].(*ast.Ident); ok {
							if code, exists := errtracker.errorVars[targetIdent.Name]; exists {
								eh.handledErrors[code] = true
							}
						}
					}
				}
			}
		}

		// Handle errors.Is(err, target)
		if call, ok := stmt.Cond.(*ast.CallExpr); ok {
			if isErrorsIsCall(call) {
				if len(call.Args) < 2 {
					return
				}
				// Second argument of errors.Is is target
				// If target is a function call (for example, usersErrs.TokenRefreshForbiddenError())
				if targetCall, ok := call.Args[1].(*ast.SelectorExpr); ok {
					if strings.HasSuffix(targetCall.Sel.Name, "Error") {
						if ident, ok := targetCall.X.(*ast.Ident); ok && strings.HasPrefix(ident.Name, "errs") {
							eh.handledErrors[strings.TrimSuffix(targetCall.Sel.Name, "Error")] = true
						}
					}
				}

				// If target is a variable (for example, err)
				if targetIdent, ok := call.Args[1].(*ast.Ident); ok {
					if code, exists := errtracker.errorVars[targetIdent.Name]; exists {
						eh.handledErrors[code] = true
					}
				}
			} else if isCustomErrorIsCall(call) {
				// handle custom ServiceError.Is(err)
				if len(call.Args) < 1 {
					return
				}

				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if strings.HasSuffix(sel.Sel.Name, "Is") {
						if errorSel, ok := sel.X.(*ast.SelectorExpr); ok {
							if strings.HasSuffix(errorSel.Sel.Name, "Error") {
								if ident, ok := errorSel.X.(*ast.Ident); ok && strings.HasPrefix(ident.Name, "errs") {
									errorName := strings.TrimSuffix(errorSel.Sel.Name, "Error")
									//fmt.Printf("[DEBUG] Found custom error check: %s.%s.Is() - marking as handled\n",
									//	ident.Name, errorSel.Sel.Name)
									eh.handledErrors[errorName] = true

									// Also mark the error with details as handled
									detailsPattern := errorName + " ("
									for k := range eh.handledErrors {
										if strings.HasPrefix(k, detailsPattern) {
											//fmt.Printf("[DEBUG] Also marking as handled with details: %s\n", k)
											eh.handledErrors[k] = true
										}
									}

									// Check full name with package prefix
									if strings.HasPrefix(ident.Name, "errs") {
										serviceName := strings.TrimPrefix(ident.Name, "errs")
										if serviceName != "" {
											servicePrefix := strings.ToLower(string(serviceName[0])) + serviceName[1:]
											fullErrorName := servicePrefix + "." + errorName
											//fmt.Printf("[DEBUG] Also marking as handled with service prefix: %s\n", fullErrorName)
											eh.handledErrors[fullErrorName] = true

											// Also mark the error with service prefix and details as handled
											detailsPattern = fullErrorName + " ("
											for k := range eh.handledErrors {
												if strings.HasPrefix(k, detailsPattern) {
													//fmt.Printf("[DEBUG] Also marking as handled with service prefix and details: %s\n", k)
													eh.handledErrors[k] = true
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

func isErrorsIsCall(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		return sel.Sel.Name == "Is" && isErrorsPackage(sel.X)
	}
	return false
}

func isErrorsPackage(expr ast.Expr) bool {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name == "errors"
	}
	return false
}

// isCustomErrorIsCall checks if this is a call to a custom Error.Is method
// ex: errs.MaxCodeChecksExceededError.Is(err)
// This is a general solution for any error object ending with "Error" that has an .Is method
func isCustomErrorIsCall(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if sel.Sel.Name == "Is" {
			if errorSel, ok := sel.X.(*ast.SelectorExpr); ok {
				return strings.HasSuffix(errorSel.Sel.Name, "Error")
			}
		}
	}
	return false
}

// ProviderCall is responsible for tracking calls to adjacent dependencies through a provider
// errors from these calls should be inserted later
type ProviderCall struct {
	Provider string
	Method   string
}

func (p ProviderCall) String() string {
	return fmt.Sprintf("[%s].%s", p.Provider, p.Method)
}

type ProviderTracker struct {
	Calls map[string][]ProviderCall // Variable name → list of calls
}

func NewProviderTracker() *ProviderTracker {
	return &ProviderTracker{
		Calls: make(map[string][]ProviderCall),
	}
}

func (t *ProviderTracker) Track(node ast.Node, ua *UsecaseAnalysis) {
	switch stmt := node.(type) {
	case *ast.AssignStmt:
		for _, expr := range stmt.Rhs {
			if call, ok := expr.(*ast.CallExpr); ok {
				// Processing function calls that return providers
				funcName := getFuncName(call)
				if funcName != "" {
					if providers, exists := ua.returnedProviders[funcName]; exists {
						for _, lhs := range stmt.Lhs {
							if ident, ok := lhs.(*ast.Ident); ok {
								t.Calls[ident.Name] = append(t.Calls[ident.Name], providers...)
							}
						}
					}
				}
				// Direct provider calls
				if provider, method, ok := extractProviderMethod(call); ok {
					for _, lhs := range stmt.Lhs {
						if ident, ok := lhs.(*ast.Ident); ok {
							t.Calls[ident.Name] = append(t.Calls[ident.Name], ProviderCall{
								Provider: provider,
								Method:   method,
							})
						}
					}
				}
			}
		}
	}
}

func getFuncName(call *ast.CallExpr) string {
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		return fn.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", exprToString(fn.X), fn.Sel.Name)
	default:
		return ""
	}
}

func extractProviderMethod(call *ast.CallExpr) (string, string, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", "", false
	}

	// Expecting chain: u.Providers.<Provider>.<Method>
	if xSel, ok := sel.X.(*ast.SelectorExpr); ok {
		if providersSel, ok := xSel.X.(*ast.SelectorExpr); ok {
			if providersSel.Sel.Name == "Providers" {
				return xSel.Sel.Name, sel.Sel.Name, true
			}
		}
	}
	return "", "", false
}
