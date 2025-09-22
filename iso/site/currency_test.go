package site

import (
	"github.com/wego/pkg/currency"
	"testing"
)

func TestCurrency(t *testing.T) {
	tests := []struct {
		name     string
		siteCode string
		want     string
		wantOk   bool
	}{
		// Major countries
		{
			name:     "United States",
			siteCode: US,
			want:     currency.USD,
			wantOk:   true,
		},
		{
			name:     "United Kingdom",
			siteCode: GB,
			want:     currency.GBP,
			wantOk:   true,
		},
		{
			name:     "Japan",
			siteCode: JP,
			want:     currency.JPY,
			wantOk:   true,
		},
		{
			name:     "China",
			siteCode: CN,
			want:     currency.CNY,
			wantOk:   true,
		},

		// Eurozone countries
		{
			name:     "Germany",
			siteCode: DE,
			want:     currency.EUR,
			wantOk:   true,
		},
		{
			name:     "France",
			siteCode: FR,
			want:     currency.EUR,
			wantOk:   true,
		},
		{
			name:     "Italy",
			siteCode: IT,
			want:     currency.EUR,
			wantOk:   true,
		},

		// Territories using parent country currency
		{
			name:     "Puerto Rico",
			siteCode: PR,
			want:     currency.USD,
			wantOk:   true,
		},
		{
			name:     "Guam",
			siteCode: GU,
			want:     currency.USD,
			wantOk:   true,
		},
		{
			name:     "British Virgin Islands",
			siteCode: VG,
			want:     currency.USD,
			wantOk:   true,
		},

		// Territories using regional currency
		{
			name:     "Aruba",
			siteCode: AW,
			want:     currency.AWG,
			wantOk:   true,
		},
		{
			name:     "Curaçao",
			siteCode: CW,
			want:     currency.ANG,
			wantOk:   true,
		},
		{
			name:     "Sint Maarten",
			siteCode: SX,
			want:     currency.ANG,
			wantOk:   true,
		},

		// Special cases
		{
			name:     "Bonaire, Sint Eustatius and Saba",
			siteCode: BQ,
			want:     currency.USD,
			wantOk:   true,
		},
		{
			name:     "Hong Kong",
			siteCode: HK,
			want:     currency.HKD,
			wantOk:   true,
		},
		{
			name:     "Macao",
			siteCode: MO,
			want:     currency.MOP,
			wantOk:   true,
		},

		// Invalid cases
		{
			name:     "Non-existent country",
			siteCode: "XX",
			want:     "",
			wantOk:   false,
		},
		{
			name:     "Empty string",
			siteCode: "",
			want:     "",
			wantOk:   false,
		},
		{
			name:     "Invalid length",
			siteCode: "XXX",
			want:     "",
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := Currency(tt.siteCode)
			if got != tt.want || gotOk != tt.wantOk {
				t.Errorf("Currency(%q) = (%q, %v), want (%q, %v)", tt.siteCode, got, gotOk, tt.want, tt.wantOk)
			}
		})
	}
}
