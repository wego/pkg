package currency_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/currency"
)

func Test_IsISO4217(t *testing.T) {
	assertions := assert.New(t)
	assertions.True(currency.IsISO4217("afn"))
	assertions.True(currency.IsISO4217("UsD"))
	assertions.True(currency.IsISO4217("UsD "))
	assertions.True(currency.IsISO4217("CNY"))
	assertions.True(currency.IsISO4217("KWD"))
	assertions.False(currency.IsISO4217("UsDD"))
	assertions.False(currency.IsISO4217(""))
}

func Test_ToMinorUnit_OK(t *testing.T) {
	assertions := assert.New(t)

	rs, err := currency.ToMinorUnit("SGD", 25.12)
	assertions.NoError(err)
	assertions.EqualValues(2512, rs)
	assertions.EqualValues(rs, currency.MinorUnitAmount("SGD", 25.12))

	rs, err = currency.ToMinorUnit("SGD", 25.123)
	assertions.NoError(err)
	assertions.EqualValues(2512, rs)
	assertions.EqualValues(rs, currency.MinorUnitAmount("SGD", 25.123))

	rs, err = currency.ToMinorUnit("BHD", 25.123)
	assertions.NoError(err)
	assertions.EqualValues(25123, rs)
	assertions.EqualValues(rs, currency.MinorUnitAmount("BHD", 25.123))

	rs, err = currency.ToMinorUnit("BHD", 25.1234)
	assertions.NoError(err)
	assertions.EqualValues(25123, rs)
	assertions.EqualValues(rs, currency.MinorUnitAmount("BHD", 25.1234))

	rs, err = currency.ToMinorUnit("VND", 25.12345)
	assertions.NoError(err)
	assertions.EqualValues(25, rs)
	assertions.EqualValues(rs, currency.MinorUnitAmount("VND", 25.12345))

	rs, err = currency.ToMinorUnit("Vnd", 25.12345)
	assertions.NoError(err)
	assertions.EqualValues(25, rs)
	assertions.EqualValues(rs, currency.MinorUnitAmount("Vnd", 25.12345))
}

func Test_ToMinorUnit_InvalidCurrency(t *testing.T) {
	assertions := assert.New(t)

	rs, err := currency.ToMinorUnit("", 25.12)
	assertions.Error(err)
	assertions.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assertions.Zero(rs)

	rs, err = currency.ToMinorUnit("sg", 25.12)
	assertions.Error(err)
	assertions.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assertions.Zero(rs)

	rs, err = currency.ToMinorUnit(" ", 25.12)
	assertions.Error(err)
	assertions.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assertions.Zero(rs)

	rs, err = currency.ToMinorUnit("      ", 25.12)
	assertions.Error(err)
	assertions.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assertions.Zero(rs)
}

func Test_ToMinorUnit_ZeroAmount(t *testing.T) {
	assertions := assert.New(t)

	rs, err := currency.ToMinorUnit("SGD", 0)
	assertions.NoError(err)
	assertions.Zero(rs)
}

func Test_ToMinorUnit_InvalidAmount(t *testing.T) {
	assertions := assert.New(t)

	rs, err := currency.ToMinorUnit("USD", -0.0001)
	assertions.Error(err)
	assertions.Contains(err.Error(), "invalid amount")
	assertions.Zero(rs)

	rs, err = currency.ToMinorUnit("SGd", -0.0001)
	assertions.Error(err)
	assertions.Contains(err.Error(), "invalid amount")
	assertions.Zero(rs)
}

