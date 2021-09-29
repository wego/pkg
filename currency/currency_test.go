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

func Test_AmountToAmountInCents_Ok(t *testing.T) {
	assert := assert.New(t)

	rs, err := currency.AmountToAmountInCents("SGD", 25.12)
	assert.NoError(err)
	assert.EqualValues(2512, rs)

	rs, err = currency.AmountToAmountInCents("SGD", 25.123)
	assert.NoError(err)
	assert.EqualValues(2512, rs)

	rs, err = currency.AmountToAmountInCents("BHD", 25.123)
	assert.NoError(err)
	assert.EqualValues(25123, rs)

	rs, err = currency.AmountToAmountInCents("BHD", 25.1234)
	assert.NoError(err)
	assert.EqualValues(25123, rs)

	rs, err = currency.AmountToAmountInCents("VND", 25.12345)
	assert.NoError(err)
	assert.EqualValues(25, rs)

	rs, err = currency.AmountToAmountInCents("Vnd", 25.12345)
	assert.NoError(err)
	assert.EqualValues(25, rs)
}

func Test_AmountToAmountInCents_InvalidCurrency(t *testing.T) {
	assert := assert.New(t)

	rs, err := currency.AmountToAmountInCents("", 25.12)
	assert.Error(err)
	assert.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assert.Zero(rs)

	rs, err = currency.AmountToAmountInCents("sg", 25.12)
	assert.Error(err)
	assert.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assert.Zero(rs)

	rs, err = currency.AmountToAmountInCents(" ", 25.12)
	assert.Error(err)
	assert.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assert.Zero(rs)

	rs, err = currency.AmountToAmountInCents("      ", 25.12)
	assert.Error(err)
	assert.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assert.Zero(rs)
}

func Test_AmountToAmountInCents_ZeroAmount(t *testing.T) {
	assert := assert.New(t)

	rs, err := currency.AmountToAmountInCents("SGD", 0)
	assert.NoError(err)
	assert.Zero(rs)
}

func Test_AmountToAmountInCents_InvalidAmount(t *testing.T) {
	assert := assert.New(t)

	rs, err := currency.AmountToAmountInCents("USD", -0.0001)
	assert.Error(err)
	assert.Contains(err.Error(), "invalid amount")
	assert.Zero(rs)

	rs, err = currency.AmountToAmountInCents("SGd", -0.0001)
	assert.Error(err)
	assert.Contains(err.Error(), "invalid amount")
	assert.Zero(rs)
}

func Test_AmountInCentsToAmount_Ok(t *testing.T) {
	assert := assert.New(t)

	rs, err := currency.AmountInCentsToAmount("SGD", 2512)
	assert.NoError(err)
	assert.EqualValues(25.12, rs)

	rs, err = currency.AmountInCentsToAmount("SGD", 25123)
	assert.NoError(err)
	assert.EqualValues(251.23, rs)

	rs, err = currency.AmountInCentsToAmount("BHD", 25123)
	assert.NoError(err)
	assert.EqualValues(25.123, rs)

	rs, err = currency.AmountInCentsToAmount("BHD", 251234)
	assert.NoError(err)
	assert.EqualValues(251.234, rs)

	rs, err = currency.AmountInCentsToAmount("VND", 2512345)
	assert.NoError(err)
	assert.EqualValues(2512345, rs)

	rs, err = currency.AmountInCentsToAmount("Vnd", 2512345)
	assert.NoError(err)
	assert.EqualValues(2512345, rs)
}

func Test_AmountInCentsToAmount_InvalidCurrency(t *testing.T) {
	assert := assert.New(t)

	rs, err := currency.AmountInCentsToAmount("", 2512)
	assert.Error(err)
	assert.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assert.Zero(rs)

	rs, err = currency.AmountInCentsToAmount("sg", 2512)
	assert.Error(err)
	assert.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assert.Zero(rs)

	rs, err = currency.AmountInCentsToAmount(" ", 2512)
	assert.Error(err)
	assert.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assert.Zero(rs)

	rs, err = currency.AmountInCentsToAmount("      ", 2512)
	assert.Error(err)
	assert.Contains(err.Error(), "is not a valid ISO 4217 currency code")
	assert.Zero(rs)
}

func Test_AmountInCentsToAmount_ZeroAmount(t *testing.T) {
	assert := assert.New(t)

	rs, err := currency.AmountInCentsToAmount("USD", 0)
	assert.NoError(err)
	assert.Zero(rs)

	rs, err = currency.AmountInCentsToAmount("SGD", 0)
	assert.NoError(err)
	assert.Zero(rs)
}
