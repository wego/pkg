package isolint

import "testing"

func TestIsCurrencyCode(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		// Uppercase (canonical)
		{"USD", true},
		{"EUR", true},
		{"GBP", true},
		{"JPY", true},
		{"SGD", true},
		{"AUD", true},
		{"CNY", true},
		{"INR", true},
		{"THB", true},
		{"VND", true},

		// Lowercase (also caught)
		{"usd", true},
		{"eur", true},
		{"sgd", true},
		{"jpy", true},

		// Edge currency codes
		{"XAU", true},  // Gold
		{"XAG", true},  // Silver
		{"XXX", true},  // No currency
		{"ZWG", true},  // Newest (2024)
		{"BOV", true},  // Rare fund code
		{"PRB", false}, // Transnistrian Ruble — constant exists but not in ISO 4217 validation
		{"CHE", true},  // WIR Euro
		{"USN", true},  // US Dollar next day

		// NOT currency codes
		{"", false},
		{"US", false},       // Too short (site code)
		{"USDD", false},     // Too long
		{"Usd", false},      // Mixed case — not matched
		{"uSd", false},      // Mixed case — not matched
		{"hello", false},    // Random string
		{"123", false},      // Numbers
		{"SG", false},       // Site code, not currency
		{"XYZ", false},      // Not a real code
		{"ABC", false},      // Not a real code
		{"FOO", false},      // Not a real code
		{"foo", false},      // Not a real code (lowercase)
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := IsCurrencyCode(tt.code)
			if got != tt.want {
				t.Errorf("IsCurrencyCode(%q) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestIsSiteCode(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		// Uppercase (canonical)
		{"SG", true},
		{"US", true},
		{"GB", true},
		{"JP", true},
		{"AU", true},
		{"DE", true},
		{"FR", true},
		{"IN", true},
		{"TH", true},
		{"VN", true},

		// Lowercase (also caught)
		{"sg", true},
		{"us", true},
		{"jp", true},
		{"gb", true},

		// Edge site codes — site.Currency() only covers sites with a currency mapping
		{"AQ", false}, // Antarctica — no currency mapping
		{"BV", true},  // Bouvet Island — maps to NOK
		{"HM", true},  // Heard Island — maps to AUD
		{"UM", true},  // US Minor Outlying Islands — maps to USD
		{"AX", false}, // Aland Islands — no currency mapping
		{"BQ", true},  // Bonaire — maps to USD

		// NOT site codes
		{"", false},
		{"USD", false},      // Too long (currency code)
		{"S", false},        // Too short
		{"SGD", false},      // Currency code, not site
		{"Sg", false},       // Mixed case — not matched
		{"hello", false},    // Random string
		{"12", false},       // Numbers
		{"XX", false},       // Not a real code
		{"ZZ", false},       // Not a real code
		{"QQ", false},       // Not a real code
		{"xx", false},       // Not a real code (lowercase)
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := IsSiteCode(tt.code)
			if got != tt.want {
				t.Errorf("IsSiteCode(%q) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestCurrencyConstName(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{"USD", "currency.USD"},
		{"EUR", "currency.EUR"},
		{"SGD", "currency.SGD"},
		{"JPY", "currency.JPY"},
		{"usd", "currency.USD"}, // Lowercase input normalizes to uppercase constant
		{"sgd", "currency.SGD"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := CurrencyConstName(tt.code)
			if got != tt.want {
				t.Errorf("CurrencyConstName(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestSiteConstName(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{"SG", "site.SG"},
		{"US", "site.US"},
		{"JP", "site.JP"},
		{"GB", "site.GB"},
		{"sg", "site.SG"}, // Lowercase input normalizes to uppercase constant
		{"us", "site.US"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := SiteConstName(tt.code)
			if got != tt.want {
				t.Errorf("SiteConstName(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestNoOverlapBetweenCurrencyAndSite(t *testing.T) {
	// Currency codes are 3 letters, site codes are 2 letters — no overlap possible.
	currencySamples := []string{"USD", "EUR", "SGD", "JPY", "GBP", "AUD", "THB", "VND"}
	siteSamples := []string{"SG", "US", "JP", "GB", "AU", "TH", "VN", "DE"}

	for _, code := range currencySamples {
		if IsSiteCode(code) {
			t.Errorf("currency code %q is also a site code — unexpected overlap", code)
		}
	}

	for _, code := range siteSamples {
		if IsCurrencyCode(code) {
			t.Errorf("site code %q is also a currency code — unexpected overlap", code)
		}
	}
}

// Benchmarks — measure the cost of checking a string literal in the hot path.

func BenchmarkIsCurrencyCode_Match(b *testing.B) {
	for b.Loop() {
		IsCurrencyCode("USD")
	}
}

func BenchmarkIsCurrencyCode_LowercaseMatch(b *testing.B) {
	for b.Loop() {
		IsCurrencyCode("usd")
	}
}

func BenchmarkIsCurrencyCode_NoMatch_WrongLength(b *testing.B) {
	for b.Loop() {
		IsCurrencyCode("error: something went wrong")
	}
}

func BenchmarkIsCurrencyCode_NoMatch_RightLength(b *testing.B) {
	for b.Loop() {
		IsCurrencyCode("foo")
	}
}

func BenchmarkIsSiteCode_Match(b *testing.B) {
	for b.Loop() {
		IsSiteCode("SG")
	}
}

func BenchmarkIsSiteCode_LowercaseMatch(b *testing.B) {
	for b.Loop() {
		IsSiteCode("sg")
	}
}

func BenchmarkIsSiteCode_NoMatch_WrongLength(b *testing.B) {
	for b.Loop() {
		IsSiteCode("this is a long string")
	}
}