func Test_FromMinorUnit_OK(t *testing.T) {
	assertions := assert.New(t)

	rs, err := currency.FromMinorUnit("SGD", 2599)
	assertions.NoError(err)
	assertions.EqualValues(25.99, rs)

	rs, err = currency.FromMinorUnit("SGD", 25123)
	assertions.NoError(err)
	assertions.EqualValues(251.23, rs)

	rs, err = currency.FromMinorUnit("BHD", 25123)
	assertions.NoError(err)
	assertions.EqualValues(25.123, rs)

	rs, err = currency.FromMinorUnit("BHD", 251987)
	assertions.NoError(err)
	assertions.EqualValues(251.987, rs)

	rs, err = currency.FromMinorUnit("VND", 2512345)
	assertions.NoError(err)
	assertions.EqualValues(2512345, rs)

	rs, err = currency.FromMinorUnit("Vnd", 2512345)
	assertions.NoError(err)
	assertions.EqualValues(2512345, rs)
}

func Test_FromMinorUnit_InvalidCurrency(t *testing.T) {
	assertions := assert.New(t)

	rs, err := currency.FromMinorUnit("", 2512)
	assertions.Error(err)
	assertions.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assertions.Zero(rs)

	rs, err = currency.FromMinorUnit("sg", 2512)
	assertions.Error(err)
	assertions.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assertions.Zero(rs)

	rs, err = currency.FromMinorUnit(" ", 2512)
	assertions.Error(err)
	assertions.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assertions.Zero(rs)

	rs, err = currency.FromMinorUnit("      ", 2512)
	assertions.Error(err)
	assertions.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assertions.Zero(rs)
}

func Test_FromMinorUnit_ZeroAmount(t *testing.T) {
	assertions := assert.New(t)

	rs, err := currency.FromMinorUnit("USD", 0)
	assertions.NoError(err)
	assertions.Zero(rs)

	rs, err = currency.FromMinorUnit("VND", 0)
	assertions.NoError(err)
	assertions.Zero(rs)

	rs, err = currency.FromMinorUnit("KWD", 0)
	assertions.NoError(err)
	assertions.Zero(rs)
}

func Test_Format(t *testing.T) {
	for n, tc := range map[string]struct {
		amount       float64
		currencyCode string
		locale       string
		want         string
		wantErr      bool
	}{
		"Invalid currency": {
			amount:       2048.172,
			currencyCode: "SG",
			locale:       "en",
			want:         "",
			wantErr:      true,
		},
		"Empty locale should be fallback to en locale": {
			amount:       2048.179,
			currencyCode: "USD",
			want:         "$2,048.18",
		},
		"Invalid locale should be fallback en locale": {
			amount:       2048.179,
			currencyCode: "USD",
			locale:       "qwertyuiop",
			want:         "$2,048.18",
		},
		"USD en_US": {
			amount:       2048.123,
			currencyCode: "USD",
			locale:       "en_US",
			want:         "$2,048.12",
		},
		"EUR en": {
			amount:       2048.125,
			currencyCode: "EUR",
			locale:       "en",
			want:         "€2,048.13",
		},
		"JPY jp": {
			amount:       2048.445,
			currencyCode: "JPY",
			locale:       "en",
			want:         "¥2,048",
		},
		"VND vi": {
			amount:       2048.584,
			currencyCode: "VND",
			locale:       "vi",
			want:         "2.049\u00a0₫",
		},
		"KWD en": {
			amount:       2048.2023,
			currencyCode: "KWD",
			locale:       "en",
			want:         "KWD\u00a02,048.202",
		},
		"BHD en": {
			amount:       2048.4555,
			currencyCode: "BHD",
			locale:       "en",
			want:         "BHD\u00a02,048.456",
		},
	} {
		t.Run(n, func(tt *testing.T) {
			got, err := currency.Format(tc.amount, tc.currencyCode, tc.locale)
			if tc.wantErr {
				assert.Error(tt, err)
				assert.Empty(tt, got)
			} else {
				assert.NoError(tt, err)
				assert.Equal(tt, tc.want, got)
			}
		})
	}
}

