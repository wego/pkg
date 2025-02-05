// Package currency provides functions to convert & format currency amount.
package currency

import (
	"fmt"
	"math"
	"strings"

	"github.com/bojanz/currency"
)

// ISO4217 currencies
const (
	AED = "AED" // United Arab Emirates Dirham
	AFN = "AFN" // Afghan Afghani
	ALL = "ALL" // Albanian Lek
	AMD = "AMD" // Armenian Dram
	ANG = "ANG" // Netherlands Antillean Guilder
	AOA = "AOA" // Angolan Kwanza
	ARS = "ARS" // Argentine Peso
	AUD = "AUD" // Australian Dollar
	AWG = "AWG" // Aruban Florin
	AZN = "AZN" // Azerbaijani Manat
	BAM = "BAM" // Bosnia and Herzegovina Convertible Mark
	BBD = "BBD" // Barbadian Dollar
	BDT = "BDT" // Bangladeshi Taka
	BGN = "BGN" // Bulgarian Lev
	BHD = "BHD" // Bahraini Dinar
	BIF = "BIF" // Burundian Franc
	BMD = "BMD" // Bermudian Dollar
	BND = "BND" // Brunei Dollar
	BOB = "BOB" // Bolivian Boliviano
	BOV = "BOV" // Bolivian Mvdol
	BRL = "BRL" // Brazilian Real
	BSD = "BSD" // Bahamian Dollar
	BTN = "BTN" // Bhutanese Ngultrum
	BWP = "BWP" // Botswana Pula
	BYN = "BYN" // Belarusian Ruble
	BZD = "BZD" // Belize Dollar
	CAD = "CAD" // Canadian Dollar
	CDF = "CDF" // Congolese Franc
	CHE = "CHE" // WIR Euro
	CHF = "CHF" // Swiss Franc
	CHW = "CHW" // WIR Franc
	CLF = "CLF" // Unidad de Fomento (funds code)
	CLP = "CLP" // Chilean Peso
	CNY = "CNY" // Chinese Yuan
	COP = "COP" // Colombian Peso
	COU = "COU" // Colombian Real Value Unit
	CRC = "CRC" // Costa Rican Colón
	CUC = "CUC" // Cuban Convertible Peso
	CUP = "CUP" // Cuban Peso
	CVE = "CVE" // Cape Verdean Escudo
	CZK = "CZK" // Czech Republic Koruna
	DJF = "DJF" // Djiboutian Franc
	DKK = "DKK" // Danish Krone
	DOP = "DOP" // Dominican Peso
	DZD = "DZD" // Algerian Dinar
	EGP = "EGP" // Egyptian Pound
	ERN = "ERN" // Eritrean Nakfa
	ETB = "ETB" // Ethiopian Birr
	EUR = "EUR" // Euro
	FJD = "FJD" // Fijian Dollar
	FKP = "FKP" // Falkland Islands Pound
	GBP = "GBP" // British Pound Sterling
	GEL = "GEL" // Georgian Lari
	GHS = "GHS" // Ghanaian Cedi
	GIP = "GIP" // Gibraltar Pound
	GMD = "GMD" // Gambian Dalasi
	GNF = "GNF" // Guinean Franc
	GTQ = "GTQ" // Guatemalan Quetzal
	GYD = "GYD" // Guyanese Dollar
	HKD = "HKD" // Hong Kong Dollar
	HNL = "HNL" // Honduran Lempira
	HRK = "HRK" // Croatian Kuna
	HTG = "HTG" // Haitian Gourde
	HUF = "HUF" // Hungarian Forint
	IDR = "IDR" // Indonesian Rupiah
	ILS = "ILS" // Israeli New Shekel
	INR = "INR" // Indian Rupee
	IQD = "IQD" // Iraqi Dinar
	IRR = "IRR" // Iranian Rial
	ISK = "ISK" // Icelandic Krona
	JMD = "JMD" // Jamaican Dollar
	JOD = "JOD" // Jordanian Dinar
	JPY = "JPY" // Japanese Yen
	KES = "KES" // Kenyan Shilling
	KGS = "KGS" // Kyrgyzstani Som
	KHR = "KHR" // Cambodian Riel
	KMF = "KMF" // Comoran Franc
	KPW = "KPW" // North Korean Won
	KRW = "KRW" // South Korean Won
	KWD = "KWD" // Kuwaiti Dinar
	KYD = "KYD" // Cayman Islands Dollar
	KZT = "KZT" // Kazakhstani Tenge
	LAK = "LAK" // Lao Kip
	LBP = "LBP" // Lebanese Pound
	LKR = "LKR" // Sri Lankan Rupee
	LRD = "LRD" // Liberian Dollar
	LSL = "LSL" // Lesotho Loti
	LYD = "LYD" // Libyan Dinar
	MAD = "MAD" // Moroccan Dirham
	MDL = "MDL" // Moldovan Leu
	MGA = "MGA" // Malagasy Ariary
	MKD = "MKD" // Macedonian Denar
	MMK = "MMK" // Myanmar Kyat
	MNT = "MNT" // Mongolian Tögrög
	MOP = "MOP" // Macanese Pataca
	MRU = "MRU" // Mauritanian Ouguiya
	MUR = "MUR" // Mauritian Rupee
	MVR = "MVR" // Maldivian Rufiyaa
	MWK = "MWK" // Malawian Kwacha
	MXN = "MXN" // Mexican Peso
	MXV = "MXV" // Mexican Investment Unit
	MYR = "MYR" // Malaysian Ringgit
	MZN = "MZN" // Mozambican Metical
	NAD = "NAD" // Namibian Dollar
	NGN = "NGN" // Nigerian Naira
	NIO = "NIO" // Nicaraguan Córdoba
	NOK = "NOK" // Norwegian Krone
	NPR = "NPR" // Nepalese Rupee
	NZD = "NZD" // New Zealand Dollar
	OMR = "OMR" // Omani Rial
	PAB = "PAB" // Panamanian Balboa
	PEN = "PEN" // Peruvian Sol
	PGK = "PGK" // Papua New Guinean Kina
	PHP = "PHP" // Philippine Peso
	PKR = "PKR" // Pakistani Rupee
	PLN = "PLN" // Polish Zloty
	PRB = "PRB" // Transnistrian Ruble
	PYG = "PYG" // Paraguayan Guarani
	QAR = "QAR" // Qatari Riyal
	RON = "RON" // Romanian Leu
	RSD = "RSD" // Serbian Dinar
	RUB = "RUB" // Russian Ruble
	RWF = "RWF" // Rwandan Franc
	SAR = "SAR" // Saudi Riyal
	SBD = "SBD" // Solomon Islands Dollar
	SCR = "SCR" // Seychellois Rupee
	SDG = "SDG" // Sudanese Pound
	SEK = "SEK" // Swedish Krona
	SGD = "SGD" // Singapore Dollar
	SHP = "SHP" // Saint Helena Pound
	SLL = "SLL" // Sierra Leonean Leone
	SOS = "SOS" // Somali Shilling
	SRD = "SRD" // Surinamese Dollar
	SSP = "SSP" // South Sudanese Pound
	STN = "STN" // São Tomé and Príncipe Dobra
	SVC = "SVC" // Salvadoran Colón
	SYP = "SYP" // Syrian Pound
	SZL = "SZL" // Swazi Lilangeni
	THB = "THB" // Thai Baht
	TJS = "TJS" // Tajikistani Somoni
	TMT = "TMT" // Turkmenistani Manat
	TND = "TND" // Tunisian Dinar
	TOP = "TOP" // Tongan Paʻanga
	TRY = "TRY" // Turkish Lira
	TTD = "TTD" // Trinidad and Tobago Dollar
	TWD = "TWD" // New Taiwan Dollar
	TZS = "TZS" // Tanzanian Shilling
	UAH = "UAH" // Ukrainian Hryvnia
	UGX = "UGX" // Ugandan Shilling
	USD = "USD" // United States Dollar
	USN = "USN" // US Dollar (Next day)
	UYI = "UYI" // Uruguayan Peso en Unidades Indexadas (URUIURUI)
	UYU = "UYU" // Uruguayan Peso
	UYW = "UYW" // Unidad previsional
	UZS = "UZS" // Uzbekistani Som
	VES = "VES" // Venezuelan Bolívar Soberano
	VND = "VND" // Vietnamese Dong
	VUV = "VUV" // Vanuatu Vatu
	WST = "WST" // Samoan Tala
	XAF = "XAF" // Central African CFA franc
	XAG = "XAG" // Silver
	XAU = "XAU" // Gold
	XBA = "XBA" // Bond Markets Unit European Composite Unit (EURCO)
	XBB = "XBB" // Bond Markets Unit European Monetary Unit (E.M.U.-6)
	XBC = "XBC" // Bond Markets Unit European Unit of Account 9 (E.U.A.-9)
	XBD = "XBD" // Bond Markets Unit European Unit of Account 17 (E.U.A.-17)
	XCD = "XCD" // East Caribbean Dollar
	XDR = "XDR" // Special Drawing Rights
	XOF = "XOF" // West African CFA franc
	XPD = "XPD" // Palladium
	XPF = "XPF" // Comptoirs Français du Pacifique
	XPT = "XPT" // Platinum
	XSU = "XSU" // Sucre
	XTS = "XTS" // Testing Currency Code
	XUA = "XUA" // ADB Unit of Account
	XXX = "XXX" // The codes assigned for transactions where no currency is involved
	YER = "YER" // Yemeni Rial
	ZAR = "ZAR" // South African Rand
	ZMW = "ZMW" // Zambian Kwacha
	ZWG = "ZWG" // Zimbabwean Dollar (From 20240901)
	ZWL = "ZWL" // Zimbabwean Dollar (2009-20240901)
)

