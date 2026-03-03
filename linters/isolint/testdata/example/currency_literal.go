package example

import "github.com/wego/pkg/currency"

func currencyChecks() {
	// Comparisons
	var code string
	if code == "USD" { // want `use currency\.USD instead of "USD"`
		println("match")
	}
	if "EUR" == code { // want `use currency\.EUR instead of "EUR"`
		println("match")
	}
	if code != "SGD" { // want `use currency\.SGD instead of "SGD"`
		println("no match")
	}

	// Assignments
	x := "JPY" // want `use currency\.JPY instead of "JPY"`
	_ = x

	// Var declarations
	var y = "GBP" // want `use currency\.GBP instead of "GBP"`
	_ = y

	// Switch/case
	switch code {
	case "THB": // want `use currency\.THB instead of "THB"`
		println("thai baht")
	case "VND": // want `use currency\.VND instead of "VND"`
		println("dong")
	}

	// Function arguments
	println("IDR") // want `use currency\.IDR instead of "IDR"`

	// Map keys
	m := map[string]int{
		"AUD": 1, // want `use currency\.AUD instead of "AUD"`
		"NZD": 2, // want `use currency\.NZD instead of "NZD"`
	}
	_ = m["CAD"] // want `use currency\.CAD instead of "CAD"`

	// Slice literals
	codes := []string{
		"MYR", // want `use currency\.MYR instead of "MYR"`
		"PHP", // want `use currency\.PHP instead of "PHP"`
	}
	_ = codes

	// Struct fields
	type config struct {
		Currency string
	}
	c := config{Currency: "INR"} // want `use currency\.INR instead of "INR"`
	_ = c

	// Return values (via helper)
	_ = getCurrency()
}

func getCurrency() string {
	return "HKD" // want `use currency\.HKD instead of "HKD"`
}

func lowercaseCurrencyChecks() {
	// Lowercase should also be caught
	x := "usd" // want `use currency\.USD instead of "usd"`
	_ = x

	y := "sgd" // want `use currency\.SGD instead of "sgd"`
	_ = y

	// Mixed case should NOT be caught
	z := "Usd"
	_ = z
}

// Ensure currency package is used (for compilation)
var _ = currency.USD
