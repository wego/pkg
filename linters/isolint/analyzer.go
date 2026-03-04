// Package isolint provides a Go analyzer that detects raw ISO code string
// literals and recommends using github.com/wego/pkg/currency and
// github.com/wego/pkg/iso/site package constants instead.
package isolint

import (
	"go/ast"
	"go/token"
	"strconv"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// skipPackages are package import paths whose source files should not be
// checked. These packages define the constants themselves and must use raw
// string literals.
var skipPackages = map[string]bool{
	"github.com/wego/pkg/currency": true,
	"github.com/wego/pkg/iso/site": true,
}

// skipMethods are method/function names whose string arguments are key or
// column names, not ISO code values. When a string literal appears as an
// argument to one of these calls, it is skipped to reduce false positives.
var skipMethods = map[string]bool{
	// HTTP framework methods — args are parameter names.
	"Query":        true,
	"QueryParam":   true,
	"Param":        true,
	"FormValue":    true,
	"GetQuery":     true,
	"DefaultQuery": true,
	"PostForm":     true,

	// ORM/DB methods — args are column names or SQL fragments.
	"Select": true,
	"Pluck":  true,
	"Omit":   true,

	// Custom filter methods — args are column names.
	"Equals":    true,
	"NotEquals": true,
}

// Analyzer is the isolint analyzer that checks for raw ISO code string literals.
var Analyzer = &analysis.Analyzer{
	Name:     "isolint",
	Doc:      "recommends using currency/site package constants over raw ISO code string literals",
	URL:      "https://github.com/wego/pkg/linters/isolint",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (any, error) {
	// Skip packages that define the constants themselves.
	// pass.Pkg may be nil under LoadModeSyntax in golangci-lint.
	if pass.Pkg != nil && skipPackages[pass.Pkg.Path()] {
		return nil, nil
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.BasicLit)(nil),
	}

	// WithStack gives us ancestor context so we can skip import paths
	// and arguments to known key-accepting methods.
	inspect.WithStack(nodeFilter, func(n ast.Node, push bool, stack []ast.Node) bool {
		if !push {
			return true
		}

		lit := n.(*ast.BasicLit)
		if lit.Kind != token.STRING {
			return true
		}

		// Fast path: currency codes are 3 chars ("XXX" = 5 bytes quoted),
		// site codes are 2 chars ("XX" = 4 bytes quoted).
		// Skip everything else before allocating via Unquote.
		vlen := len(lit.Value)
		if vlen != 4 && vlen != 5 {
			return true
		}

		// Import path literals (e.g. "io") are not ISO codes.
		if isImportPath(stack) {
			return true
		}

		// String arguments to ORM, HTTP, and filter methods are column
		// or parameter names (e.g. db.Select("id"), c.Query("to")),
		// not ISO code values.
		if isArgToSkipMethod(stack) {
			return true
		}

		value, err := strconv.Unquote(lit.Value)
		if err != nil {
			return true
		}

		// Route by length — currency and site codes can never overlap.
		switch len(value) {
		case 3:
			if IsCurrencyCode(value) {
				reportCurrencyDiagnostic(pass, lit, value)
			}
		case 2:
			if IsSiteCode(value) {
				reportSiteDiagnostic(pass, lit, value)
			}
		}

		return true
	})

	return nil, nil
}

// isImportPath reports whether the BasicLit at the top of the stack is the
// path of an import declaration.
func isImportPath(stack []ast.Node) bool {
	if len(stack) < 2 {
		return false
	}
	_, ok := stack[len(stack)-2].(*ast.ImportSpec)
	return ok
}

// isArgToSkipMethod reports whether the BasicLit at the top of the stack is
// a direct argument to a call expression whose method/function name appears
// in skipMethods.
func isArgToSkipMethod(stack []ast.Node) bool {
	if len(stack) < 2 {
		return false
	}
	call, ok := stack[len(stack)-2].(*ast.CallExpr)
	if !ok {
		return false
	}
	return skipMethods[callName(call)]
}

// callName extracts the method or function name from a call expression.
// For selector expressions (x.Method), it returns the method name.
// For plain identifiers (funcName), it returns the function name.
// Returns "" if the pattern doesn't match.
func callName(call *ast.CallExpr) string {
	switch fn := call.Fun.(type) {
	case *ast.SelectorExpr:
		return fn.Sel.Name
	case *ast.Ident:
		return fn.Name
	}
	return ""
}
