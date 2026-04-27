// Package isolint provides a Go analyzer that detects raw ISO code string
// literals and recommends using github.com/wego/pkg/currency and
// github.com/wego/pkg/iso/site package constants instead.
package isolint

import (
	"flag"
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Settings configures the isolint analyzer. It is populated from the
// `linters.settings.custom.isolint.settings` block in .golangci.yml when
// running as a module plugin, or from CLI flags when running standalone.
type Settings struct {
	// SkipFields is a list of struct field names whose ISO-like string
	// literal values should not be flagged. This handles domain collisions
	// where a card scheme abbreviation (e.g. "MC" for MasterCard) shares
	// the form of an ISO 3166-1 alpha-2 code (Monaco).
	//
	// Two assignment shapes are matched syntactically:
	//   - x.Field = ...               (AssignStmt with SelectorExpr LHS)
	//   - StructType{Field: ...}      (KeyValueExpr inside a CompositeLit)
	SkipFields []string `json:"skip-fields"`
}

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

// Analyzer is the default isolint analyzer with empty settings. The
// standalone CLI in cmd/isolint and existing analystest drivers consume
// this. Plugin and configured-CLI callers should use NewAnalyzer.
var Analyzer = NewAnalyzer(Settings{})

// NewAnalyzer returns an isolint analyzer configured with the given
// settings. It is the entry point for the module plugin and any caller
// that needs per-invocation configuration.
//
// The returned analyzer also exposes a -skip-fields flag (comma-separated)
// on its FlagSet, seeded from s.SkipFields. The standalone CLI consumes
// this flag via singlechecker. Plugin and analystest callers do not parse
// flags, so the seeded default is what they see.
func NewAnalyzer(s Settings) *analysis.Analyzer {
	skipFields := stringSliceFlag(append([]string(nil), s.SkipFields...))

	a := &analysis.Analyzer{
		Name:     "isolint",
		Doc:      "recommends using currency/site package constants over raw ISO code string literals",
		URL:      "https://github.com/wego/pkg/linters/isolint",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
	a.Flags.Var(&skipFields, "skip-fields",
		"comma-separated struct field names whose ISO-like string values should not be flagged "+
			"(e.g. CardSchemes for card scheme abbreviations like \"MC\")")
	a.Run = func(pass *analysis.Pass) (any, error) {
		set := make(map[string]bool, len(skipFields))
		for _, f := range skipFields {
			if f != "" {
				set[f] = true
			}
		}
		return run(pass, set)
	}
	return a
}

// stringSliceFlag implements flag.Value over a comma-separated list,
// preserving order. Empty Set ("") clears the slice.
type stringSliceFlag []string

func (s *stringSliceFlag) String() string {
	if s == nil {
		return ""
	}
	return strings.Join(*s, ",")
}

func (s *stringSliceFlag) Set(v string) error {
	if v == "" {
		*s = nil
		return nil
	}
	*s = strings.Split(v, ",")
	return nil
}

// Compile-time assertion that stringSliceFlag implements flag.Value.
var _ flag.Value = (*stringSliceFlag)(nil)

func run(pass *analysis.Pass, skipFields map[string]bool) (any, error) {
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
		// or parameter names (e.g. db.Select("to")), not ISO codes.
		if isArgToSkipMethod(stack) {
			return true
		}

		// String literals assigned to configured fields (e.g. CardSchemes)
		// are domain values that happen to share the form of an ISO code.
		if isAssignToSkipField(stack, skipFields) {
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

// isAssignToSkipField reports whether the BasicLit at the top of the stack
// is part of a value assigned to a struct field whose name appears in
// skipFields.
//
// Two assignment shapes are recognized:
//
//	x.Field = ...                 // AssignStmt with single SelectorExpr LHS
//	StructType{Field: ...}        // KeyValueExpr inside a CompositeLit
//
// In both shapes the literal may be nested inside a CompositeLit RHS
// (e.g. pq.StringArray{"MC"}), so the stack is walked from the top down
// to find the nearest enclosing assignment. Tuple assignments
// (len(Lhs) > 1) are intentionally not handled — without type information
// the analyzer cannot correlate which RHS literal belongs to which LHS
// target, so it falls through and flags every literal as usual.
func isAssignToSkipField(stack []ast.Node, skipFields map[string]bool) bool {
	if len(skipFields) == 0 {
		return false
	}
	// Walk ancestors from nearest to farthest. We stop at function
	// boundaries because a literal beyond a FuncLit/FuncDecl is no
	// longer "inside" the original assignment.
	for i := len(stack) - 2; i >= 0; i-- {
		switch parent := stack[i].(type) {
		case *ast.KeyValueExpr:
			if name, ok := identName(parent.Key); ok && skipFields[name] {
				return true
			}
			// A KeyValueExpr binds the field; no need to keep walking.
			return false
		case *ast.AssignStmt:
			if len(parent.Lhs) != 1 {
				return false
			}
			if name, ok := fieldNameFromLHS(parent.Lhs[0]); ok && skipFields[name] {
				return true
			}
			return false
		case *ast.FuncLit, *ast.FuncDecl:
			return false
		}
	}
	return false
}

// identName returns the name of an *ast.Ident expression.
func identName(expr ast.Expr) (string, bool) {
	id, ok := expr.(*ast.Ident)
	if !ok {
		return "", false
	}
	return id.Name, true
}

// fieldNameFromLHS extracts a struct field name from an assignment LHS.
// Only *ast.SelectorExpr (e.g. a.CardSchemes) is matched — bare identifiers
// are local variables, not struct fields, and a literal assigned to a local
// var that happens to share a skip-field name should still be flagged.
func fieldNameFromLHS(expr ast.Expr) (string, bool) {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}
	return sel.Sel.Name, true
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
