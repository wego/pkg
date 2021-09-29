package currency

import (
	"fmt"
	"math"
	"strings"
)

var iso4217Currencies = map[string]bool{
	"AFN": true, "EUR": true, "ALL": true, "DZD": true, "USD": true,
	"AOA": true, "XCD": true, "ARS": true, "AMD": true, "AWG": true,
	"AUD": true, "AZN": true, "BSD": true, "BHD": true, "BDT": true,
	"BBD": true, "BYN": true, "BZD": true, "XOF": true, "BMD": true,
	"INR": true, "BTN": true, "BOB": true, "BOV": true, "BAM": true,
	"BWP": true, "NOK": true, "BRL": true, "BND": true, "BGN": true,
	"BIF": true, "CVE": true, "KHR": true, "XAF": true, "CAD": true,
	"KYD": true, "CLP": true, "CLF": true, "CNY": true, "COP": true,
	"COU": true, "KMF": true, "CDF": true, "NZD": true, "CRC": true,
	"HRK": true, "CUP": true, "CUC": true, "ANG": true, "CZK": true,
	"DKK": true, "DJF": true, "DOP": true, "EGP": true, "SVC": true,
	"ERN": true, "SZL": true, "ETB": true, "FKP": true, "FJD": true,
	"XPF": true, "GMD": true, "GEL": true, "GHS": true, "GIP": true,
	"GTQ": true, "GBP": true, "GNF": true, "GYD": true, "HTG": true,
	"HNL": true, "HKD": true, "HUF": true, "ISK": true, "IDR": true,
	"XDR": true, "IRR": true, "IQD": true, "ILS": true, "JMD": true,
	"JPY": true, "JOD": true, "KZT": true, "KES": true, "KPW": true,
	"KRW": true, "KWD": true, "KGS": true, "LAK": true, "LBP": true,
	"LSL": true, "ZAR": true, "LRD": true, "LYD": true, "CHF": true,
	"MOP": true, "MKD": true, "MGA": true, "MWK": true, "MYR": true,
	"MVR": true, "MRU": true, "MUR": true, "XUA": true, "MXN": true,
	"MXV": true, "MDL": true, "MNT": true, "MAD": true, "MZN": true,
	"MMK": true, "NAD": true, "NPR": true, "NIO": true, "NGN": true,
	"OMR": true, "PKR": true, "PAB": true, "PGK": true, "PYG": true,
	"PEN": true, "PHP": true, "PLN": true, "QAR": true, "RON": true,
	"RUB": true, "RWF": true, "SHP": true, "WST": true, "STN": true,
	"SAR": true, "RSD": true, "SCR": true, "SLL": true, "SGD": true,
	"XSU": true, "SBD": true, "SOS": true, "SSP": true, "LKR": true,
	"SDG": true, "SRD": true, "SEK": true, "CHE": true, "CHW": true,
	"SYP": true, "TWD": true, "TJS": true, "TZS": true, "THB": true,
	"TOP": true, "TTD": true, "TND": true, "TRY": true, "TMT": true,
	"UGX": true, "UAH": true, "AED": true, "USN": true, "UYU": true,
	"UYI": true, "UYW": true, "UZS": true, "VUV": true, "VES": true,
	"VND": true, "YER": true, "ZMW": true, "ZWL": true, "XBA": true,
	"XBB": true, "XBC": true, "XBD": true, "XTS": true, "XXX": true,
	"XAU": true, "XPD": true, "XPT": true, "XAG": true,
}

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
	if !IsISO4217(currencyCode) {
		err = fmt.Errorf("%s is not a valid ISO 4217 currency code", currencyCode)
		return
	}
	switch {
	case amount > 0:
		factor := getCurrencyFactor(currencyCode)
		amountInCents = uint64(math.Round(amount * float64(factor)))
	case amount == 0:
		amountInCents = 0
	default:
		err = fmt.Errorf("invalid amount: %f", amount)
	}
	return
}

// AmountInCentsToAmount converts amount in the smallest unit of a currency to amount in that currency
func AmountInCentsToAmount(currencyCode string, amountInCents uint64) (amount float64, err error) {
	if !IsISO4217(currencyCode) {
		err = fmt.Errorf("%s is not a valid ISO 4217 currency code", currencyCode)
		return
	}
	switch {
	case amountInCents > 0:
		factor := getCurrencyFactor(currencyCode)
		amount = float64(amountInCents) / float64(factor)
		amount = math.Round(amount*float64(factor)) / float64(factor)
		return
	case amountInCents == 0:
		amount = 0
	default:
		err = fmt.Errorf("invalid amount: %f", amount)
	}
	return
}

// IsISO4217 check a currency code is an ISO 4217 currency code
func IsISO4217(code string) bool {
	return iso4217Currencies[strings.ToUpper(strings.TrimSpace(code))]
}

func getCurrencyFactor(currency string) (factor uint) {
	factor, ok := currencyFactors[strings.ToUpper(currency)]
	if !ok {
		factor = defaultCurrencyFactor
	}
	return
}