var iso4217Currencies = map[string]bool{
	AED: true, AFN: true, ALL: true, AMD: true, ANG: true,
	AOA: true, ARS: true, AUD: true, AWG: true, AZN: true,
	BAM: true, BBD: true, BDT: true, BGN: true, BHD: true,
	BIF: true, BMD: true, BND: true, BOB: true, BOV: true,
	BRL: true, BSD: true, BTN: true, BWP: true, BYN: true,
	BZD: true, CAD: true, CDF: true, CHE: true, CHF: true,
	CHW: true, CLF: true, CLP: true, CNY: true, COP: true,
	COU: true, CRC: true, CUC: true, CUP: true, CVE: true,
	CZK: true, DJF: true, DKK: true, DOP: true, DZD: true,
	EGP: true, ERN: true, ETB: true, EUR: true, FJD: true,
	FKP: true, GBP: true, GEL: true, GHS: true, GIP: true,
	GMD: true, GNF: true, GTQ: true, GYD: true, HKD: true,
	HNL: true, HRK: true, HTG: true, HUF: true, IDR: true,
	ILS: true, INR: true, IQD: true, IRR: true, ISK: true,
	JMD: true, JOD: true, JPY: true, KES: true, KGS: true,
	KHR: true, KMF: true, KPW: true, KRW: true, KWD: true,
	KYD: true, KZT: true, LAK: true, LBP: true, LKR: true,
	LRD: true, LSL: true, LYD: true, MAD: true, MDL: true,
	MGA: true, MKD: true, MMK: true, MNT: true, MOP: true,
	MRU: true, MUR: true, MVR: true, MWK: true, MXN: true,
	MXV: true, MYR: true, MZN: true, NAD: true, NGN: true,
	NIO: true, NOK: true, NPR: true, NZD: true, OMR: true,
	PAB: true, PEN: true, PGK: true, PHP: true, PKR: true,
	PLN: true, PYG: true, QAR: true, RON: true, RSD: true,
	RUB: true, RWF: true, SAR: true, SBD: true, SCR: true,
	SDG: true, SEK: true, SGD: true, SHP: true, SLL: true,
	SOS: true, SRD: true, SSP: true, STN: true, SVC: true,
	SYP: true, SZL: true, THB: true, TJS: true, TMT: true,
	TND: true, TOP: true, TRY: true, TTD: true, TWD: true,
	TZS: true, UAH: true, UGX: true, USD: true, USN: true,
	UYI: true, UYU: true, UYW: true, UZS: true, VES: true,
	VND: true, VUV: true, WST: true, XAF: true, XAG: true,
	XAU: true, XBA: true, XBB: true, XBC: true, XBD: true,
	XCD: true, XDR: true, XOF: true, XPD: true, XPF: true,
	XPT: true, XSU: true, XTS: true, XUA: true, XXX: true,
	YER: true, ZAR: true, ZMW: true, ZWG: true, ZWL: true,
}

