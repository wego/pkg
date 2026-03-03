package example

import "github.com/wego/pkg/iso/site"

func siteChecks() {
	// Comparisons
	var code string
	if code == "SG" { // want `use site\.SG instead of "SG"`
		println("match")
	}
	if "US" == code { // want `use site\.US instead of "US"`
		println("match")
	}
	if code != "JP" { // want `use site\.JP instead of "JP"`
		println("no match")
	}

	// Assignments
	x := "GB" // want `use site\.GB instead of "GB"`
	_ = x

	// Switch/case
	switch code {
	case "TH": // want `use site\.TH instead of "TH"`
		println("thailand")
	case "VN": // want `use site\.VN instead of "VN"`
		println("vietnam")
	}

	// Function arguments
	println("ID") // want `use site\.ID instead of "ID"`

	// Map keys
	m := map[string]int{
		"AU": 1, // want `use site\.AU instead of "AU"`
		"NZ": 2, // want `use site\.NZ instead of "NZ"`
	}
	_ = m["CA"] // want `use site\.CA instead of "CA"`

	// Slice literals
	codes := []string{
		"MY", // want `use site\.MY instead of "MY"`
		"PH", // want `use site\.PH instead of "PH"`
	}
	_ = codes

	// Struct fields
	type config struct {
		Site string
	}
	c := config{Site: "IN"} // want `use site\.IN instead of "IN"`
	_ = c

	// Return values
	_ = getSite()
}

func getSite() string {
	return "HK" // want `use site\.HK instead of "HK"`
}

func lowercaseSiteChecks() {
	// Lowercase should also be caught
	x := "sg" // want `use site\.SG instead of "sg"`
	_ = x

	y := "us" // want `use site\.US instead of "us"`
	_ = y

	// Mixed case should NOT be caught
	z := "Sg"
	_ = z
}

// Ensure site package is used (for compilation)
var _ = site.SG
