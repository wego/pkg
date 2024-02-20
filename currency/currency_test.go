package currency_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/currency"
)

func Test_IsISO4217(t *testing.T) {
	assert := assert.New(t)
	assert.True(currency.IsISO4217("afn"))
	assert.True(currency.IsISO4217("UsD"))
	assert.True(currency.IsISO4217("UsD "))
	assert.True(currency.IsISO4217("CNY"))
	assert.True(currency.IsISO4217("KWD"))
	assert.False(currency.IsISO4217("UsDD"))
	assert.False(currency.IsISO4217(""))
}

func Test_ToMinorUnit_OK(t *testing.T) {
	assert := assert.New(t)

	rs, err := currency.ToMinorUnit("SGD", 25.12)
	assert.NoError(err)
	assert.EqualValues(2512, rs)

	rs, err = currency.ToMinorUnit("SGD", 25.123)
	assert.NoError(err)
	assert.EqualValues(2512, rs)

	rs, err = currency.ToMinorUnit("BHD", 25.123)
	assert.NoError(err)
	assert.EqualValues(25123, rs)

	rs, err = currency.ToMinorUnit("BHD", 25.1234)
	assert.NoError(err)
	assert.EqualValues(25123, rs)

	rs, err = currency.ToMinorUnit("VND", 25.12345)
	assert.NoError(err)
	assert.EqualValues(25, rs)

	rs, err = currency.ToMinorUnit("Vnd", 25.12345)
	assert.NoError(err)
	assert.EqualValues(25, rs)
}

func Test_ToMinorUnit_InvalidCurrency(t *testing.T) {
	assert := assert.New(t)

	rs, err := currency.ToMinorUnit("", 25.12)
	assert.Error(err)
	assert.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assert.Zero(rs)

	rs, err = currency.ToMinorUnit("sg", 25.12)
	assert.Error(err)
	assert.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assert.Zero(rs)

	rs, err = currency.ToMinorUnit(" ", 25.12)
	assert.Error(err)
	assert.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assert.Zero(rs)

	rs, err = currency.ToMinorUnit("      ", 25.12)
	assert.Error(err)
	assert.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assert.Zero(rs)
}

func Test_ToMinorUnit_ZeroAmount(t *testing.T) {
	assert := assert.New(t)

	rs, err := currency.ToMinorUnit("SGD", 0)
	assert.NoError(err)
	assert.Zero(rs)
}

func Test_ToMinorUnit_InvalidAmount(t *testing.T) {
	assert := assert.New(t)

	rs, err := currency.ToMinorUnit("USD", -0.0001)
	assert.Error(err)
	assert.Contains(err.Error(), "invalid amount")
	assert.Zero(rs)

	rs, err = currency.ToMinorUnit("SGd", -0.0001)
	assert.Error(err)
	assert.Contains(err.Error(), "invalid amount")
	assert.Zero(rs)
}

func Test_FromMinorUnit_OK(t *testing.T) {
	assert := assert.New(t)

	rs, err := currency.FromMinorUnit("SGD", 2599)
	assert.NoError(err)
	assert.EqualValues(25.99, rs)

	rs, err = currency.FromMinorUnit("SGD", 25123)
	assert.NoError(err)
	assert.EqualValues(251.23, rs)

	rs, err = currency.FromMinorUnit("BHD", 25123)
	assert.NoError(err)
	assert.EqualValues(25.123, rs)

	rs, err = currency.FromMinorUnit("BHD", 251987)
	assert.NoError(err)
	assert.EqualValues(251.987, rs)

	rs, err = currency.FromMinorUnit("VND", 2512345)
	assert.NoError(err)
	assert.EqualValues(2512345, rs)

	rs, err = currency.FromMinorUnit("Vnd", 2512345)
	assert.NoError(err)
	assert.EqualValues(2512345, rs)
}

func Test_FromMinorUnit_InvalidCurrency(t *testing.T) {
	assert := assert.New(t)

	rs, err := currency.FromMinorUnit("", 2512)
	assert.Error(err)
	assert.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assert.Zero(rs)

	rs, err = currency.FromMinorUnit("sg", 2512)
	assert.Error(err)
	assert.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assert.Zero(rs)

	rs, err = currency.FromMinorUnit(" ", 2512)
	assert.Error(err)
	assert.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assert.Zero(rs)

	rs, err = currency.FromMinorUnit("      ", 2512)
	assert.Error(err)
	assert.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assert.Zero(rs)
}

func Test_FromMinorUnit_ZeroAmount(t *testing.T) {
	assert := assert.New(t)

	rs, err := currency.FromMinorUnit("USD", 0)
	assert.NoError(err)
	assert.Zero(rs)

	rs, err = currency.FromMinorUnit("VND", 0)
	assert.NoError(err)
	assert.Zero(rs)

	rs, err = currency.FromMinorUnit("KWD", 0)
	assert.NoError(err)
	assert.Zero(rs)
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
	assert := assert.New(t)

	factorOf100 := []string{"AED", "USD", "PHP", "qwerty", "1234", "SAR"}
	factorOf1000 := []string{"BHD", "IQD", "JOD", "KWD", "LYD", "OMR", "TND"}
	factorOfOne := []string{"BIF", "CLF", "DJF", "GNF", "ISK", "JPY", "KMF",
		"KRW", "PYG", "RWF", "UGX", "VUV", "VND", "XAF", "XOF", "XPF"}

	for _, cur := range factorOf100 {
		factor := currency.GetCurrencyFactor(cur)
		assert.Equal(100.0, factor)
	}

	for _, cur := range factorOf1000 {
		factor := currency.GetCurrencyFactor(cur)
		assert.Equal(1000.0, factor)
	}

	for _, cur := range factorOfOne {
		factor := currency.GetCurrencyFactor(cur)
		assert.Equal(1.0, factor)
	}
}

func Test_Round_InvalidCurrency(t *testing.T) {
	assert := assert.New(t)

	rs, err := currency.Round("", 2512)
	assert.Error(err)
	assert.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assert.Zero(rs)

	rs, err = currency.Round("sg", 2512)
	assert.Error(err)
	assert.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assert.Zero(rs)

	rs, err = currency.Round(" ", 2512)
	assert.Error(err)
	assert.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assert.Zero(rs)

	rs, err = currency.Round("      ", 2512)
	assert.Error(err)
	assert.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assert.Zero(rs)
}

func Test_Round_InvalidAmount(t *testing.T) {
	assert := assert.New(t)

	rs, err := currency.Round("USD", -0.0001)
	assert.Error(err)
	assert.Contains(err.Error(), "invalid amount")
	assert.Zero(rs)

	rs, err = currency.Round("SGd", -0.0001)
	assert.Error(err)
	assert.Contains(err.Error(), "invalid amount")
	assert.Zero(rs)
}

func Test_Round_ZeroAmount(t *testing.T) {
	assert := assert.New(t)

	rs, err := currency.Round("USD", 0)
	assert.NoError(err)
	assert.Zero(rs)

	rs, err = currency.Round("VND", 0)
	assert.NoError(err)
	assert.Zero(rs)

	rs, err = currency.Round("KWD", 0)
	assert.NoError(err)
	assert.Zero(rs)
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
			assert := assert.New(t)

			res, err := currency.Round(testcase.currency, testcase.amount)
			assert.NoError(err)

			assert.Equal(testcase.roundedAmount, res)
		})
	}
}