// https://www.checkout.com/docs/payments/accept-payments/format-the-payment-amount
const defaultCurrencyFactor float64 = 100

var currencyFactors = map[string]float64{
	// Currencies have full value
	BIF: 1, // Burundian Franc
	CLF: 1, // Chilean Unidad de Fomentos
	DJF: 1, // Djiboutian Franc
	GNF: 1, // Guinean Franc
	ISK: 1, // Icelandic Krona
	JPY: 1, // Japanese Yen
	KMF: 1, // Comoran Franc
	KRW: 1, // South Korean Won
	PYG: 1, // Paraguayan Guarani
	RWF: 1, // Rwandan Franc
	UGX: 1, // Ugandan Shilling
	VUV: 1, // Vanuatu Vatu
	VND: 1, // Vietnamese Dong
	XAF: 1, // Central African Franc
	XOF: 1, // West African CFA franc
	XPF: 1, // Comptoirs Français du Pacifique

	// Currencies have value divided by 1000
	BHD: 1000, // Bahraini Dinar
	IQD: 1000, // Iraqi Dinar
	JOD: 1000, // Jordanian Dinar
	KWD: 1000, // Kuwaiti Dinar
	LYD: 1000, // Libyan Dinar
	OMR: 1000, // Omani Rial
	TND: 1000, // Tunisian Dinar
}

