package currency

import (
	"fmt"
	"math"
	"strings"
)

var iso4217Currencies = map[string]bool{
	"AED": true, "AFN": true, "ALL": true, "AMD": true, "ANG": true,
	"AOA": true, "ARS": true, "AUD": true, "AWG": true, "AZN": true,
	"BAM": true, "BBD": true, "BDT": true, "BGN": true, "BHD": true,
	"BIF": true, "BMD": true, "BND": true, "BOB": true, "BOV": true,
	"BRL": true, "BSD": true, "BTN": true, "BWP": true, "BYN": true,
	"BZD": true, "CAD": true, "CDF": true, "CHE": true, "CHF": true,
	"CHW": true, "CLF": true, "CLP": true, "CNY": true, "COP": true,
	"COU": true, "CRC": true, "CUC": true, "CUP": true, "CVE": true,
	"CZK": true, "DJF": true, "DKK": true, "DOP": true, "DZD": true,
	"EGP": true, "ERN": true, "ETB": true, "EUR": true, "FJD": true,
	"FKP": true, "GBP": true, "GEL": true, "GHS": true, "GIP": true,
	"GMD": true, "GNF": true, "GTQ": true, "GYD": true, "HKD": true,
	"HNL": true, "HRK": true, "HTG": true, "HUF": true, "IDR": true,
	"ILS": true, "INR": true, "IQD": true, "IRR": true, "ISK": true,
	"JMD": true, "JOD": true, "JPY": true, "KES": true, "KGS": true,
	"KHR": true, "KMF": true, "KPW": true, "KRW": true, "KWD": true,
	"KYD": true, "KZT": true, "LAK": true, "LBP": true, "LKR": true,
	"LRD": true, "LSL": true, "LYD": true, "MAD": true, "MDL": true,
	"MGA": true, "MKD": true, "MMK": true, "MNT": true, "MOP": true,
	"MRU": true, "MUR": true, "MVR": true, "MWK": true, "MXN": true,
	"MXV": true, "MYR": true, "MZN": true, "NAD": true, "NGN": true,
	"NIO": true, "NOK": true, "NPR": true, "NZD": true, "OMR": true,
	"PAB": true, "PEN": true, "PGK": true, "PHP": true, "PKR": true,
	"PLN": true, "PYG": true, "QAR": true, "RON": true, "RSD": true,
	"RUB": true, "RWF": true, "SAR": true, "SBD": true, "SCR": true,
	"SDG": true, "SEK": true, "SGD": true, "SHP": true, "SLL": true,
	"SOS": true, "SRD": true, "SSP": true, "STN": true, "SVC": true,
	"SYP": true, "SZL": true, "THB": true, "TJS": true, "TMT": true,
	"TND": true, "TOP": true, "TRY": true, "TTD": true, "TWD": true,
	"TZS": true, "UAH": true, "UGX": true, "USD": true, "USN": true,
	"UYI": true, "UYU": true, "UYW": true, "UZS": true, "VES": true,
	"VND": true, "VUV": true, "WST": true, "XAF": true, "XAG": true,
	"XAU": true, "XBA": true, "XBB": true, "XBC": true, "XBD": true,
	"XCD": true, "XDR": true, "XOF": true, "XPD": true, "XPF": true,
	"XPT": true, "XSU": true, "XTS": true, "XUA": true, "XXX": true,
	"YER": true, "ZAR": true, "ZMW": true, "ZWL": true,
}

//  https://www.checkout.com/docs/resources/calculating-the-value
const defaultCurrencyFactor float64 = 100

var currencyFactors = map[string]float64{
	// Currencies have full value
	"BIF": 1, // Burundian Franc
	"CLF": 1, // Chilean Unidad de Fomentos
	"DJF": 1, // Djiboutian Franc
	"GNF": 1, // Guinean Franc
	"ISK": 1, // Icelandic Krona
	"JPY": 1, // Japanese Yen
	"KMF": 1, // Comoran Franc
	"KRW": 1, // South Korean Won
	"PYG": 1, // Paraguayan Guarani
	"RWF": 1, // Rwandan Franc
	"UGX": 1, // Ugandan Shilling
	"VUV": 1, // Vanuatu Vatu
	"VND": 1, // Vietnamese Dong
	"XAF": 1, // Central African Franc
	"XOF": 1, // West African CFA franc
	"XPF": 1, // Comptoirs Français du Pacifique

	// Currencies have value divided by 1000
	"BHD": 1000, // Bahraini Dinar
	"IQD": 1000, // Iraqi Dinar
	"JOD": 1000, // Jordanian Dinar
	"KWD": 1000, // Kuwaiti Dinar
	"LYD": 1000, // Libyan Dinar
	"OMR": 1000, // Omani Rial
	"TND": 1000, // Tunisian Dinar
}

// ToMinorUnit converts amount in a currency to amount in smallest unit (minor unit)
// https://en.wikipedia.org/wiki/ISO_4217#Minor_units_of_currency
func ToMinorUnit(currencyCode string, amount float64) (minorUnitAmount uint64, err error) {
	if !IsISO4217(currencyCode) {
		err = fmt.Errorf("%s is not a valid ISO 4217 currency code", currencyCode)
		return
	}
	switch {
	case amount > 0:
		factor := getCurrencyFactor(currencyCode)
		minorUnitAmount = uint64(math.Round(amount * factor))
	case amount == 0:
		minorUnitAmount = 0
	default:
		err = fmt.Errorf("invalid amount: %f", amount)
	}
	return
}

// FromMinorUnit converts amount in the smallest unit (minor unit) of a currency to amount in that currency
func FromMinorUnit(currencyCode string, minorUnitAmount uint64) (amount float64, err error) {
	if !IsISO4217(currencyCode) {
		err = fmt.Errorf("%s is not a valid ISO 4217 currency code", currencyCode)
		return
	}

	if minorUnitAmount > 0 {
		factor := getCurrencyFactor(currencyCode)
		amount = float64(minorUnitAmount) / factor
		amount = math.Round(amount*factor) / factor
	}
	return
}

// IsISO4217 check a currency code is an ISO 4217 currency code
func IsISO4217(code string) bool {
	return iso4217Currencies[strings.ToUpper(strings.TrimSpace(code))]
}

func getCurrencyFactor(currency string) (factor float64) {
	factor, ok := currencyFactors[strings.ToUpper(currency)]
	if !ok {
		factor = defaultCurrencyFactor
	}
	return
}
