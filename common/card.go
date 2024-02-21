package common

import (
	"github.com/Ardesco/credit-card-generator/api/cards"
	"github.com/Ardesco/credit-card-generator/api/models"
)

// ValidateCardNumber will check the credit card's number against the Luhn algorithm
func ValidateCardNumber(number string) bool {
	var sum int
	var alternate bool

	// Gets the Card number length
	numberLen := len(number)

	// For numbers that is lower than 8 and bigger than 19, must return as false
	if numberLen < 8 || numberLen > 19 {
		return false
	}

	// Parse all numbers of the card into a for loop
	for i := numberLen - 1; i > -1; i-- {
		chr := number[i]
		if chr < '0' || chr > '9' {
			return false
		}

		// Takes the digit, converting the current number in integer
		digit := int(chr - '0')
		if alternate {
			digit *= 2
		}
		sum += digit / 10
		sum += digit % 10
		alternate = !alternate
	}

	return sum%10 == 0
}

// GenerateCardNumberFromBin will generate credit card number based on bin number and the card number length
func GenerateCardNumberFromBin(bin string, cardNumberLen int) string {
	return cards.GeneratePAN(models.CardProperties{
		Prefix:  []string{bin},
		PanSize: cardNumberLen,
	}).Raw
}
