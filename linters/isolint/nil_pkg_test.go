package isolint

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// TestRunWithNilPkg verifies that the analyzer does not panic when pass.Pkg is
// nil. golangci-lint may leave pass.Pkg unset under LoadModeSyntax for certain
// packages, so the analyzer must tolerate it.
func TestRunWithNilPkg(t *testing.T) {
	const src = `package example

func f() {
	_ = "USD"
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "example.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	ins := inspector.New([]*ast.File{file})

	pass := &analysis.Pass{
		Analyzer: Analyzer,
		Fset:     fset,
		Files:    []*ast.File{file},
		Pkg:      nil, // simulate LoadModeSyntax with nil Pkg
		ResultOf: map[*analysis.Analyzer]any{inspect.Analyzer: ins},
		Report:   func(d analysis.Diagnostic) {},
	}

	// Must not panic.
	_, err = run(pass)
	if err != nil {
		t.Fatalf("run() returned unexpected error: %v", err)
	}
}
