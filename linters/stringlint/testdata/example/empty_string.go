package example

import wegostrings "github.com/wego/pkg/strings"

func emptyStringChecks(s string) {
	// These should trigger warnings
	if s == "" { // want `use wegostrings.IsEmpty\(s\) instead of direct comparison`
		println("empty")
	}

	if s != "" { // want `use wegostrings.IsNotEmpty\(s\) instead of direct comparison`
		println("not empty")
	}

	if "" == s { // want `use wegostrings.IsEmpty\(s\) instead of direct comparison`
		println("empty reversed")
	}

	if "" != s { // want `use wegostrings.IsNotEmpty\(s\) instead of direct comparison`
		println("not empty reversed")
	}

	// Backtick empty string
	if s == `` { // want `use wegostrings.IsEmpty\(s\) instead of direct comparison`
		println("backtick empty")
	}
}

// Ensure wegostrings package is used (for golden file)
var _ = wegostrings.IsEmpty
