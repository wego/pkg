package country

import "testing"

func TestFromNumeric(t *testing.T) {
	tests := []struct {
		name   string
		given  string
		want   string
		wantOk bool
	}{
		{name: "unknown numeric", given: "000", want: "", wantOk: false},
		{name: "empty", given: "", want: "", wantOk: false},
		{name: "whitespace only", given: "   ", want: "", wantOk: false},
		{name: "non-numeric input", given: "12a", want: "", wantOk: false},
		{name: "too long", given: "1234", want: "", wantOk: false},

		{name: "single-digit Brazil", given: "76", want: BR, wantOk: true},
		{name: "padded Brazil", given: "076", want: BR, wantOk: true},
		{name: "padded surrounded by whitespace", given: " 076 ", want: BR, wantOk: true},

		{name: "India", given: "356", want: IN, wantOk: true},
		{name: "United States", given: "840", want: US, wantOk: true},
		{name: "United Kingdom", given: "826", want: GB, wantOk: true},
		{name: "UAE", given: "784", want: AE, wantOk: true},
		{name: "Saudi Arabia", given: "682", want: SA, wantOk: true},
		{name: "Egypt", given: "818", want: EG, wantOk: true},
		{name: "Singapore", given: "702", want: SG, wantOk: true},
		{name: "Indonesia", given: "360", want: ID, wantOk: true},
		{name: "China", given: "156", want: CN, wantOk: true},
		{name: "Japan", given: "392", want: JP, wantOk: true},
		{name: "Germany", given: "276", want: DE, wantOk: true},
		{name: "France", given: "250", want: FR, wantOk: true},
		{name: "Hong Kong", given: "344", want: HK, wantOk: true},

		{name: "Austria (40 → 040)", given: "40", want: AT, wantOk: true},
		{name: "Australia (36 → 036)", given: "36", want: AU, wantOk: true},
		{name: "Argentina (32 → 032)", given: "32", want: AR, wantOk: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := FromNumeric(tt.given)
			if got != tt.want || ok != tt.wantOk {
				t.Errorf("FromNumeric(%q) = (%q, %v), want (%q, %v)",
					tt.given, got, ok, tt.want, tt.wantOk)
			}
		})
	}
}

func TestNumeric(t *testing.T) {
	tests := []struct {
		name   string
		given  string
		want   string
		wantOk bool
	}{
		{name: "unknown alpha-2", given: "ZZ", want: "", wantOk: false},
		{name: "empty", given: "", want: "", wantOk: false},

		{name: "United States", given: US, want: "840", wantOk: true},
		{name: "Brazil zero-pads", given: BR, want: "076", wantOk: true},
		{name: "India", given: IN, want: "356", wantOk: true},
		{name: "lowercase normalised", given: "us", want: "840", wantOk: true},
		{name: "whitespace stripped", given: "  GB ", want: "826", wantOk: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Numeric(tt.given)
			if got != tt.want || ok != tt.wantOk {
				t.Errorf("Numeric(%q) = (%q, %v), want (%q, %v)",
					tt.given, got, ok, tt.want, tt.wantOk)
			}
		})
	}
}

func TestNumericRoundTrip(t *testing.T) {
	for numeric, a2 := range numericToAlpha2 {
		got, ok := FromNumeric(numeric)
		if !ok || got != a2 {
			t.Errorf("FromNumeric(%q) = (%q, %v); want (%q, true)",
				numeric, got, ok, a2)
		}
		back, ok := Numeric(a2)
		if !ok || back != numeric {
			t.Errorf("Numeric(%q) = (%q, %v); want (%q, true)",
				a2, back, ok, numeric)
		}
	}
}
