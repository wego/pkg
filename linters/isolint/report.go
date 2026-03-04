package isolint

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

const (
	currencyPkgPath = "github.com/wego/pkg/currency"
	sitePkgPath     = "github.com/wego/pkg/iso/site"
)

func reportCurrencyDiagnostic(pass *analysis.Pass, lit *ast.BasicLit, code string) {
	constName := CurrencyConstName(code)

	pass.Report(analysis.Diagnostic{
		Pos:     lit.Pos(),
		End:     lit.End(),
		Message: fmt.Sprintf("use %s instead of %q (import %q)", constName, code, currencyPkgPath),
		SuggestedFixes: []analysis.SuggestedFix{
			{
				Message: fmt.Sprintf("Replace with %s", constName),
				TextEdits: []analysis.TextEdit{
					{
						Pos:     lit.Pos(),
						End:     lit.End(),
						NewText: []byte(constName),
					},
				},
			},
		},
	})
}

func reportSiteDiagnostic(pass *analysis.Pass, lit *ast.BasicLit, code string) {
	constName := SiteConstName(code)

	pass.Report(analysis.Diagnostic{
		Pos:     lit.Pos(),
		End:     lit.End(),
		Message: fmt.Sprintf("use %s instead of %q (import %q)", constName, code, sitePkgPath),
		SuggestedFixes: []analysis.SuggestedFix{
			{
				Message: fmt.Sprintf("Replace with %s", constName),
				TextEdits: []analysis.TextEdit{
					{
						Pos:     lit.Pos(),
						End:     lit.End(),
						NewText: []byte(constName),
					},
				},
			},
		},
	})
}