func Test_FormatAmount(t *testing.T) {
	for n, tc := range map[string]struct {
		currencyCode string
		amount       float64
		want         string
	}{
		"Invalid currency": {
			currencyCode: "SG",
			amount:       2048.172,
			want:         "2048.17",
		},
		"SGD": {
			currencyCode: "SGD",
			amount:       252.2048,
			want:         "252.20",
		},
		"USD round up": {
			currencyCode: "USD",
			amount:       2048.1758,
			want:         "2048.18",
		},
		"BHD": {
			currencyCode: "BHD",
			amount:       2048.256432,
			want:         "2048.256",
		},
		"KWD round up": {
			currencyCode: "KWD",
			amount:       267.251678,
			want:         "267.252",
		},
		"JPY": {
			currencyCode: "JPY",
			amount:       272500.423,
			want:         "272500",
		},
		"VND round up": {
			currencyCode: "VND",
			amount:       262500000.523,
			want:         "262500001",
		},
	} {
		t.Run(n, func(tt *testing.T) {
			got := currency.FormatAmount(tc.amount, tc.currencyCode)
			assert.Equal(tt, tc.want, got)
		})
	}
}

func Test_GetCurrencyFactor(t *testing.T) {
	assertions := assert.New(t)

	factorOf100 := []string{"AED", "USD", "PHP", "qwerty", "1234", "SAR"}
	factorOf1000 := []string{"BHD", "IQD", "JOD", "KWD", "LYD", "OMR", "TND"}
	factorOfOne := []string{"BIF", "CLF", "DJF", "GNF", "ISK", "JPY", "KMF",
		"KRW", "PYG", "RWF", "UGX", "VUV", "VND", "XAF", "XOF", "XPF"}

	for _, cur := range factorOf100 {
		factor := currency.GetCurrencyFactor(cur)
		assertions.Equal(100.0, factor)
	}

	for _, cur := range factorOf1000 {
		factor := currency.GetCurrencyFactor(cur)
		assertions.Equal(1000.0, factor)
	}

	for _, cur := range factorOfOne {
		factor := currency.GetCurrencyFactor(cur)
		assertions.Equal(1.0, factor)
	}
}

func Test_Round_InvalidCurrency(t *testing.T) {
	assertions := assert.New(t)

	rs, err := currency.Round("", 2512)
	assertions.Error(err)
	assertions.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assertions.Zero(rs)

	rs, err = currency.Round("sg", 2512)
	assertions.Error(err)
	assertions.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assertions.Zero(rs)

	rs, err = currency.Round(" ", 2512)
	assertions.Error(err)
	assertions.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assertions.Zero(rs)

	rs, err = currency.Round("      ", 2512)
	assertions.Error(err)
	assertions.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assertions.Zero(rs)
}

func Test_Round_InvalidAmount(t *testing.T) {
	assertions := assert.New(t)

	rs, err := currency.Round("USD", -0.0001)
	assertions.Error(err)
	assertions.Contains(err.Error(), "invalid amount")
	assertions.Zero(rs)

	rs, err = currency.Round("SGd", -0.0001)
	assertions.Error(err)
	assertions.Contains(err.Error(), "invalid amount")
	assertions.Zero(rs)
}

func Test_Round_ZeroAmount(t *testing.T) {
	assertions := assert.New(t)

	rs, err := currency.Round("USD", 0)
	assertions.NoError(err)
	assertions.Zero(rs)

	rs, err = currency.Round("VND", 0)
	assertions.NoError(err)
	assertions.Zero(rs)

	rs, err = currency.Round("KWD", 0)
	assertions.NoError(err)
	assertions.Zero(rs)
}