// ToMinorUnit converts amount in a currency to amount in smallest unit (minor unit)
// https://en.wikipedia.org/wiki/ISO_4217#Minor_units_of_currency
func ToMinorUnit(currencyCode string, amount float64) (minorUnitAmount uint64, err error) {
	if !IsISO4217(currencyCode) {
		return 0, fmt.Errorf("%s is not a valid ISO 4217 currency code", currencyCode)
	}
	switch {
	case amount > 0:
		factor := GetCurrencyFactor(currencyCode)
		minorUnitAmount = uint64(math.Round(amount * factor))
	case amount < 0:
		err = fmt.Errorf("invalid amount: %f", amount)
	}
	return
}

// FromMinorUnit converts amount in the smallest unit (minor unit) of a currency to amount in that currency
func FromMinorUnit(currencyCode string, minorUnitAmount uint64) (amount float64, err error) {
	if !IsISO4217(currencyCode) {
		return 0, fmt.Errorf("%s is not a valid ISO 4217 currency code", currencyCode)
	}

	if minorUnitAmount > 0 {
		factor := GetCurrencyFactor(currencyCode)
		amount = float64(minorUnitAmount) / factor
		amount = math.Round(amount*factor) / factor
	}
	return
}

// IsISO4217 check a currency code is an ISO 4217 currency code
func IsISO4217(code string) bool {
	return iso4217Currencies[strings.ToUpper(strings.TrimSpace(code))]
}

// Format formats a currency amount for display purpose in given locale.
// Empty or invalid locale will fallback to "en".
func Format(amount float64, currencyCode string, locale string) (string, error) {
	if locale == "" {
		locale = "en"
	}

	amt, err := currency.NewAmount(fmt.Sprint(amount), currencyCode)
	if err != nil {
		return "", err
	}
	loc := currency.NewLocale(locale)
	formatter := currency.NewFormatter(loc)
	formatter.MaxDigits = decimalPlaces(currencyCode)
	return formatter.Format(amt), nil
}

// FormatAmount formats a currency amount without the currency symbol & grouping separator
func FormatAmount(amount float64, currencyCode string) string {
	d := decimalPlaces(currencyCode)
	f := fmt.Sprintf("%%.%df", d)
	return fmt.Sprintf(f, amount)
}

// GetCurrencyFactor returns the currency factor
func GetCurrencyFactor(currency string) (factor float64) {
	factor, ok := currencyFactors[strings.ToUpper(currency)]
	if !ok {
		factor = defaultCurrencyFactor
	}
	return
}

// decimalPlaces returns the number of digits after the decimal point
func decimalPlaces(currencyCode string) uint8 {
	return uint8(math.Log10(GetCurrencyFactor(currencyCode)))
}

// Round half up the amount to the currency factor's decimal places
func Round(currencyCode string, amount float64) (roundedAmount float64, err error) {
	if !IsISO4217(currencyCode) {
		return 0, fmt.Errorf("%s is not a valid ISO 4217 currency code", currencyCode)
	}
	switch {
	case amount > 0:
		factor := GetCurrencyFactor(currencyCode)
		minorUnitAmount := math.Round(amount * factor)
		roundedAmount = minorUnitAmount / factor
	case amount < 0:
		err = fmt.Errorf("invalid amount: %f", amount)
	}
	return
}

// Equal returns true if two amounts are equal
func Equal(currencyCode string, amount1 float64, amount2 float64) bool {
	if !IsISO4217(currencyCode) {
		return false
	}
	return Amount(currencyCode, amount1) == Amount(currencyCode, amount2)
}

// Amount returns a amount in a currency, fixed to the currency factor's decimal places
// NOTE: if the currency is invalid, it will return 0 and ignore the error
func Amount(currencyCode string, amount float64) float64 {
	amount, _ = Round(currencyCode, amount)
	return amount
}

// MinorUnitAmount returns a amount in the smallest unit (minor unit) of a currency
// NOTE: if the currency is invalid, it will return 0 and ignore the error
func MinorUnitAmount(currencyCode string, amount float64) uint64 {
	minorUnitAmount, _ := ToMinorUnit(currencyCode, amount)
	return minorUnitAmount
}
