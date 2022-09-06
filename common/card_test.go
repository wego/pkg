package common_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/common"
)

func Test_ValidateCardNumber(t *testing.T) {
	assert := assert.New(t)

	assert.True(common.ValidateCardNumber("4242424242424242"))
	assert.True(common.ValidateCardNumber("341111111111111"))
	assert.True(common.ValidateCardNumber("378282246310005"))
	assert.True(common.ValidateCardNumber("371449635398431"))
	assert.True(common.ValidateCardNumber("378734493671000"))
	assert.True(common.ValidateCardNumber("30569309025904"))
	assert.True(common.ValidateCardNumber("38520000023237"))
	assert.True(common.ValidateCardNumber("6011601160116611"))
	assert.True(common.ValidateCardNumber("6011111111111117"))
	assert.True(common.ValidateCardNumber("6011000990139424"))
	assert.True(common.ValidateCardNumber("3530111333300000"))
	assert.True(common.ValidateCardNumber("3566002020360505"))
	assert.True(common.ValidateCardNumber("5431111111111111"))
	assert.True(common.ValidateCardNumber("5555555555554444"))
	assert.True(common.ValidateCardNumber("5105105105105100"))
	assert.True(common.ValidateCardNumber("4111111111111111"))
	assert.True(common.ValidateCardNumber("4012888888881881"))
	assert.True(common.ValidateCardNumber("4222222222222"))
	assert.False(common.ValidateCardNumber("1234567812345678"))
}

func TestGenerateCardNumberFromBin(t *testing.T) {
	type args struct {
		bin           string
		cardNumberLen int
	}
	tests := []struct {
		name string
		args args
	}{
		{"generate visa card number", args{"409636", 16}},
		{"generate amex card number", args{"376212", 15}},
		{"generate mastercard card number", args{"547071", 16}},
		{"generate random card number", args{"123456", 16}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if cardNum := common.GenerateCardNumberFromBin(tt.args.bin, tt.args.cardNumberLen); !common.ValidateCardNumber(cardNum) {
				t.Errorf("GenerateCardNumberFromBin() = %v card number is invalid", cardNum)
			}
		})
	}
}
