package site

import "strings"

// siteZones maps ISO 3166-1 alpha-2 site codes to time zones.
var siteZones = map[string]string{
	// A
	AE: "Asia/Dubai",                     // United Arab Emirates
	AF: "Asia/Kabul",                     // Afghanistan
	AL: "Europe/Tirane",                  // Albania
	AM: "Asia/Yerevan",                   // Armenia
	AO: "Africa/Luanda",                  // Angola
	AR: "America/Argentina/Buenos_Aires", // Argentina
	AU: "Australia/Sydney",               // Australia
	AW: "America/Aruba",                  // Aruba
	AZ: "Asia/Baku",                      // Azerbaijan

	// B
	BA: "Europe/Sarajevo",       // Bosnia and Herzegovina
	BB: "America/Barbados",      // Barbados
	BD: "Asia/Dhaka",            // Bangladesh
	BE: "Europe/Brussels",       // Belgium
	BF: "Africa/Ouagadougou",    // Burkina Faso
	BG: "Europe/Sofia",          // Bulgaria
	BH: "Asia/Bahrain",          // Bahrain
	BI: "Africa/Bujumbura",      // Burundi
	BJ: "Africa/Porto-Novo",     // Benin
	BM: "Atlantic/Bermuda",      // Bermuda
	BN: "Asia/Brunei",           // Brunei Darussalam
	BO: "America/La_Paz",        // Bolivia, Plurinational State of
	BQ: "America/Kralendijk",    // Bonaire, Sint Eustatius and Saba
	BR: "America/Sao_Paulo",     // Brazil
	BS: "America/Nassau",        // Bahamas
	BT: "Asia/Thimphu",          // Bhutan
	BV: "Antarctica/South_Pole", // Bouvet Island
	BW: "Africa/Gaborone",       // Botswana
	BY: "Europe/Minsk",          // Belarus
	BZ: "America/Belize",        // Belize

	// C
	CA: "America/Toronto",     // Canada
	CC: "Australia/Perth",     // Cocos (Keeling) Islands
	CD: "Africa/Kinshasa",     // Congo, the Democratic Republic of the
	CF: "Africa/Bangui",       // Central African Republic
	CG: "Africa/Brazzaville",  // Congo
	CH: "Europe/Zurich",       // Switzerland
	CI: "Africa/Abidjan",      // Côte d'Ivoire
	CK: "Pacific/Rarotonga",   // Cook Islands
	CL: "America/Santiago",    // Chile
	CM: "Africa/Douala",       // Cameroon
	CN: "Asia/Shanghai",       // China
	CO: "America/Bogota",      // Colombia
	CR: "America/Costa_Rica",  // Costa Rica
	CU: "America/Havana",      // Cuba
	CV: "Atlantic/Cape_Verde", // Cape Verde
	CW: "America/Curacao",     // Curaçao
	CX: "Indian/Christmas",    // Christmas Island
	CY: "Asia/Nicosia",        // Cyprus
	CZ: "Europe/Prague",       // Czech Republic

	// D
	DE: "Europe/Berlin",         // Germany
	DJ: "Africa/Djibouti",       // Djibouti
	DK: "Europe/Copenhagen",     // Denmark
	DM: "America/Dominica",      // Dominica
	DO: "America/Santo_Domingo", // Dominican Republic
	DZ: "Africa/Algiers",        // Algeria

	// E
	EC: "America/Guayaquil",  // Ecuador
	EE: "Europe/Tallinn",     // Estonia
	EG: "Africa/Cairo",       // Egypt
	EH: "Africa/El_Aaiun",    // Western Sahara
	ER: "Africa/Asmara",      // Eritrea
	ES: "Europe/Madrid",      // Spain
	ET: "Africa/Addis_Ababa", // Ethiopia

	// F
	FI: "Europe/Helsinki",  // Finland
	FJ: "Pacific/Fiji",     // Fiji
	FK: "Atlantic/Stanley", // Falkland Islands (Malvinas)
	FM: "Pacific/Guam",     // Federated States of Micronesia
	FO: "Atlantic/Faroe",   // Faroe Islands
	FR: "Europe/Paris",     // France

	// G
	GA: "Africa/Libreville",      // Gabon
	GB: "Europe/London",          // United Kingdom
	GD: "America/Grenada",        // Grenada
	GE: "Asia/Tbilisi",           // Georgia
	GF: "America/Cayenne",        // French Guiana
	GG: "Europe/Guernsey",        // Guernsey
	GH: "Africa/Accra",           // Ghana
	GI: "Europe/Gibraltar",       // Gibraltar
	GL: "America/Godthab",        // Greenland
	GM: "Africa/Banjul",          // Gambia
	GN: "Africa/Conakry",         // Guinea
	GP: "America/Guadeloupe",     // Guadeloupe
	GQ: "Africa/Malabo",          // Equatorial Guinea
	GR: "Europe/Athens",          // Greece
	GS: "Atlantic/South_Georgia", // South Georgia and the South Sandwich Islands
	GT: "America/Guatemala",      // Guatemala
	GU: "Pacific/Guam",           // Guam
	GW: "Africa/Bissau",          // Guinea-Bissau
	GY: "America/Guyana",         // Guyana

	// H
	HK: "Asia/Hong_Kong",         // Hong Kong
	HM: "Antarctica/Macquarie",   // Heard Island and McDonald Islands
	HN: "America/Tegucigalpa",    // Honduras
	HR: "Europe/Zagreb",          // Croatia
	HT: "America/Port-au-Prince", // Haiti
	HU: "Europe/Budapest",        // Hungary

	// I
	ID: "Asia/Jakarta",       // Indonesia
	IE: "Europe/Dublin",      // Ireland
	IL: "Asia/Jerusalem",     // Israel
	IM: "Europe/London",      // Isle of Man
	IN: "Asia/Kolkata",       // India
	IO: "Indian/Chagos",      // British Indian Ocean Territory
	IQ: "Asia/Baghdad",       // Iraq
	IR: "Asia/Tehran",        // Iran, Islamic Republic of
	IS: "Atlantic/Reykjavik", // Iceland
	IT: "Europe/Rome",        // Italy

	// J
	JE: "Europe/Jersey",   // Jersey
	JM: "America/Jamaica", // Jamaica
	JO: "Asia/Amman",      // Jordan
	JP: "Asia/Tokyo",      // Japan

	// K
	KE: "Africa/Nairobi",   // Kenya
	KG: "Asia/Bishkek",     // Kyrgyzstan
	KH: "Asia/Phnom_Penh",  // Cambodia
	KI: "Pacific/Tarawa",   // Kiribati
	KM: "Indian/Comoro",    // Comoros
	KN: "America/St_Kitts", // Saint Kitts and Nevis
	KP: "Asia/Pyongyang",   // Korea, Democratic People's Republic of
	KR: "Asia/Seoul",       // Korea, Republic of
	KW: "Asia/Kuwait",      // Kuwait
	KY: "America/Cayman",   // Cayman Islands
	KZ: "Asia/Almaty",      // Kazakhstan

	// L
	LA: "Asia/Vientiane",    // Lao People's Democratic Republic
	LB: "Asia/Beirut",       // Lebanon
	LC: "America/St_Lucia",  // Saint Lucia
	LI: "Europe/Vaduz",      // Liechtenstein
	LK: "Asia/Colombo",      // Sri Lanka
	LR: "Africa/Monrovia",   // Liberia
	LS: "Africa/Maseru",     // Lesotho
	LT: "Europe/Vilnius",    // Lithuania
	LU: "Europe/Luxembourg", // Luxembourg
	LV: "Europe/Riga",       // Latvia
	LY: "Africa/Tripoli",    // Libya

	// M
	MA: "Africa/Casablanca",   // Morocco
	MC: "Europe/Monaco",       // Monaco
	MD: "Europe/Chisinau",     // Moldova, Republic of
	ME: "Europe/Podgorica",    // Montenegro
	MF: "America/Marigot",     // Saint Martin (French part)
	MG: "Indian/Antananarivo", // Madagascar
	MH: "Pacific/Majuro",      // Marshall Islands
	MK: "Europe/Skopje",       // Macedonia, the former Yugoslav Republic of
	ML: "Africa/Bamako",       // Mali
	MM: "Asia/Yangon",         // Myanmar
	MN: "Asia/Ulaanbaatar",    // Mongolia
	MO: "Asia/Macau",          // Macao
	MP: "Pacific/Saipan",      // Northern Mariana Islands
	MQ: "America/Martinique",  // Martinique
	MR: "Africa/Nouakchott",   // Mauritania
	MS: "America/Montserrat",  // Montserrat
	MT: "Europe/Malta",        // Malta
	MU: "Indian/Mauritius",    // Mauritius
	MV: "Indian/Maldives",     // Maldives
	MW: "Africa/Blantyre",     // Malawi
	MX: "America/Mexico_City", // Mexico
	MY: "Asia/Kuala_Lumpur",   // Malaysia
	MZ: "Africa/Maputo",       // Mozambique

	// N
	NA: "Africa/Windhoek",  // Namibia
	NC: "Pacific/Noumea",   // New Caledonia
	NE: "Africa/Niamey",    // Niger
	NF: "Pacific/Norfolk",  // Norfolk Island
	NG: "Africa/Lagos",     // Nigeria
	NI: "America/Managua",  // Nicaragua
	NL: "Europe/Amsterdam", // Netherlands
	NO: "Europe/Oslo",      // Norway
	NP: "Asia/Kathmandu",   // Nepal
	NR: "Pacific/Nauru",    // Nauru
	NU: "Pacific/Niue",     // Niue
	NZ: "Pacific/Auckland", // New Zealand

	// O
	OM: "Asia/Muscat", // Oman

	// P
	PA: "America/Panama",       // Panama
	PE: "America/Lima",         // Peru
	PF: "Pacific/Tahiti",       // French Polynesia
	PG: "Pacific/Port_Moresby", // Papua New Guinea
	PH: "Asia/Manila",          // Philippines
	PK: "Asia/Karachi",         // Pakistan
	PL: "Europe/Warsaw",        // Poland
	PM: "America/Miquelon",     // Saint Pierre and Miquelon
	PN: "Pacific/Pitcairn",     // Pitcairn
	PR: "America/Puerto_Rico",  // Puerto Rico
	PS: "Asia/Hebron",          // Palestinian Territory, Occupied
	PT: "Europe/Lisbon",        // Portugal
	PW: "Pacific/Palau",        // Palau
	PY: "America/Asuncion",     // Paraguay

	// Q
	QA: "Asia/Qatar", // Qatar

	// R
	RE: "Indian/Reunion",   // Réunion
	RO: "Europe/Bucharest", // Romania
	RS: "Europe/Belgrade",  // Serbia
	RU: "Europe/Moscow",    // Russian Federation
	RW: "Africa/Kigali",    // Rwanda

	// S
	SA: "Asia/Riyadh",         // Saudi Arabia
	SB: "Pacific/Guadalcanal", // Solomon Islands
	SC: "Indian/Mahe",         // Seychelles
	SD: "Africa/Khartoum",     // Sudan
	SE: "Europe/Stockholm",    // Sweden
	SG: "Asia/Singapore",      // Singapore
	SH: "Atlantic/St_Helena",  // Saint Helena, Ascension and Tristan da Cunha
	SI: "Europe/Ljubljana",    // Slovenia
	SJ: "Arctic/Longyearbyen", // Svalbard and Jan Mayen
	SK: "Europe/Bratislava",   // Slovakia
	SL: "Africa/Freetown",     // Sierra Leone
	SM: "Europe/San_Marino",   // San Marino
	SN: "Africa/Dakar",        // Senegal
	SO: "Africa/Mogadishu",    // Somalia
	SR: "America/Paramaribo",  // Suriname
	SS: "Africa/Juba",         // South Sudan
	ST: "Africa/Sao_Tome",     // Sao Tome and Principe
	SV: "America/El_Salvador", // El Salvador
	SX: "America/Curacao",     // Sint Maarten (Dutch part)
	SY: "Asia/Damascus",       // Syrian Arab Republic
	SZ: "Africa/Mbabane",      // Eswatini (fmr. "Swaziland")

	// T
	TC: "America/Grand_Turk",    // Turks and Caicos Islands
	TD: "Africa/Ndjamena",       // Chad
	TF: "Indian/Kerguelen",      // French Southern Territories
	TG: "Africa/Lome",           // Togo
	TH: "Asia/Bangkok",          // Thailand
	TJ: "Asia/Dushanbe",         // Tajikistan
	TK: "Pacific/Fakaofo",       // Tokelau
	TL: "Asia/Dili",             // Timor-Leste
	TM: "Asia/Ashgabat",         // Turkmenistan
	TN: "Africa/Tunis",          // Tunisia
	TO: "Pacific/Tongatapu",     // Tonga
	TR: "Europe/Istanbul",       // Turkey
	TT: "America/Port_of_Spain", // Trinidad and Tobago
	TV: "Pacific/Funafuti",      // Tuvalu
	TW: "Asia/Taipei",           // Taiwan, Province of China
	TZ: "Africa/Dar_es_Salaam",  // Tanzania, United Republic of

	// U
	UA: "Europe/Kiev",        // Ukraine
	UG: "Africa/Kampala",     // Uganda
	UM: "Pacific/Wake",       // United States Minor Outlying Islands
	US: "America/New_York",   // United States of America
	UY: "America/Montevideo", // Uruguay
	UZ: "Asia/Tashkent",      // Uzbekistan

	// V
	VA: "Europe/Vatican",     // Holy See (Vatican City State)
	VC: "America/St_Vincent", // Saint Vincent and the Grenadines
	VE: "America/Caracas",    // Venezuela, Bolivarian Republic of
	VG: "America/Tortola",    // Virgin Islands, British
	VI: "America/St_Thomas",  // Virgin Islands, U.S.
	VN: "Asia/Ho_Chi_Minh",   // Vietnam
	VU: "Pacific/Efate",      // Vanuatu

	// W
	WF: "Pacific/Wallis", // Wallis and Futuna
	WS: "Pacific/Apia",   // Samoa

	// Y
	YE: "Asia/Aden",      // Yemen
	YT: "Indian/Mayotte", // Mayotte

	// Z
	ZA: "Africa/Johannesburg", // South Africa
	ZM: "Africa/Lusaka",       // Zambia
	ZW: "Africa/Harare",       // Zimbabwe
}

// TimeZone returns the IANA timezone identifier for a given ISO 3166-1 alpha-2 site code
// and a boolean indicating whether the timezone was found
func TimeZone(siteCode string) (timezone string, found bool) {
	siteCode = strings.ToUpper(siteCode)
	timezone, found = siteZones[siteCode]
	return
}
