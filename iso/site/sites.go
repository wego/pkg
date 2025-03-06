package site

import "strings"

// ISO 3166-1 alpha-2
const (
	AD = "AD" // Andorra
	AE = "AE" // United Arab Emirates
	AF = "AF" // Afghanistan
	AG = "AG" // Antigua and Barbuda
	AI = "AI" // Anguilla
	AL = "AL" // Albania
	AM = "AM" // Armenia
	AO = "AO" // Angola
	AQ = "AQ" // Antarctica
	AR = "AR" // Argentina
	AS = "AS" // American Samoa
	AT = "AT" // Austria
	AU = "AU" // Australia
	AW = "AW" // Aruba
	AX = "AX" // Åland Islands
	AZ = "AZ" // Azerbaijan
	BA = "BA" // Bosnia and Herzegovina
	BB = "BB" // Barbados
	BD = "BD" // Bangladesh
	BE = "BE" // Belgium
	BF = "BF" // Burkina Faso
	BG = "BG" // Bulgaria
	BH = "BH" // Bahrain
	BI = "BI" // Burundi
	BJ = "BJ" // Benin
	BL = "BL" // Saint Barthélemy
	BM = "BM" // Bermuda
	BN = "BN" // Brunei Darussalam
	BO = "BO" // Bolivia, Plurinational State of
	BQ = "BQ" // Bonaire, Sint Eustatius and Saba
	BR = "BR" // Brazil
	BS = "BS" // Bahamas
	BT = "BT" // Bhutan
	BV = "BV" // Bouvet Island
	BW = "BW" // Botswana
	BY = "BY" // Belarus
	BZ = "BZ" // Belize
	CA = "CA" // Canada
	CC = "CC" // Cocos (Keeling) Islands
	CD = "CD" // Congo, the Democratic Republic of the
	CF = "CF" // Central African Republic
	CG = "CG" // Congo
	CH = "CH" // Switzerland
	CI = "CI" // Côte d'Ivoire
	CK = "CK" // Cook Islands
	CL = "CL" // Chile
	CM = "CM" // Cameroon
	CN = "CN" // China
	CO = "CO" // Colombia
	CR = "CR" // Costa Rica
	CU = "CU" // Cuba
	CV = "CV" // Cape Verde
	CW = "CW" // Curaçao
	CX = "CX" // Christmas Island
	CY = "CY" // Cyprus
	CZ = "CZ" // Czech Republic
	DE = "DE" // Germany
	DJ = "DJ" // Djibouti
	DK = "DK" // Denmark
	DM = "DM" // Dominica
	DO = "DO" // Dominican Republic
	DZ = "DZ" // Algeria
	EC = "EC" // Ecuador
	EE = "EE" // Estonia
	EG = "EG" // Egypt
	EH = "EH" // Western Sahara
	ER = "ER" // Eritrea
	ES = "ES" // Spain
	ET = "ET" // Ethiopia
	FI = "FI" // Finland
	FJ = "FJ" // Fiji
	FK = "FK" // Falkland Islands (Malvinas)
	FM = "FM" // Micronesia, Federated States of Micronesia
	FO = "FO" // Faroe Islands
	FR = "FR" // France
	GA = "GA" // Gabon
	GB = "GB" // United Kingdom
	GD = "GD" // Grenada
	GE = "GE" // Georgia
	GF = "GF" // French Guiana
	GG = "GG" // Guernsey
	GH = "GH" // Ghana
	GI = "GI" // Gibraltar
	GL = "GL" // Greenland
	GM = "GM" // Gambia
	GN = "GN" // Guinea
	GP = "GP" // Guadeloupe
	GQ = "GQ" // Equatorial Guinea
	GR = "GR" // Greece
	GS = "GS" // South Georgia and the South Sandwich Islands
	GT = "GT" // Guatemala
	GU = "GU" // Guam
	GW = "GW" // Guinea-Bissau
	GY = "GY" // Guyana
	HK = "HK" // Hong Kong
	HM = "HM" // Heard Island and McDonald Islands
	HN = "HN" // Honduras
	HR = "HR" // Croatia
	HT = "HT" // Haiti
	HU = "HU" // Hungary
	ID = "ID" // Indonesia
	IE = "IE" // Ireland
	IL = "IL" // The State of Israel
	IM = "IM" // Isle of Man
	IN = "IN" // India
	IO = "IO" // British Indian Ocean Territory
	IQ = "IQ" // Iraq
	IR = "IR" // Iran, Islamic Republic of
	IS = "IS" // Iceland
	IT = "IT" // Italy
	JE = "JE" // Jersey
	JM = "JM" // Jamaica
	JO = "JO" // Jordan
	JP = "JP" // Japan
	KE = "KE" // Kenya
	KG = "KG" // Kyrgyzstan
	KH = "KH" // Cambodia
	KI = "KI" // Kiribati
	KM = "KM" // Comoros
	KN = "KN" // Saint Kitts and Nevis
	KP = "KP" // Korea, Democratic People's Republic of Korea
	KR = "KR" // Korea, Republic of Korea
	KW = "KW" // Kuwait
	KY = "KY" // Cayman Islands
	KZ = "KZ" // Kazakhstan
	LA = "LA" // Lao People's Democratic Republic
	LB = "LB" // Lebanon
	LC = "LC" // Saint Lucia
	LI = "LI" // Liechtenstein
	LK = "LK" // Sri Lanka
	LR = "LR" // Liberia
	LS = "LS" // Lesotho
	LT = "LT" // Lithuania
	LU = "LU" // Luxembourg
	LV = "LV" // Latvia
	LY = "LY" // Libya
	MA = "MA" // Morocco
	MC = "MC" // Monaco
	MD = "MD" // Moldova, Republic of Moldova
	ME = "ME" // Montenegro
	MF = "MF" // Saint Martin (French part)
	MG = "MG" // Madagascar
	MH = "MH" // Marshall Islands
	MK = "MK" // Macedonia, the former Yugoslav Republic of Macedonia
	ML = "ML" // Mali
	MM = "MM" // Myanmar
	MN = "MN" // Mongolia
	MO = "MO" // Macao
	MP = "MP" // Northern Mariana Islands
	MQ = "MQ" // Martinique
	MR = "MR" // Mauritania
	MS = "MS" // Montserrat
	MT = "MT" // Malta
	MU = "MU" // Mauritius
	MV = "MV" // Maldives
	MW = "MW" // Malawi
	MX = "MX" // Mexico
	MY = "MY" // Malaysia
	MZ = "MZ" // Mozambique
	NA = "NA" // Namibia
	NC = "NC" // New Caledonia
	NE = "NE" // Niger
	NF = "NF" // Norfolk Island
	NG = "NG" // Nigeria
	NI = "NI" // Nicaragua
	NL = "NL" // Netherlands
	NO = "NO" // Norway
	NP = "NP" // Nepal
	NR = "NR" // Nauru
	NU = "NU" // Niue
	NZ = "NZ" // New Zealand
	OM = "OM" // Oman
	PA = "PA" // Panama
	PE = "PE" // Peru
	PF = "PF" // French Polynesia
	PG = "PG" // Papua New Guinea
	PH = "PH" // Philippines
	PK = "PK" // Pakistan
	PL = "PL" // Poland
	PM = "PM" // Saint Pierre and Miquelon
	PN = "PN" // Pitcairn
	PR = "PR" // Puerto Rico
	PS = "PS" // The State of Palestine
	PT = "PT" // Portugal
	PW = "PW" // Palau
	PY = "PY" // Paraguay
	QA = "QA" // Qatar
	RE = "RE" // Réunion
	RO = "RO" // Romania
	RS = "RS" // Serbia
	RU = "RU" // Russian Federation
	RW = "RW" // Rwanda
	SA = "SA" // Saudi Arabia
	SB = "SB" // Solomon Islands
	SC = "SC" // Seychelles
	SD = "SD" // Sudan
	SE = "SE" // Sweden
	SG = "SG" // Singapore
	SH = "SH" // Saint Helena, Ascension and Tristan da Cunha
	SI = "SI" // Slovenia
	SJ = "SJ" // Svalbard and Jan Mayen
	SK = "SK" // Slovakia
	SL = "SL" // Sierra Leone
	SM = "SM" // San Marino
	SN = "SN" // Senegal
	SO = "SO" // Somalia
	SR = "SR" // Suriname
	SS = "SS" // South Sudan
	ST = "ST" // Sao Tome and Principe
	SV = "SV" // El Salvador
	SX = "SX" // Sint Maarten (Dutch part)
	SY = "SY" // Syrian Arab Republic
	SZ = "SZ" // Swaziland
	TC = "TC" // Turks and Caicos Islands
	TD = "TD" // Chad
	TF = "TF" // French Southern Territories
	TG = "TG" // Togo
	TH = "TH" // Thailand
	TJ = "TJ" // Tajikistan
	TK = "TK" // Tokelau
	TL = "TL" // Timor-Leste
	TM = "TM" // Turkmenistan
	TN = "TN" // Tunisia
	TO = "TO" // Tonga
	TR = "TR" // Turkey
	TT = "TT" // Trinidad and Tobago
	TV = "TV" // Tuvalu
	TW = "TW" // Taiwan, Province of China
	TZ = "TZ" // Tanzania, United Republic of
	UA = "UA" // Ukraine
	UG = "UG" // Uganda
	UM = "UM" // United States Minor Outlying Islands
	US = "US" // United States
	UY = "UY" // Uruguay
	UZ = "UZ" // Uzbekistan
	VA = "VA" // Holy See (Vatican City State)
	VC = "VC" // Saint Vincent and the Grenadines
	VE = "VE" // Venezuela, Bolivarian Republic of
	VG = "VG" // Virgin Islands, British
	VI = "VI" // Virgin Islands, U.S.
	VN = "VN" // Viet Nam
	VU = "VU" // Vanuatu
	WF = "WF" // Wallis and Futuna
	WS = "WS" // Samoa
	YE = "YE" // Yemen
	YT = "YT" // Mayotte
	ZA = "ZA" // South Africa
	ZM = "ZM" // Zambia
	ZW = "ZW" // Zimbabwe
)

