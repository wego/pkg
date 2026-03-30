package isolint_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/wego/pkg/linters/isolint"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, isolint.Analyzer, "./example")
}

func TestAnalyzerWithFixes(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.RunWithSuggestedFixes(t, testdata, isolint.Analyzer, "./example")
}
