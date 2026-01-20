package example

import wegostrings "github.com/wego/pkg/strings"

func lenChecks(s string) {
	// These should trigger warnings
	if len(s) == 0 { // want `use wegostrings.IsEmpty\(s\) instead of direct comparison`
		println("len empty")
	}

	if len(s) != 0 { // want `use wegostrings.IsNotEmpty\(s\) instead of direct comparison`
		println("len not empty")
	}

	if len(s) > 0 { // want `use wegostrings.IsNotEmpty\(s\) instead of direct comparison`
		println("len greater")
	}

	if len(s) >= 1 { // want `use wegostrings.IsNotEmpty\(s\) instead of direct comparison`
		println("len gte 1")
	}

	if len(s) < 1 { // want `use wegostrings.IsEmpty\(s\) instead of direct comparison`
		println("len lt 1")
	}

	if 0 == len(s) { // want `use wegostrings.IsEmpty\(s\) instead of direct comparison`
		println("reversed")
	}
}

// Ensure wegostrings package is used (for golden file)
var _ = wegostrings.IsEmpty
