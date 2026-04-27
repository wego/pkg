package isolint

import (
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin(Analyzer.Name, New)
}

// New is the entry point for the golangci-lint module plugin system.
// The signature must be: func New(any) (register.LinterPlugin, error).
//
// Settings are decoded from the .golangci.yml block:
//
//	linters:
//	  settings:
//	    custom:
//	      isolint:
//	        type: module
//	        settings:
//	          skip-fields: [CardSchemes]
func New(settings any) (register.LinterPlugin, error) {
	s, err := register.DecodeSettings[Settings](settings)
	if err != nil {
		return nil, err
	}
	return plugin{settings: s}, nil
}

type plugin struct {
	settings Settings
}

func (p plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{NewAnalyzer(p.settings)}, nil
}

func (plugin) GetLoadMode() string {
	return register.LoadModeSyntax
}
