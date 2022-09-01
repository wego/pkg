package common

import "math/rand"

// ValidateCardNumber will check the credit card's number against the Luhn algorithm
func ValidateCardNumber(number string) bool {
	var sum int
	var alternate bool

	// Gets the Card number length
	numberLen := len(number)

	// For numbers that is lower than 13 and
	// bigger than 19, must return as false
	if numberLen < 13 || numberLen > 19 {
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

// Generate a random number with the specified length
func Generate(bin string, length int) string {
	result := make([]byte, length)
	binLen := len(bin)
	for i := 0; i < binLen; i++ {
		result[i] = bin[i]
	}
	remainingLen := length - (len(bin) + 1)
	for i := 0; i < remainingLen; i++ {
		result[binLen+i] = byte(rand.Intn(9) + 48)
	}
	result[length-1] = byte(getCheckDigit(result, length-1) + 48)
	return string(result)
}

func getCheckDigit(number []byte, end int) int {
	var sum int
	for i := 0; i < end; i++ {

		// Get the digit at the current position.
		digit := int(number[i] - '0')

		if (i % 2) == 0 {
			digit = digit * 2
			if digit > 9 {
				digit = (digit / 10) + (digit % 10)
			}
		}
		sum += digit
	}

	mod := sum % 10
	if mod == 0 {
		return 0
	}
	return 10 - mod
}
