package site

import "testing"

func TestCurrency(t *testing.T) {
	tests := []struct {
		name     string
		siteCode string
		want     string
	}{
		{
			name:     "United States",
			siteCode: "US",
			want:     "USD",
		},
		{
			name:     "United Kingdom",
			siteCode: "GB",
			want:     "GBP",
		},
		{
			name:     "European Union",
			siteCode: "EU",
			want:     "EUR",
		},
		{
			name:     "Japan",
			siteCode: "JP",
			want:     "JPY",
		},
		{
			name:     "China",
			siteCode: "CN",
			want:     "CNY",
		},
		{
			name:     "Case insensitive",
			siteCode: "us",
			want:     "USD",
		},
		{
			name:     "With whitespace",
			siteCode: " US ",
			want:     "USD",
		},
		{
			name:     "Non-existent country",
			siteCode: "XX",
			want:     "",
		},
		{
			name:     "Empty string",
			siteCode: "",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Currency(tt.siteCode)
			if got != tt.want {
				t.Errorf("Currency(%q) = %q, want %q", tt.siteCode, got, tt.want)
			}
		})
	}
}