func Test_Round_OK(t *testing.T) {
	for _, testcase := range []struct {
		name          string
		currency      string
		amount        float64
		roundedAmount float64
	}{
		{
			name:          "factor of 1, rounded-down",
			currency:      "JPY",
			amount:        100.1234567,
			roundedAmount: 100,
		},
		{
			name:          "factor of 1, rounded-up",
			currency:      "JPY",
			amount:        100.98765,
			roundedAmount: 101,
		},
		{
			name:          "factor of 100, rounded-down",
			currency:      "AED",
			amount:        100.1234567,
			roundedAmount: 100.12,
		},
		{
			name:          "factor of 100, rounded-up",
			currency:      "AED",
			amount:        100.1256789,
			roundedAmount: 100.13,
		},
		{
			name:          "factor of 1000, rounded-down",
			currency:      "BHD",
			amount:        100.1234567,
			roundedAmount: 100.123,
		},
		{
			name:          "factor of 1000, rounded-up",
			currency:      "BHD",
			amount:        100.1236789,
			roundedAmount: 100.124,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			assertions := assert.New(t)

			res, err := currency.Round(testcase.currency, testcase.amount)
			assertions.NoError(err)

			assertions.Equal(testcase.roundedAmount, res)
			assertions.Equal(testcase.roundedAmount, currency.Amount(testcase.currency, testcase.amount))
			assertions.True(currency.Equal(testcase.currency, testcase.amount, testcase.roundedAmount))
		})
	}
}

func TestRoundWithSign(t *testing.T) {
	tests := []struct {
		name         string
		currencyCode string
		amount       float64
		want         float64
		wantErr      bool
	}{
		{
			name:         "Positive USD amount",
			currencyCode: currency.USD,
			amount:       123.456,
			want:         123.46,
			wantErr:      false,
		},
		{
			name:         "Negative USD amount",
			currencyCode: currency.USD,
			amount:       -123.456,
			want:         -123.46,
			wantErr:      false,
		},
		{
			name:         "Zero amount",
			currencyCode: currency.USD,
			amount:       0,
			want:         0,
			wantErr:      false,
		},
		{
			name:         "JPY positive amount",
			currencyCode: currency.JPY,
			amount:       123.456,
			want:         123,
			wantErr:      false,
		},
		{
			name:         "JPY negative amount",
			currencyCode: currency.JPY,
			amount:       -123.456,
			want:         -123,
			wantErr:      false,
		},
		{
			name:         "KWD positive amount (3 decimal places)",
			currencyCode: currency.KWD,
			amount:       123.4567,
			want:         123.460,
			wantErr:      false,
		},
		{
			name:         "KWD negative amount (3 decimal places)",
			currencyCode: currency.KWD,
			amount:       -123.4567,
			want:         -123.460,
			wantErr:      false,
		},
		{
			name:         "Invalid currency code",
			currencyCode: "YYY",
			amount:       123.456,
			want:         0,
			wantErr:      true,
		},
		{
			name:         "Very small positive number",
			currencyCode: currency.USD,
			amount:       0.0000001,
			want:         0,
			wantErr:      false,
		},
		{
			name:         "Very small negative number",
			currencyCode: currency.USD,
			amount:       -0.0000001,
			want:         0,
			wantErr:      false,
		},
		{
			name:         "Large positive number with factor 100",
			currencyCode: currency.USD,
			amount:       999999.999,
			want:         1000000.00,
			wantErr:      false,
		},
		{
			name:         "Large negative number with factor 100",
			currencyCode: currency.USD,
			amount:       -999999.999,
			want:         -1000000.00,
			wantErr:      false,
		},
		{
			name:         "Large positive number with factor 1",
			currencyCode: currency.JPY,
			amount:       999999.999,
			want:         1000000,
			wantErr:      false,
		},
		{
			name:         "Large negative number with factor 1",
			currencyCode: currency.JPY,
			amount:       -999999.999,
			want:         -1000000,
			wantErr:      false,
		},
		{
			name:         "KWD Large positive number with factor 1000",
			currencyCode: currency.KWD,
			amount:       999999.999,
			want:         1000000.000,
			wantErr:      false,
		},
		{
			name:         "KWD Large negative number with factor 1000",
			currencyCode: currency.KWD,
			amount:       -999999.999,
			want:         -1000000.000,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := currency.RoundWithSign(tt.currencyCode, tt.amount)
			if tt.wantErr && err == nil {
				t.Errorf("RoundWithSign() wantErr")
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("RoundWithSign() = %v, want %v", got, tt.want)
			}
		})
	}
}
