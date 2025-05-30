package site

import "github.com/wego/pkg/currency"

// currencySite maps ISO 3166-1 alpha-2 site codes to ISO 4217 currency codes
var currencySite = map[string]string{
	// A
	AE: currency.AED, // United Arab Emirates Dirham
	AF: currency.AFN, // Afghan Afghani
	AL: currency.ALL, // Albanian Lek
	AM: currency.AMD, // Armenian Dram
	AO: currency.AOA, // Angolan Kwanza
	AR: currency.ARS, // Argentine Peso
	AU: currency.AUD, // Australian Dollar
	AW: currency.AWG, // Aruban Florin
	AZ: currency.AZN, // Azerbaijani Manat

	// B
	BA: currency.BAM, // Bosnia and Herzegovina Convertible Mark
	BB: currency.BBD, // Barbadian Dollar
	BD: currency.BDT, // Bangladeshi Taka
	BE: currency.EUR, // Belgian Euro
	BF: currency.XOF, // West African CFA Franc
	BG: currency.BGN, // Bulgarian Lev
	BH: currency.BHD, // Bahraini Dinar
	BI: currency.BIF, // Burundian Franc
	BJ: currency.XOF, // West African CFA Franc
	BM: currency.BMD, // Bermudian Dollar
	BN: currency.BND, // Brunei Dollar
	BO: currency.BOB, // Bolivian Boliviano
	BQ: currency.USD, // US Dollar (Bonaire, Sint Eustatius and Saba)
	BR: currency.BRL, // Brazilian Real
	BS: currency.BSD, // Bahamian Dollar
	BT: currency.BTN, // Bhutanese Ngultrum
	BV: currency.NOK, // Norwegian Krone
	BW: currency.BWP, // Botswana Pula
	BY: currency.BYN, // Belarusian Ruble
	BZ: currency.BZD, // Belize Dollar

	// C
	CA: currency.CAD, // Canadian Dollar
	CC: currency.AUD, // Australian Dollar
	CD: currency.CDF, // Congolese Franc
	CF: currency.XAF, // Central African CFA Franc
	CG: currency.XAF, // Central African CFA Franc
	CH: currency.CHF, // Swiss Franc
	CI: currency.XOF, // West African CFA Franc
	CK: currency.NZD, // New Zealand Dollar
	CL: currency.CLP, // Chilean Peso
	CM: currency.XAF, // Central African CFA Franc
	CN: currency.CNY, // Chinese Yuan
	CO: currency.COP, // Colombian Peso
	CR: currency.CRC, // Costa Rican Colón
	CU: currency.CUP, // Cuban Peso
	CV: currency.CVE, // Cape Verdean Escudo
	CW: currency.ANG, // Netherlands Antillean Guilder
	CX: currency.AUD, // Australian Dollar
	CY: currency.EUR, // Cypriot Euro
	CZ: currency.CZK, // Czech Republic Koruna

	// D
	DE: currency.EUR, // German Euro
	DJ: currency.DJF, // Djiboutian Franc
	DK: currency.DKK, // Danish Krone
	DM: currency.XCD, // East Caribbean Dollar
	DO: currency.DOP, // Dominican Peso
	DZ: currency.DZD, // Algerian Dinar

	// E
	EC: currency.USD, // US Dollar
	EE: currency.EUR, // Estonian Euro
	EG: currency.EGP, // Egyptian Pound
	EH: currency.MAD, // Moroccan Dirham
	ER: currency.ERN, // Eritrean Nakfa
	ES: currency.EUR, // Spanish Euro
	ET: currency.ETB, // Ethiopian Birr

	// F
	FI: currency.EUR, // Finnish Euro
	FJ: currency.FJD, // Fijian Dollar
	FK: currency.FKP, // Falkland Islands Pound
	FM: currency.USD, // US Dollar
	FO: currency.DKK, // Danish Krone
	FR: currency.EUR, // French Euro

	// G
	GA: currency.XAF, // Central African CFA Franc
	GB: currency.GBP, // British Pound Sterling
	GD: currency.XCD, // East Caribbean Dollar
	GE: currency.GEL, // Georgian Lari
	GF: currency.EUR, // Euro
	GG: currency.GBP, // British Pound Sterling
	GH: currency.GHS, // Ghanaian Cedi
	GI: currency.GIP, // Gibraltar Pound
	GL: currency.DKK, // Danish Krone
	GM: currency.GMD, // Gambian Dalasi
	GN: currency.GNF, // Guinean Franc
	GP: currency.EUR, // Euro
	GQ: currency.XAF, // Central African CFA Franc
	GR: currency.EUR, // Greek Euro
	GS: currency.GBP, // British Pound Sterling
	GT: currency.GTQ, // Guatemalan Quetzal
	GU: currency.USD, // US Dollar
	GW: currency.XOF, // West African CFA Franc
	GY: currency.GYD, // Guyanese Dollar

	// H
	HK: currency.HKD, // Hong Kong Dollar
	HM: currency.AUD, // Australian Dollar
	HN: currency.HNL, // Honduran Lempira
	HR: currency.EUR, // Croatian Euro
	HT: currency.HTG, // Haitian Gourde
	HU: currency.HUF, // Hungarian Forint

	// I
	ID: currency.IDR, // Indonesian Rupiah
	IE: currency.EUR, // Irish Euro
	IL: currency.ILS, // Israeli New Shekel
	IM: currency.GBP, // British Pound Sterling
	IN: currency.INR, // Indian Rupee
	IO: currency.USD, // US Dollar
	IQ: currency.IQD, // Iraqi Dinar
	IR: currency.IRR, // Iranian Rial
	IS: currency.ISK, // Icelandic Krona
	IT: currency.EUR, // Italian Euro

	// J
	JE: currency.GBP, // British Pound Sterling
	JM: currency.JMD, // Jamaican Dollar
	JO: currency.JOD, // Jordanian Dinar
	JP: currency.JPY, // Japanese Yen

	// K
	KE: currency.KES, // Kenyan Shilling
	KG: currency.KGS, // Kyrgyzstani Som
	KH: currency.KHR, // Cambodian Riel
	KI: currency.AUD, // Australian Dollar
	KM: currency.KMF, // Comoran Franc
	KN: currency.XCD, // East Caribbean Dollar
	KP: currency.KPW, // North Korean Won
	KR: currency.KRW, // South Korean Won
	KW: currency.KWD, // Kuwaiti Dinar
	KY: currency.KYD, // Cayman Islands Dollar
	KZ: currency.KZT, // Kazakhstani Tenge

	// L
	LA: currency.LAK, // Lao Kip
	LB: currency.LBP, // Lebanese Pound
	LC: currency.XCD, // East Caribbean Dollar
	LI: currency.CHF, // Swiss Franc
	LK: currency.LKR, // Sri Lankan Rupee
	LR: currency.LRD, // Liberian Dollar
	LS: currency.LSL, // Lesotho Loti
	LT: currency.EUR, // Lithuanian Euro
	LU: currency.EUR, // Luxembourg Euro
	LV: currency.EUR, // Latvian Euro
	LY: currency.LYD, // Libyan Dinar

	// M
	MA: currency.MAD, // Moroccan Dirham
	MC: currency.EUR, // Euro
	MD: currency.MDL, // Moldovan Leu
	ME: currency.EUR, // Montenegrin Euro
	MF: currency.EUR, // Euro
	MG: currency.MGA, // Malagasy Ariary
	MH: currency.USD, // US Dollar
	MK: currency.MKD, // Macedonian Denar
	ML: currency.XOF, // West African CFA Franc
	MM: currency.MMK, // Myanmar Kyat
	MN: currency.MNT, // Mongolian Tögrög
	MO: currency.MOP, // Macanese Pataca
	MP: currency.USD, // US Dollar
	MQ: currency.EUR, // Euro
	MR: currency.MRU, // Mauritanian Ouguiya
	MS: currency.XCD, // East Caribbean Dollar
	MT: currency.EUR, // Maltese Euro
	MU: currency.MUR, // Mauritian Rupee
	MV: currency.MVR, // Maldivian Rufiyaa
	MW: currency.MWK, // Malawian Kwacha
	MX: currency.MXN, // Mexican Peso
	MY: currency.MYR, // Malaysian Ringgit
	MZ: currency.MZN, // Mozambican Metical

	// N
	NA: currency.NAD, // Namibian Dollar
	NC: currency.XPF, // CFP Franc
	NE: currency.XOF, // West African CFA Franc
	NF: currency.AUD, // Australian Dollar
	NG: currency.NGN, // Nigerian Naira
	NI: currency.NIO, // Nicaraguan Córdoba
	NL: currency.EUR, // Dutch Euro
	NO: currency.NOK, // Norwegian Krone
	NP: currency.NPR, // Nepalese Rupee
	NR: currency.AUD, // Australian Dollar
	NU: currency.NZD, // New Zealand Dollar
	NZ: currency.NZD, // New Zealand Dollar

	// O
	OM: currency.OMR, // Omani Rial

	// P
	PA: currency.PAB, // Panamanian Balboa
	PE: currency.PEN, // Peruvian Sol
	PF: currency.XPF, // CFP Franc
	PG: currency.PGK, // Papua New Guinean Kina
	PH: currency.PHP, // Philippine Peso
	PK: currency.PKR, // Pakistani Rupee
	PL: currency.PLN, // Polish Zloty
	PM: currency.EUR, // Euro
	PN: currency.NZD, // New Zealand Dollar
	PR: currency.USD, // US Dollar
	PS: currency.ILS, // Israeli New Shekel
	PT: currency.EUR, // Portuguese Euro
	PW: currency.USD, // US Dollar
	PY: currency.PYG, // Paraguayan Guarani

	// Q
	QA: currency.QAR, // Qatari Riyal

	// R
	RE: currency.EUR, // Euro
	RO: currency.RON, // Romanian Leu
	RS: currency.RSD, // Serbian Dinar
	RU: currency.RUB, // Russian Ruble
	RW: currency.RWF, // Rwandan Franc

	// S
	SA: currency.SAR, // Saudi Riyal
	SB: currency.SBD, // Solomon Islands Dollar
	SC: currency.SCR, // Seychellois Rupee
	SD: currency.SDG, // Sudanese Pound
	SE: currency.SEK, // Swedish Krona
	SG: currency.SGD, // Singapore Dollar
	SH: currency.SHP, // Saint Helena Pound
	SI: currency.EUR, // Slovenian Euro
	SJ: currency.NOK, // Norwegian Krone
	SK: currency.EUR, // Slovak Euro
	SL: currency.SLL, // Sierra Leonean Leone
	SM: currency.EUR, // Euro
	SN: currency.XOF, // West African CFA Franc
	SO: currency.SOS, // Somali Shilling
	SR: currency.SRD, // Surinamese Dollar
	SS: currency.SSP, // South Sudanese Pound
	ST: currency.STN, // São Tomé and Príncipe Dobra
	SV: currency.SVC, // Salvadoran Colón
	SX: currency.ANG, // Netherlands Antillean Guilder
	SY: currency.SYP, // Syrian Pound
	SZ: currency.SZL, // Swazi Lilangeni

	// T
	TC: currency.USD, // US Dollar
	TD: currency.XAF, // Central African CFA Franc
	TF: currency.EUR, // Euro
	TG: currency.XOF, // West African CFA Franc
	TH: currency.THB, // Thai Baht
	TJ: currency.TJS, // Tajikistani Somoni
	TK: currency.NZD, // New Zealand Dollar
	TL: currency.USD, // US Dollar
	TM: currency.TMT, // Turkmenistani Manat
	TN: currency.TND, // Tunisian Dinar
	TO: currency.TOP, // Tongan Paʻanga
	TR: currency.TRY, // Turkish Lira
	TT: currency.TTD, // Trinidad and Tobago Dollar
	TV: currency.AUD, // Australian Dollar
	TW: currency.TWD, // New Taiwan Dollar
	TZ: currency.TZS, // Tanzanian Shilling

	// U
	UA: currency.UAH, // Ukrainian Hryvnia
	UG: currency.UGX, // Ugandan Shilling
	UM: currency.USD, // US Dollar
	US: currency.USD, // US Dollar
	UY: currency.UYU, // Uruguayan Peso
	UZ: currency.UZS, // Uzbekistani Som

	// V
	VA: currency.EUR, // Euro
	VC: currency.XCD, // East Caribbean Dollar
	VE: currency.VES, // Venezuelan Bolívar Soberano
	VG: currency.USD, // US Dollar
	VI: currency.USD, // US Dollar
	VN: currency.VND, // Vietnamese Dong
	VU: currency.VUV, // Vanuatu Vatu

	// W
	WF: currency.XPF, // CFP Franc
	WS: currency.WST, // Samoan Tala

	// Y
	YE: currency.YER, // Yemeni Rial
	YT: currency.EUR, // Euro

	// Z
	ZA: currency.ZAR, // South African Rand
	ZM: currency.ZMW, // Zambian Kwacha
	ZW: currency.ZWL, // Zimbabwean Dollar
}

// Currency returns the ISO 4217 currency code for a given ISO 3166-1 alpha-2 site code
// and a boolean indicating whether the currency was found
func Currency(siteCode string) (currency string, found bool) {
	currency, found = currencySite[siteCode]
	return
}
