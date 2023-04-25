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