// currencySite maps ISO 3166-1 alpha-2 site codes to ISO 4217 currency codes
var currencySite = map[string]string{
	// A
	"AE": "AED", // United Arab Emirates Dirham
	"AF": "AFN", // Afghan Afghani
	"AL": "ALL", // Albanian Lek
	"AM": "AMD", // Armenian Dram
	"AN": "ANG", // Netherlands Antillean Guilder
	"AO": "AOA", // Angolan Kwanza
	"AR": "ARS", // Argentine Peso
	"AU": "AUD", // Australian Dollar
	"AW": "AWG", // Aruban Florin
	"AZ": "AZN", // Azerbaijani Manat

	// B
	"BA": "BAM", // Bosnia and Herzegovina Convertible Mark
	"BB": "BBD", // Barbadian Dollar
	"BD": "BDT", // Bangladeshi Taka
	"BG": "BGN", // Bulgarian Lev
	"BH": "BHD", // Bahraini Dinar
	"BI": "BIF", // Burundian Franc
	"BM": "BMD", // Bermudian Dollar
	"BN": "BND", // Brunei Dollar
	"BO": "BOB", // Bolivian Boliviano
	"BR": "BRL", // Brazilian Real
	"BS": "BSD", // Bahamian Dollar
	"BT": "BTN", // Bhutanese Ngultrum
	"BW": "BWP", // Botswana Pula
	"BY": "BYN", // Belarusian Ruble
	"BZ": "BZD", // Belize Dollar

	// C
	"CA": "CAD", // Canadian Dollar
	"CD": "CDF", // Congolese Franc
	"CH": "CHF", // Swiss Franc
	"CL": "CLP", // Chilean Peso
	"CN": "CNY", // Chinese Yuan
	"CO": "COP", // Colombian Peso
	"CR": "CRC", // Costa Rican Colón
	"CU": "CUP", // Cuban Peso
	"CV": "CVE", // Cape Verdean Escudo
	"CZ": "CZK", // Czech Republic Koruna

	// D
	"DJ": "DJF", // Djiboutian Franc
	"DK": "DKK", // Danish Krone
	"DO": "DOP", // Dominican Peso
	"DZ": "DZD", // Algerian Dinar

	// E
	"EG": "EGP", // Egyptian Pound
	"ER": "ERN", // Eritrean Nakfa
	"ET": "ETB", // Ethiopian Birr
	"EU": "EUR", // Euro

	// F
	"FJ": "FJD", // Fijian Dollar
	"FK": "FKP", // Falkland Islands Pound

	// G
	"GB": "GBP", // British Pound Sterling
	"GE": "GEL", // Georgian Lari
	"GH": "GHS", // Ghanaian Cedi
	"GI": "GIP", // Gibraltar Pound
	"GM": "GMD", // Gambian Dalasi
	"GN": "GNF", // Guinean Franc
	"GT": "GTQ", // Guatemalan Quetzal
	"GY": "GYD", // Guyanese Dollar

	// H
	"HK": "HKD", // Hong Kong Dollar
	"HN": "HNL", // Honduran Lempira
	"HR": "HRK", // Croatian Kuna
	"HT": "HTG", // Haitian Gourde
	"HU": "HUF", // Hungarian Forint

	// I
	"ID": "IDR", // Indonesian Rupiah
	"IL": "ILS", // Israeli New Shekel
	"IN": "INR", // Indian Rupee
	"IQ": "IQD", // Iraqi Dinar
	"IR": "IRR", // Iranian Rial
	"IS": "ISK", // Icelandic Krona

	// J
	"JM": "JMD", // Jamaican Dollar
	"JO": "JOD", // Jordanian Dinar
	"JP": "JPY", // Japanese Yen

	// K
	"KE": "KES", // Kenyan Shilling
	"KG": "KGS", // Kyrgyzstani Som
	"KH": "KHR", // Cambodian Riel
	"KM": "KMF", // Comoran Franc
	"KP": "KPW", // North Korean Won
	"KR": "KRW", // South Korean Won
	"KW": "KWD", // Kuwaiti Dinar
	"KY": "KYD", // Cayman Islands Dollar
	"KZ": "KZT", // Kazakhstani Tenge

	// L
	"LA": "LAK", // Lao Kip
	"LB": "LBP", // Lebanese Pound
	"LK": "LKR", // Sri Lankan Rupee
	"LR": "LRD", // Liberian Dollar
	"LS": "LSL", // Lesotho Loti
	"LY": "LYD", // Libyan Dinar

	// M
	"MA": "MAD", // Moroccan Dirham
	"MD": "MDL", // Moldovan Leu
	"MG": "MGA", // Malagasy Ariary
	"MK": "MKD", // Macedonian Denar
	"MM": "MMK", // Myanmar Kyat
	"MN": "MNT", // Mongolian Tögrög
	"MO": "MOP", // Macanese Pataca
	"MR": "MRU", // Mauritanian Ouguiya
	"MU": "MUR", // Mauritian Rupee
	"MV": "MVR", // Maldivian Rufiyaa
	"MW": "MWK", // Malawian Kwacha
	"MX": "MXN", // Mexican Peso
	"MY": "MYR", // Malaysian Ringgit
	"MZ": "MZN", // Mozambican Metical

	// N
	"NA": "NAD", // Namibian Dollar
	"NG": "NGN", // Nigerian Naira
	"NI": "NIO", // Nicaraguan Córdoba
	"NO": "NOK", // Norwegian Krone
	"NP": "NPR", // Nepalese Rupee
	"NZ": "NZD", // New Zealand Dollar

	// O
	"OM": "OMR", // Omani Rial

	// P
	"PA": "PAB", // Panamanian Balboa
	"PE": "PEN", // Peruvian Sol
	"PG": "PGK", // Papua New Guinean Kina
	"PH": "PHP", // Philippine Peso
	"PK": "PKR", // Pakistani Rupee
	"PL": "PLN", // Polish Zloty
	"PY": "PYG", // Paraguayan Guarani

	// Q
	"QA": "QAR", // Qatari Riyal

	// R
	"RO": "RON", // Romanian Leu
	"RS": "RSD", // Serbian Dinar
	"RU": "RUB", // Russian Ruble
	"RW": "RWF", // Rwandan Franc

	// S
	"SA": "SAR", // Saudi Riyal
	"SB": "SBD", // Solomon Islands Dollar
	"SC": "SCR", // Seychellois Rupee
	"SD": "SDG", // Sudanese Pound
	"SE": "SEK", // Swedish Krona
	"SG": "SGD", // Singapore Dollar
	"SH": "SHP", // Saint Helena Pound
	"SL": "SLL", // Sierra Leonean Leone
	"SO": "SOS", // Somali Shilling
	"SR": "SRD", // Surinamese Dollar
	"SS": "SSP", // South Sudanese Pound
	"ST": "STN", // São Tomé and Príncipe Dobra
	"SV": "SVC", // Salvadoran Colón
	"SY": "SYP", // Syrian Pound
	"SZ": "SZL", // Swazi Lilangeni

	// T
	"TH": "THB", // Thai Baht
	"TJ": "TJS", // Tajikistani Somoni
	"TM": "TMT", // Turkmenistani Manat
	"TN": "TND", // Tunisian Dinar
	"TO": "TOP", // Tongan Paʻanga
	"TR": "TRY", // Turkish Lira
	"TT": "TTD", // Trinidad and Tobago Dollar
	"TW": "TWD", // New Taiwan Dollar
	"TZ": "TZS", // Tanzanian Shilling

	// U
	"UA": "UAH", // Ukrainian Hryvnia
	"UG": "UGX", // Ugandan Shilling
	"US": "USD", // United States Dollar
	"UY": "UYU", // Uruguayan Peso
	"UZ": "UZS", // Uzbekistani Som

	// V
	"VE": "VES", // Venezuelan Bolívar Soberano
	"VN": "VND", // Vietnamese Dong
	"VU": "VUV", // Vanuatu Vatu

	// W
	"WS": "WST", // Samoan Tala

	// Y
	"YE": "YER", // Yemeni Rial

	// Z
	"ZA": "ZAR", // South African Rand
	"ZM": "ZMW", // Zambian Kwacha
	"ZW": "ZWL", // Zimbabwean Dollar
}

// Currency returns the ISO 4217 currency code for a given ISO 3166-1 alpha-2 site code
func Currency(siteCode string) string {
	siteCode = strings.ToUpper(strings.TrimSpace(siteCode))
	return currencySite[siteCode]
}
