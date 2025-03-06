package site

import "testing"

func TestCurrency(t *testing.T) {
	tests := []struct {
		name     string
		siteCode Site
		want     string
		wantOk   bool
	}{
		// Major countries
		{
			name:     "United States",
			siteCode: Site(US),
			want:     "USD",
			wantOk:   true,
		},
		{
			name:     "United Kingdom",
			siteCode: Site(GB),
			want:     "GBP",
			wantOk:   true,
		},
		{
			name:     "Japan",
			siteCode: Site(JP),
			want:     "JPY",
			wantOk:   true,
		},
		{
			name:     "China",
			siteCode: Site(CN),
			want:     "CNY",
			wantOk:   true,
		},

		// Eurozone countries
		{
			name:     "Germany",
			siteCode: Site(DE),
			want:     "EUR",
			wantOk:   true,
		},
		{
			name:     "France",
			siteCode: Site(FR),
			want:     "EUR",
			wantOk:   true,
		},
		{
			name:     "Italy",
			siteCode: Site(IT),
			want:     "EUR",
			wantOk:   true,
		},

		// Territories using parent country currency
		{
			name:     "Puerto Rico",
			siteCode: Site(PR),
			want:     "USD",
			wantOk:   true,
		},
		{
			name:     "Guam",
			siteCode: Site(GU),
			want:     "USD",
			wantOk:   true,
		},
		{
			name:     "British Virgin Islands",
			siteCode: Site(VG),
			want:     "USD",
			wantOk:   true,
		},

		// Territories using regional currency
		{
			name:     "Aruba",
			siteCode: Site(AW),
			want:     "AWG",
			wantOk:   true,
		},
		{
			name:     "Curaçao",
			siteCode: Site(CW),
			want:     "ANG",
			wantOk:   true,
		},
		{
			name:     "Sint Maarten",
			siteCode: Site(SX),
			want:     "ANG",
			wantOk:   true,
		},

		// Special cases
		{
			name:     "Bonaire, Sint Eustatius and Saba",
			siteCode: Site(BQ),
			want:     "USD",
			wantOk:   true,
		},
		{
			name:     "Hong Kong",
			siteCode: Site(HK),
			want:     "HKD",
			wantOk:   true,
		},
		{
			name:     "Macao",
			siteCode: Site(MO),
			want:     "MOP",
			wantOk:   true,
		},

		// Invalid cases
		{
			name:     "Non-existent country",
			siteCode: Site("XX"),
			want:     "",
			wantOk:   false,
		},
		{
			name:     "Empty string",
			siteCode: Site(""),
			want:     "",
			wantOk:   false,
		},
		{
			name:     "Invalid length",
			siteCode: Site("XXX"),
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
