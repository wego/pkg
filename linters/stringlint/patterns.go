package stringlint

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// checkResult holds the result of a pattern check.
type checkResult struct {
	expr        ast.Expr // The expression to replace (e.g., the variable or *ptr).
	isPointer   bool     // Whether expr is a pointer dereference.
	isEmptyTest bool     // true for == "", false for != "".
}

// checkEmptyStringComparison checks for patterns like:
//   - s == ""  -> IsEmpty(s)
//   - s != ""  -> IsNotEmpty(s)
//   - *ptr == "" -> IsEmptyP(ptr)
//   - *ptr != "" -> IsNotEmptyP(ptr)
func checkEmptyStringComparison(pass *analysis.Pass, binExpr *ast.BinaryExpr) *checkResult {
	switch binExpr.Op {
	case token.EQL, token.NEQ:
		// continue
	default:
		return nil
	}

	var strExpr ast.Expr
	var emptyLit *ast.BasicLit

	// Check both orderings: s == "" or "" == s.
	if lit, ok := binExpr.Y.(*ast.BasicLit); ok && isEmptyStringLit(lit) {
		strExpr = binExpr.X
		emptyLit = lit
	} else if lit, ok := binExpr.X.(*ast.BasicLit); ok && isEmptyStringLit(lit) {
		strExpr = binExpr.Y
		emptyLit = lit
	}

	if emptyLit == nil {
		return nil
	}

	// Check if it's a pointer dereference.
	isPointer := false
	exprToReport := strExpr

	if starExpr, ok := strExpr.(*ast.StarExpr); ok {
		// *ptr == ""
		innerType := pass.TypesInfo.TypeOf(starExpr.X)
		if innerType == nil {
			return nil
		}
		if ptr, ok := innerType.Underlying().(*types.Pointer); ok {
			if basic, ok := ptr.Elem().Underlying().(*types.Basic); ok && basic.Kind() == types.String {
				isPointer = true
				exprToReport = starExpr.X // Report on ptr, not *ptr.
			}
		}
	} else {
		// s == ""
		t := pass.TypesInfo.TypeOf(strExpr)
		if t == nil {
			return nil
		}
		basic, ok := t.Underlying().(*types.Basic)
		if !ok || basic.Kind() != types.String {
			return nil
		}
	}

	return &checkResult{
		expr:        exprToReport,
		isPointer:   isPointer,
		isEmptyTest: binExpr.Op == token.EQL,
	}
}

