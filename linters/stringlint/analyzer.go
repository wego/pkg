// Package stringlint provides a Go analyzer that detects direct string
// comparison patterns and recommends using github.com/wego/pkg/strings
// utility functions instead.
package stringlint

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer is the stringlint analyzer that checks for direct string comparisons.
var Analyzer = &analysis.Analyzer{
	Name:     "stringlint",
	Doc:      "recommends using github.com/wego/pkg/strings functions over direct string comparisons",
	URL:      "https://github.com/wego/pkg/linters/stringlint",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (any, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.BinaryExpr)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		binExpr := n.(*ast.BinaryExpr)

		// Only check comparison operators.
		switch binExpr.Op {
		case token.EQL, token.NEQ, token.GTR, token.GEQ, token.LSS, token.LEQ:
			// continue
		default:
			return
		}

		// Check for empty string comparison: s == "" or s != "".
		if result := checkEmptyStringComparison(pass, binExpr); result != nil {
			reportDiagnostic(pass, binExpr, result)
			return
		}

		// Check for len comparison: len(s) == 0 or len(s) > 0.
		if result := checkLenComparison(pass, binExpr); result != nil {
			reportDiagnostic(pass, binExpr, result)
			return
		}
	})

	return nil, nil
}
