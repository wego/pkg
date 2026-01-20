package stringlint

import "golang.org/x/tools/go/analysis"

// New is the entry point for the golangci-lint module plugin system.
// The signature must be: func New(any) ([]*analysis.Analyzer, error).
func New(conf any) ([]*analysis.Analyzer, error) {
	// conf contains settings from .golangci.yml if any.
	// Currently no configuration options are supported.
	return []*analysis.Analyzer{Analyzer}, nil
}
