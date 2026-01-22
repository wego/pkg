package stringlint_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/wego/pkg/linters/stringlint"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, stringlint.Analyzer, "./example")
}

func TestAnalyzerWithFixes(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.RunWithSuggestedFixes(t, testdata, stringlint.Analyzer, "./example")
}
