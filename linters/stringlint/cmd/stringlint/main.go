// Command stringlint runs the stringlint analyzer.
package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/wego/pkg/linters/stringlint"
)

func main() {
	singlechecker.Main(stringlint.Analyzer)
}
