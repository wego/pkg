package stringlint

import (
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin(Analyzer.Name, New)
}

// New is the entry point for the golangci-lint module plugin system.
// The signature must be: func New(any) (register.LinterPlugin, error).
func New(_ any) (register.LinterPlugin, error) {
	// conf contains settings from .golangci.yml if any.
	// Currently no configuration options are supported.
	return plugin{}, nil
}

type plugin struct{}

func (plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{Analyzer}, nil
}

func (plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
