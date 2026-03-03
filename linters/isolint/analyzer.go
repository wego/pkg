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
	"github.com/wego/pkg/currency":  true,
	"github.com/wego/pkg/iso/site":  true,
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
	if skipPackages[pass.Pkg.Path()] {
		return nil, nil
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.BasicLit)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		lit := n.(*ast.BasicLit)
		if lit.Kind != token.STRING {
			return
		}

		// Fast path: currency codes are 3 chars ("XXX" = 5 bytes quoted),
		// site codes are 2 chars ("XX" = 4 bytes quoted).
		// Skip everything else before allocating via Unquote.
		vlen := len(lit.Value)
		if vlen != 4 && vlen != 5 {
			return
		}

		value, err := strconv.Unquote(lit.Value)
		if err != nil {
			return
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
	})

	return nil, nil
}
