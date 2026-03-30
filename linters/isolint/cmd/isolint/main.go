// Command isolint runs the isolint analyzer.
package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/wego/pkg/linters/isolint"
)

func main() {
	singlechecker.Main(isolint.Analyzer)
}
