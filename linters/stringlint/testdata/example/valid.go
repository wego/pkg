package example

import wegostrings "github.com/wego/pkg/strings"

func validCode(s string, ptr *string) {
	// These should NOT trigger warnings

	// Already using the recommended functions
	if wegostrings.IsEmpty(s) {
		println("good")
	}
	if wegostrings.IsNotEmpty(s) {
		println("good")
	}
	if wegostrings.IsEmptyP(ptr) {
		println("good")
	}

	// Comparing to non-empty strings - this is intentional
	if s == "hello" {
		println("specific comparison is fine")
	}
	if s != "world" {
		println("specific comparison is fine")
	}

	// len() on slices - not our concern
	slice := []int{1, 2, 3}
	if len(slice) == 0 {
		println("slice is empty")
	}

	// len() on maps - not our concern
	m := map[string]int{}
	if len(m) == 0 {
		println("map is empty")
	}

	// Comparisons that aren't empty checks
	if len(s) == 5 {
		println("length is 5")
	}
	if len(s) > 10 {
		println("length > 10")
	}

	// String comparison between two variables
	other := "test"
	if s == other {
		println("comparing two strings")
	}
}

// Ensure wegostrings package is used
var _ = wegostrings.IsEmpty