// checkLenComparison checks for patterns like:
//   - len(s) == 0 -> IsEmpty(s)
//   - len(s) != 0 -> IsNotEmpty(s)
//   - len(s) > 0  -> IsNotEmpty(s)
//   - len(s) < 1  -> IsEmpty(s)
//   - 0 == len(s) -> IsEmpty(s) (reversed)
func checkLenComparison(pass *analysis.Pass, binExpr *ast.BinaryExpr) *checkResult {
	var lenCall *ast.CallExpr
	var numLit *ast.BasicLit
	var reversed bool

	// Check both orderings: len(s) == 0 or 0 == len(s).
	if call, ok := binExpr.X.(*ast.CallExpr); ok {
		if isLenCall(call) {
			lenCall = call
			if lit, ok := binExpr.Y.(*ast.BasicLit); ok {
				numLit = lit
			}
		}
	}
	if lenCall == nil {
		if call, ok := binExpr.Y.(*ast.CallExpr); ok {
			if isLenCall(call) {
				lenCall = call
				reversed = true
				if lit, ok := binExpr.X.(*ast.BasicLit); ok {
					numLit = lit
				}
			}
		}
	}

	if lenCall == nil || numLit == nil || len(lenCall.Args) != 1 {
		return nil
	}

	// Check if comparing to 0 or 1.
	if numLit.Kind != token.INT {
		return nil
	}

	arg := lenCall.Args[0]
	t := pass.TypesInfo.TypeOf(arg)
	if t == nil {
		return nil
	}

	// Only handle string types (not slices, maps, etc.).
	basic, ok := t.Underlying().(*types.Basic)
	if !ok || basic.Kind() != types.String {
		return nil
	}

	// Determine if this is an empty test or not-empty test.
	isEmptyTest := false
	op := binExpr.Op

	// Handle reversed comparisons (0 == len(s) vs len(s) == 0).
	if reversed {
		switch op {
		case token.LSS: // 0 < len(s) means not empty.
			op = token.GTR
		case token.GTR: // 0 > len(s) means empty (always false, but handle it).
			op = token.LSS
		case token.LEQ: // 0 <= len(s) always true for len.
			return nil
		case token.GEQ: // 0 >= len(s) means empty.
			op = token.LEQ
		}
	}

	switch {
	case op == token.EQL && numLit.Value == "0":
		// len(s) == 0
		isEmptyTest = true
	case op == token.NEQ && numLit.Value == "0":
		// len(s) != 0
		isEmptyTest = false
	case op == token.GTR && numLit.Value == "0":
		// len(s) > 0
		isEmptyTest = false
	case op == token.GEQ && numLit.Value == "1":
		// len(s) >= 1
		isEmptyTest = false
	case op == token.LSS && numLit.Value == "1":
		// len(s) < 1
		isEmptyTest = true
	case op == token.LEQ && numLit.Value == "0":
		// len(s) <= 0
		isEmptyTest = true
	default:
		return nil
	}

	return &checkResult{
		expr:        arg,
		isPointer:   false, // len() doesn't work with pointers.
		isEmptyTest: isEmptyTest,
	}
}

// isEmptyStringLit checks if the literal is an empty string "".
func isEmptyStringLit(lit *ast.BasicLit) bool {
	return lit.Kind == token.STRING && (lit.Value == `""` || lit.Value == "``")
}

// isLenCall checks if the call expression is a call to the builtin len().
func isLenCall(call *ast.CallExpr) bool {
	ident, ok := call.Fun.(*ast.Ident)
	return ok && ident.Name == "len"
}

// pkgAlias is the recommended import alias for github.com/wego/pkg/strings
// to avoid conflict with the stdlib "strings" package.
const pkgAlias = "wegostrings"

// reportDiagnostic reports the diagnostic with a suggested fix.
func reportDiagnostic(pass *analysis.Pass, binExpr *ast.BinaryExpr, result *checkResult) {
	funcName := getSuggestedFunc(result)
	exprStr := render(pass.Fset, result.expr)

	// Build the replacement: wegostrings.IsEmpty(s) or wegostrings.IsEmptyP(ptr).
	replacement := fmt.Sprintf("%s.%s(%s)", pkgAlias, funcName, exprStr)

	pass.Report(analysis.Diagnostic{
		Pos:     binExpr.Pos(),
		End:     binExpr.End(),
		Message: fmt.Sprintf("use %s.%s(%s) instead of direct comparison (import %s \"github.com/wego/pkg/strings\")", pkgAlias, funcName, exprStr, pkgAlias),
		SuggestedFixes: []analysis.SuggestedFix{
			{
				Message: fmt.Sprintf("Replace with %s.%s(%s)", pkgAlias, funcName, exprStr),
				TextEdits: []analysis.TextEdit{
					{
						Pos:     binExpr.Pos(),
						End:     binExpr.End(),
						NewText: []byte(replacement),
					},
				},
			},
		},
	})
}

// getSuggestedFunc returns the appropriate function name based on the check result.
func getSuggestedFunc(result *checkResult) string {
	if result.isPointer {
		if result.isEmptyTest {
			return "IsEmptyP"
		}
		return "IsNotEmptyP"
	}
	if result.isEmptyTest {
		return "IsEmpty"
	}
	return "IsNotEmpty"
}

// render renders an AST node to a string.
func render(fset *token.FileSet, node any) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, node); err != nil {
		return ""
	}
	return buf.String()
}
