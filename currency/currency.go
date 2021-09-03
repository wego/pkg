package currency

import (
	"fmt"
	"math"
	"strings"
)

//  https://docs.checkout.com/docs/calculating-the-value
const defaultCurrencyFactor = 100

var currencyFactors = map[string]uint{
	// Currencies have full value
	"BIF": 1, // Burundian Franc
	"DJF": 1, // Djiboutian Franc
	"GNF": 1, // Guinean Franc
	"ISK": 1, // Icelandic Krona
	"KMF": 1, // Comoran Franc
	"XAF": 1, // Central African Franc
	"CLF": 1, // Chilean Unidad de Fomentos
	"XPF": 1, // Comptoirs Français du Pacifique
	"JPY": 1, // Japanese Yen
	"PYG": 1, // Paraguayan Guarani
	"RWF": 1, // Rwandan Franc
	"KRW": 1, // South Korean Won
	"VUV": 1, // Vanuatu Vatu
	"VND": 1, // Vietnamese Dong
	"XOF": 1, // West African CFA franc

	// Currencies have value divided by 1000
	"BHD": 1000, // Bahraini Dinar
	"IQD": 1000, // Iraqi Dinar
	"JOD": 1000, // Jordanian Dinar
	"KWD": 1000, // Kuwaiti Dinar
	"LYD": 1000, // Libyan Dinar
	"OMR": 1000, // Omani Rial
	"TND": 1000, // Tunisian Dinar
}

// AmountToAmountInCents converts amount in a currency to amount in smallest unit of that currency
func AmountToAmountInCents(currencyCode string, amount float64) (amountInCents uint64, err error) {
	switch {
	case amount > 0:
		factor, e := getCurrencyFactor(currencyCode)
		if e != nil {
			return 0, e
		}
		amountInCents = uint64(math.Round(amount * float64(factor)))
	case amount == 0:
		amountInCents = 0
	default:
		err = fmt.Errorf("invalid amount: %f", amount)
	}
	return
}

// AmountInCentsToAmount converts amount in smallest unit of a currency to amount in that currency
func AmountInCentsToAmount(currency string, amountInCents uint64) (amount float64, err error) {
	if amountInCents > 0 {
		factor, e := getCurrencyFactor(currency)
		if e != nil {
			return 0, e
		}

		amount = float64(amountInCents) / float64(factor)
		amount = math.Round(amount*float64(factor)) / float64(factor)
	}
	return
}

func getCurrencyFactor(currency string) (factor uint, err error) {
	if len(strings.TrimSpace(currency)) != 3 {
		err = fmt.Errorf("invalid currency: %s", currency)
		return
	}

	factor, ok := currencyFactors[strings.ToUpper(currency)]
	if !ok {
		factor = defaultCurrencyFactor
	}
	return
}
