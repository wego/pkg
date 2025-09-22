package site

import (
	"fmt"
	"testing"
	"time"
)

func TestTimeZone(t *testing.T) {
	tests := []struct {
		name       string
		siteCode   string
		want       string
		wantOk     bool
		wantOffset string // Expected UTC offset (e.g., "UTC+0", "UTC-5", "UTC+5.5")
	}{
		// A
		{name: "United Arab Emirates", siteCode: AE, want: "Asia/Dubai", wantOk: true, wantOffset: "UTC+4"},
		{name: "Afghanistan", siteCode: AF, want: "Asia/Kabul", wantOk: true, wantOffset: "UTC+4.5"},
		{name: "Albania", siteCode: AL, want: "Europe/Tirane", wantOk: true, wantOffset: "UTC+1/+2"},
		{name: "Armenia", siteCode: AM, want: "Asia/Yerevan", wantOk: true, wantOffset: "UTC+4"},
		{name: "Angola", siteCode: AO, want: "Africa/Luanda", wantOk: true, wantOffset: "UTC+1"},
		{name: "Argentina", siteCode: AR, want: "America/Argentina/Buenos_Aires", wantOk: true, wantOffset: "UTC-3"},
		{name: "Australia", siteCode: AU, want: "Australia/Sydney", wantOk: true, wantOffset: "UTC+10/+11"},
		{name: "Aruba", siteCode: AW, want: "America/Aruba", wantOk: true, wantOffset: "UTC-4"},
		{name: "Azerbaijan", siteCode: AZ, want: "Asia/Baku", wantOk: true, wantOffset: "UTC+4"},

		// B
		{name: "Bosnia and Herzegovina", siteCode: BA, want: "Europe/Sarajevo", wantOk: true, wantOffset: "UTC+1/+2"},
		{name: "Barbados", siteCode: BB, want: "America/Barbados", wantOk: true, wantOffset: "UTC-4"},
		{name: "Bangladesh", siteCode: BD, want: "Asia/Dhaka", wantOk: true, wantOffset: "UTC+6"},
		{name: "Belgium", siteCode: BE, want: "Europe/Brussels", wantOk: true, wantOffset: "UTC+1/+2"},
		{name: "Burkina Faso", siteCode: BF, want: "Africa/Ouagadougou", wantOk: true, wantOffset: "UTC+0"},
		{name: "Bulgaria", siteCode: BG, want: "Europe/Sofia", wantOk: true, wantOffset: "UTC+2/+3"},
		{name: "Bahrain", siteCode: BH, want: "Asia/Bahrain", wantOk: true, wantOffset: "UTC+3"},
		{name: "Burundi", siteCode: BI, want: "Africa/Bujumbura", wantOk: true, wantOffset: "UTC+2"},
		{name: "Benin", siteCode: BJ, want: "Africa/Porto-Novo", wantOk: true, wantOffset: "UTC+1"},
		{name: "Bermuda", siteCode: BM, want: "Atlantic/Bermuda", wantOk: true, wantOffset: "UTC-4/-3"},
		{name: "Brunei Darussalam", siteCode: BN, want: "Asia/Brunei", wantOk: true, wantOffset: "UTC+8"},
		{name: "Bolivia", siteCode: BO, want: "America/La_Paz", wantOk: true, wantOffset: "UTC-4"},
		{name: "Bonaire, Sint Eustatius and Saba", siteCode: BQ, want: "America/Kralendijk", wantOk: true, wantOffset: "UTC-4"},
		{name: "Brazil", siteCode: BR, want: "America/Sao_Paulo", wantOk: true, wantOffset: "UTC-3"},
		{name: "Bahamas", siteCode: BS, want: "America/Nassau", wantOk: true, wantOffset: "UTC-5/-4"},
		{name: "Bhutan", siteCode: BT, want: "Asia/Thimphu", wantOk: true, wantOffset: "UTC+6"},
		{name: "Bouvet Island", siteCode: BV, want: "Antarctica/South_Pole", wantOk: true, wantOffset: "UTC+12/+13"},
		{name: "Botswana", siteCode: BW, want: "Africa/Gaborone", wantOk: true, wantOffset: "UTC+2"},
		{name: "Belarus", siteCode: BY, want: "Europe/Minsk", wantOk: true, wantOffset: "UTC+3"},
		{name: "Belize", siteCode: BZ, want: "America/Belize", wantOk: true, wantOffset: "UTC-6"},

		// C
		{name: "Canada", siteCode: CA, want: "America/Toronto", wantOk: true, wantOffset: "UTC-5/-4"},
		{name: "Cocos (Keeling) Islands", siteCode: CC, want: "Australia/Perth", wantOk: true, wantOffset: "UTC+8"},
		{name: "Congo, the Democratic Republic of the", siteCode: CD, want: "Africa/Kinshasa", wantOk: true, wantOffset: "UTC+1"},
		{name: "Central African Republic", siteCode: CF, want: "Africa/Bangui", wantOk: true, wantOffset: "UTC+1"},
		{name: "Congo", siteCode: CG, want: "Africa/Brazzaville", wantOk: true, wantOffset: "UTC+1"},
		{name: "Switzerland", siteCode: CH, want: "Europe/Zurich", wantOk: true, wantOffset: "UTC+1/+2"},
		{name: "Côte d'Ivoire", siteCode: CI, want: "Africa/Abidjan", wantOk: true, wantOffset: "UTC+0"},
		{name: "Cook Islands", siteCode: CK, want: "Pacific/Rarotonga", wantOk: true, wantOffset: "UTC-10"},
		{name: "Chile", siteCode: CL, want: "America/Santiago", wantOk: true, wantOffset: "UTC-4/-3"},
		{name: "Cameroon", siteCode: CM, want: "Africa/Douala", wantOk: true, wantOffset: "UTC+1"},
		{name: "China", siteCode: CN, want: "Asia/Shanghai", wantOk: true, wantOffset: "UTC+8"},
		{name: "Colombia", siteCode: CO, want: "America/Bogota", wantOk: true, wantOffset: "UTC-5"},
		{name: "Costa Rica", siteCode: CR, want: "America/Costa_Rica", wantOk: true, wantOffset: "UTC-6"},
		{name: "Cuba", siteCode: CU, want: "America/Havana", wantOk: true, wantOffset: "UTC-5/-4"},
		{name: "Cape Verde", siteCode: CV, want: "Atlantic/Cape_Verde", wantOk: true, wantOffset: "UTC-1"},
		{name: "Curaçao", siteCode: CW, want: "America/Curacao", wantOk: true, wantOffset: "UTC-4"},
		{name: "Christmas Island", siteCode: CX, want: "Indian/Christmas", wantOk: true, wantOffset: "UTC+7"},
		{name: "Cyprus", siteCode: CY, want: "Asia/Nicosia", wantOk: true, wantOffset: "UTC+2/+3"},
		{name: "Czech Republic", siteCode: CZ, want: "Europe/Prague", wantOk: true, wantOffset: "UTC+1/+2"},

		// D
		{name: "Germany", siteCode: DE, want: "Europe/Berlin", wantOk: true, wantOffset: "UTC+1/+2"},
		{name: "Djibouti", siteCode: DJ, want: "Africa/Djibouti", wantOk: true, wantOffset: "UTC+3"},
		{name: "Denmark", siteCode: DK, want: "Europe/Copenhagen", wantOk: true, wantOffset: "UTC+1/+2"},
		{name: "Dominica", siteCode: DM, want: "America/Dominica", wantOk: true, wantOffset: "UTC-4"},
		{name: "Dominican Republic", siteCode: DO, want: "America/Santo_Domingo", wantOk: true, wantOffset: "UTC-4"},
		{name: "Algeria", siteCode: DZ, want: "Africa/Algiers", wantOk: true, wantOffset: "UTC+1"},

		// E
		{name: "Ecuador", siteCode: EC, want: "America/Guayaquil", wantOk: true, wantOffset: "UTC-5"},
		{name: "Estonia", siteCode: EE, want: "Europe/Tallinn", wantOk: true, wantOffset: "UTC+2/+3"},
		{name: "Egypt", siteCode: EG, want: "Africa/Cairo", wantOk: true, wantOffset: "UTC+2/+3"},
		{name: "Western Sahara", siteCode: EH, want: "Africa/El_Aaiun", wantOk: true, wantOffset: "UTC+1"},
		{name: "Eritrea", siteCode: ER, want: "Africa/Asmara", wantOk: true, wantOffset: "UTC+3"},
		{name: "Spain", siteCode: ES, want: "Europe/Madrid", wantOk: true, wantOffset: "UTC+1/+2"},
		{name: "Ethiopia", siteCode: ET, want: "Africa/Addis_Ababa", wantOk: true, wantOffset: "UTC+3"},

		// F
		{name: "Finland", siteCode: FI, want: "Europe/Helsinki", wantOk: true, wantOffset: "UTC+2/+3"},
		{name: "Fiji", siteCode: FJ, want: "Pacific/Fiji", wantOk: true, wantOffset: "UTC+12"},
		{name: "Falkland Islands", siteCode: FK, want: "Atlantic/Stanley", wantOk: true, wantOffset: "UTC-3"},
		{name: "Federated States of Micronesia", siteCode: FM, want: "Pacific/Guam", wantOk: true, wantOffset: "UTC+10"},
		{name: "Faroe Islands", siteCode: FO, want: "Atlantic/Faroe", wantOk: true, wantOffset: "UTC+0/+1"},
		{name: "France", siteCode: FR, want: "Europe/Paris", wantOk: true, wantOffset: "UTC+1/+2"},

		// G
		{name: "Gabon", siteCode: GA, want: "Africa/Libreville", wantOk: true, wantOffset: "UTC+1"},
		{name: "United Kingdom", siteCode: GB, want: "Europe/London", wantOk: true, wantOffset: "UTC+0/+1"},
		{name: "Grenada", siteCode: GD, want: "America/Grenada", wantOk: true, wantOffset: "UTC-4"},
		{name: "Georgia", siteCode: GE, want: "Asia/Tbilisi", wantOk: true, wantOffset: "UTC+4"},
		{name: "French Guiana", siteCode: GF, want: "America/Cayenne", wantOk: true, wantOffset: "UTC-3"},
		{name: "Guernsey", siteCode: GG, want: "Europe/Guernsey", wantOk: true, wantOffset: "UTC+0/+1"},
		{name: "Ghana", siteCode: GH, want: "Africa/Accra", wantOk: true, wantOffset: "UTC+0"},
		{name: "Gibraltar", siteCode: GI, want: "Europe/Gibraltar", wantOk: true, wantOffset: "UTC+1/+2"},
		{name: "Greenland", siteCode: GL, want: "America/Godthab", wantOk: true, wantOffset: "UTC-2/-1"},
		{name: "Gambia", siteCode: GM, want: "Africa/Banjul", wantOk: true, wantOffset: "UTC+0"},
		{name: "Guinea", siteCode: GN, want: "Africa/Conakry", wantOk: true, wantOffset: "UTC+0"},
		{name: "Guadeloupe", siteCode: GP, want: "America/Guadeloupe", wantOk: true, wantOffset: "UTC-4"},
		{name: "Equatorial Guinea", siteCode: GQ, want: "Africa/Malabo", wantOk: true, wantOffset: "UTC+1"},
		{name: "Greece", siteCode: GR, want: "Europe/Athens", wantOk: true, wantOffset: "UTC+2/+3"},
		{name: "South Georgia and the South Sandwich Islands", siteCode: GS, want: "Atlantic/South_Georgia", wantOk: true, wantOffset: "UTC-2"},
		{name: "Guatemala", siteCode: GT, want: "America/Guatemala", wantOk: true, wantOffset: "UTC-6"},
		{name: "Guam", siteCode: GU, want: "Pacific/Guam", wantOk: true, wantOffset: "UTC+10"},
		{name: "Guinea-Bissau", siteCode: GW, want: "Africa/Bissau", wantOk: true, wantOffset: "UTC+0"},
		{name: "Guyana", siteCode: GY, want: "America/Guyana", wantOk: true, wantOffset: "UTC-4"},

		// H
		{name: "Hong Kong", siteCode: HK, want: "Asia/Hong_Kong", wantOk: true, wantOffset: "UTC+8"},
		{name: "Heard Island and McDonald Islands", siteCode: HM, want: "Antarctica/Macquarie", wantOk: true, wantOffset: "UTC+10/+11"},
		{name: "Honduras", siteCode: HN, want: "America/Tegucigalpa", wantOk: true, wantOffset: "UTC-6"},
		{name: "Croatia", siteCode: HR, want: "Europe/Zagreb", wantOk: true, wantOffset: "UTC+1/+2"},
		{name: "Haiti", siteCode: HT, want: "America/Port-au-Prince", wantOk: true, wantOffset: "UTC-5/-4"},
		{name: "Hungary", siteCode: HU, want: "Europe/Budapest", wantOk: true, wantOffset: "UTC+1/+2"},

		// I
		{name: "Indonesia", siteCode: ID, want: "Asia/Jakarta", wantOk: true, wantOffset: "UTC+7"},
		{name: "Ireland", siteCode: IE, want: "Europe/Dublin", wantOk: true, wantOffset: "UTC+0/+1"},
		{name: "Israel", siteCode: IL, want: "Asia/Jerusalem", wantOk: true, wantOffset: "UTC+2/+3"},
		{name: "Isle of Man", siteCode: IM, want: "Europe/London", wantOk: true, wantOffset: "UTC+0/+1"},
		{name: "India", siteCode: IN, want: "Asia/Kolkata", wantOk: true, wantOffset: "UTC+5.5"},
		{name: "British Indian Ocean Territory", siteCode: IO, want: "Indian/Chagos", wantOk: true, wantOffset: "UTC+6"},
		{name: "Iraq", siteCode: IQ, want: "Asia/Baghdad", wantOk: true, wantOffset: "UTC+3"},
		{name: "Iran", siteCode: IR, want: "Asia/Tehran", wantOk: true, wantOffset: "UTC+3.5"},
		{name: "Iceland", siteCode: IS, want: "Atlantic/Reykjavik", wantOk: true, wantOffset: "UTC+0"},
		{name: "Italy", siteCode: IT, want: "Europe/Rome", wantOk: true, wantOffset: "UTC+1/+2"},

		// J
		{name: "Jersey", siteCode: JE, want: "Europe/Jersey", wantOk: true, wantOffset: "UTC+0/+1"},
		{name: "Jamaica", siteCode: JM, want: "America/Jamaica", wantOk: true, wantOffset: "UTC-5"},
		{name: "Jordan", siteCode: JO, want: "Asia/Amman", wantOk: true, wantOffset: "UTC+3"},
		{name: "Japan", siteCode: JP, want: "Asia/Tokyo", wantOk: true, wantOffset: "UTC+9"},

		// K
		{name: "Kenya", siteCode: KE, want: "Africa/Nairobi", wantOk: true, wantOffset: "UTC+3"},
		{name: "Kyrgyzstan", siteCode: KG, want: "Asia/Bishkek", wantOk: true, wantOffset: "UTC+6"},
		{name: "Cambodia", siteCode: KH, want: "Asia/Phnom_Penh", wantOk: true, wantOffset: "UTC+7"},
		{name: "Kiribati", siteCode: KI, want: "Pacific/Tarawa", wantOk: true, wantOffset: "UTC+12"},
		{name: "Comoros", siteCode: KM, want: "Indian/Comoro", wantOk: true, wantOffset: "UTC+3"},
		{name: "Saint Kitts and Nevis", siteCode: KN, want: "America/St_Kitts", wantOk: true, wantOffset: "UTC-4"},
		{name: "North Korea", siteCode: KP, want: "Asia/Pyongyang", wantOk: true, wantOffset: "UTC+9"},
		{name: "South Korea", siteCode: KR, want: "Asia/Seoul", wantOk: true, wantOffset: "UTC+9"},
		{name: "Kuwait", siteCode: KW, want: "Asia/Kuwait", wantOk: true, wantOffset: "UTC+3"},
		{name: "Cayman Islands", siteCode: KY, want: "America/Cayman", wantOk: true, wantOffset: "UTC-5"},
		{name: "Kazakhstan", siteCode: KZ, want: "Asia/Almaty", wantOk: true, wantOffset: "UTC+5/+6"},

		// L
		{name: "Laos", siteCode: LA, want: "Asia/Vientiane", wantOk: true, wantOffset: "UTC+7"},
		{name: "Lebanon", siteCode: LB, want: "Asia/Beirut", wantOk: true, wantOffset: "UTC+2/+3"},
		{name: "Saint Lucia", siteCode: LC, want: "America/St_Lucia", wantOk: true, wantOffset: "UTC-4"},
		{name: "Liechtenstein", siteCode: LI, want: "Europe/Vaduz", wantOk: true, wantOffset: "UTC+1/+2"},
		{name: "Sri Lanka", siteCode: LK, want: "Asia/Colombo", wantOk: true, wantOffset: "UTC+5.5"},
		{name: "Liberia", siteCode: LR, want: "Africa/Monrovia", wantOk: true, wantOffset: "UTC+0"},
		{name: "Lesotho", siteCode: LS, want: "Africa/Maseru", wantOk: true, wantOffset: "UTC+2"},
		{name: "Lithuania", siteCode: LT, want: "Europe/Vilnius", wantOk: true, wantOffset: "UTC+2/+3"},
		{name: "Luxembourg", siteCode: LU, want: "Europe/Luxembourg", wantOk: true, wantOffset: "UTC+1/+2"},
		{name: "Latvia", siteCode: LV, want: "Europe/Riga", wantOk: true, wantOffset: "UTC+2/+3"},
		{name: "Libya", siteCode: LY, want: "Africa/Tripoli", wantOk: true, wantOffset: "UTC+2"},

		// M
		{name: "Morocco", siteCode: MA, want: "Africa/Casablanca", wantOk: true, wantOffset: "UTC+1"},
		{name: "Monaco", siteCode: MC, want: "Europe/Monaco", wantOk: true, wantOffset: "UTC+1/+2"},
		{name: "Moldova", siteCode: MD, want: "Europe/Chisinau", wantOk: true, wantOffset: "UTC+2/+3"},
		{name: "Montenegro", siteCode: ME, want: "Europe/Podgorica", wantOk: true, wantOffset: "UTC+1/+2"},
		{name: "Saint Martin", siteCode: MF, want: "America/Marigot", wantOk: true, wantOffset: "UTC-4"},
		{name: "Madagascar", siteCode: MG, want: "Indian/Antananarivo", wantOk: true, wantOffset: "UTC+3"},
		{name: "Marshall Islands", siteCode: MH, want: "Pacific/Majuro", wantOk: true, wantOffset: "UTC+12"},
		{name: "Macedonia", siteCode: MK, want: "Europe/Skopje", wantOk: true, wantOffset: "UTC+1/+2"},
		{name: "Mali", siteCode: ML, want: "Africa/Bamako", wantOk: true, wantOffset: "UTC+0"},
		{name: "Myanmar", siteCode: MM, want: "Asia/Yangon", wantOk: true, wantOffset: "UTC+6.5"},
		{name: "Mongolia", siteCode: MN, want: "Asia/Ulaanbaatar", wantOk: true, wantOffset: "UTC+8"},
		{name: "Macao", siteCode: MO, want: "Asia/Macau", wantOk: true, wantOffset: "UTC+8"},
		{name: "Northern Mariana Islands", siteCode: MP, want: "Pacific/Saipan", wantOk: true, wantOffset: "UTC+10"},
		{name: "Martinique", siteCode: MQ, want: "America/Martinique", wantOk: true, wantOffset: "UTC-4"},
		{name: "Mauritania", siteCode: MR, want: "Africa/Nouakchott", wantOk: true, wantOffset: "UTC+0"},
		{name: "Montserrat", siteCode: MS, want: "America/Montserrat", wantOk: true, wantOffset: "UTC-4"},
		{name: "Malta", siteCode: MT, want: "Europe/Malta", wantOk: true, wantOffset: "UTC+1/+2"},
		{name: "Mauritius", siteCode: MU, want: "Indian/Mauritius", wantOk: true, wantOffset: "UTC+4"},
		{name: "Maldives", siteCode: MV, want: "Indian/Maldives", wantOk: true, wantOffset: "UTC+5"},
		{name: "Malawi", siteCode: MW, want: "Africa/Blantyre", wantOk: true, wantOffset: "UTC+2"},
		{name: "Mexico", siteCode: MX, want: "America/Mexico_City", wantOk: true, wantOffset: "UTC-6"},
		{name: "Malaysia", siteCode: MY, want: "Asia/Kuala_Lumpur", wantOk: true, wantOffset: "UTC+8"},
		{name: "Mozambique", siteCode: MZ, want: "Africa/Maputo", wantOk: true, wantOffset: "UTC+2"},

		// N
		{name: "Namibia", siteCode: NA, want: "Africa/Windhoek", wantOk: true, wantOffset: "UTC+2"},
		{name: "New Caledonia", siteCode: NC, want: "Pacific/Noumea", wantOk: true, wantOffset: "UTC+11"},
		{name: "Niger", siteCode: NE, want: "Africa/Niamey", wantOk: true, wantOffset: "UTC+1"},
		{name: "Norfolk Island", siteCode: NF, want: "Pacific/Norfolk", wantOk: true, wantOffset: "UTC+11/+12"},
		{name: "Nigeria", siteCode: NG, want: "Africa/Lagos", wantOk: true, wantOffset: "UTC+1"},
		{name: "Nicaragua", siteCode: NI, want: "America/Managua", wantOk: true, wantOffset: "UTC-6"},
		{name: "Netherlands", siteCode: NL, want: "Europe/Amsterdam", wantOk: true, wantOffset: "UTC+1/+2"},
		{name: "Norway", siteCode: NO, want: "Europe/Oslo", wantOk: true, wantOffset: "UTC+1/+2"},
		{name: "Nepal", siteCode: NP, want: "Asia/Kathmandu", wantOk: true, wantOffset: "UTC+5.75"},
		{name: "Nauru", siteCode: NR, want: "Pacific/Nauru", wantOk: true, wantOffset: "UTC+12"},
		{name: "Niue", siteCode: NU, want: "Pacific/Niue", wantOk: true, wantOffset: "UTC-11"},
		{name: "New Zealand", siteCode: NZ, want: "Pacific/Auckland", wantOk: true, wantOffset: "UTC+12/+13"},

		// O
		{name: "Oman", siteCode: OM, want: "Asia/Muscat", wantOk: true, wantOffset: "UTC+4"},

		// P
		{name: "Panama", siteCode: PA, want: "America/Panama", wantOk: true, wantOffset: "UTC-5"},
		{name: "Peru", siteCode: PE, want: "America/Lima", wantOk: true, wantOffset: "UTC-5"},
		{name: "French Polynesia", siteCode: PF, want: "Pacific/Tahiti", wantOk: true, wantOffset: "UTC-10"},
		{name: "Papua New Guinea", siteCode: PG, want: "Pacific/Port_Moresby", wantOk: true, wantOffset: "UTC+10"},
		{name: "Philippines", siteCode: PH, want: "Asia/Manila", wantOk: true, wantOffset: "UTC+8"},
		{name: "Pakistan", siteCode: PK, want: "Asia/Karachi", wantOk: true, wantOffset: "UTC+5"},
		{name: "Poland", siteCode: PL, want: "Europe/Warsaw", wantOk: true, wantOffset: "UTC+1/+2"},
		{name: "Saint Pierre and Miquelon", siteCode: PM, want: "America/Miquelon", wantOk: true, wantOffset: "UTC-3/-2"},
		{name: "Pitcairn", siteCode: PN, want: "Pacific/Pitcairn", wantOk: true, wantOffset: "UTC-8"},
		{name: "Puerto Rico", siteCode: PR, want: "America/Puerto_Rico", wantOk: true, wantOffset: "UTC-4"},
		{name: "Palestinian Territory", siteCode: PS, want: "Asia/Hebron", wantOk: true, wantOffset: "UTC+2/+3"},
		{name: "Portugal", siteCode: PT, want: "Europe/Lisbon", wantOk: true, wantOffset: "UTC+0/+1"},
		{name: "Palau", siteCode: PW, want: "Pacific/Palau", wantOk: true, wantOffset: "UTC+9"},
		{name: "Paraguay", siteCode: PY, want: "America/Asuncion", wantOk: true, wantOffset: "UTC-4/-3"},

		// Q
		{name: "Qatar", siteCode: QA, want: "Asia/Qatar", wantOk: true, wantOffset: "UTC+3"},

		// R
		{name: "Réunion", siteCode: RE, want: "Indian/Reunion", wantOk: true, wantOffset: "UTC+4"},
		{name: "Romania", siteCode: RO, want: "Europe/Bucharest", wantOk: true, wantOffset: "UTC+2/+3"},
		{name: "Serbia", siteCode: RS, want: "Europe/Belgrade", wantOk: true, wantOffset: "UTC+1/+2"},
		{name: "Russian Federation", siteCode: RU, want: "Europe/Moscow", wantOk: true, wantOffset: "UTC+3"},
		{name: "Rwanda", siteCode: RW, want: "Africa/Kigali", wantOk: true, wantOffset: "UTC+2"},

		// S
		{name: "Saudi Arabia", siteCode: SA, want: "Asia/Riyadh", wantOk: true, wantOffset: "UTC+3"},
		{name: "Solomon Islands", siteCode: SB, want: "Pacific/Guadalcanal", wantOk: true, wantOffset: "UTC+11"},
		{name: "Seychelles", siteCode: SC, want: "Indian/Mahe", wantOk: true, wantOffset: "UTC+4"},
		{name: "Sudan", siteCode: SD, want: "Africa/Khartoum", wantOk: true, wantOffset: "UTC+2"},
		{name: "Sweden", siteCode: SE, want: "Europe/Stockholm", wantOk: true, wantOffset: "UTC+1/+2"},
		{name: "Singapore", siteCode: SG, want: "Asia/Singapore", wantOk: true, wantOffset: "UTC+8"},
		{name: "Saint Helena", siteCode: SH, want: "Atlantic/St_Helena", wantOk: true, wantOffset: "UTC+0"},
		{name: "Slovenia", siteCode: SI, want: "Europe/Ljubljana", wantOk: true, wantOffset: "UTC+1/+2"},
		{name: "Svalbard and Jan Mayen", siteCode: SJ, want: "Arctic/Longyearbyen", wantOk: true, wantOffset: "UTC+1/+2"},
		{name: "Slovakia", siteCode: SK, want: "Europe/Bratislava", wantOk: true, wantOffset: "UTC+1/+2"},
		{name: "Sierra Leone", siteCode: SL, want: "Africa/Freetown", wantOk: true, wantOffset: "UTC+0"},
		{name: "San Marino", siteCode: SM, want: "Europe/San_Marino", wantOk: true, wantOffset: "UTC+1/+2"},
		{name: "Senegal", siteCode: SN, want: "Africa/Dakar", wantOk: true, wantOffset: "UTC+0"},
		{name: "Somalia", siteCode: SO, want: "Africa/Mogadishu", wantOk: true, wantOffset: "UTC+3"},
		{name: "Suriname", siteCode: SR, want: "America/Paramaribo", wantOk: true, wantOffset: "UTC-3"},
		{name: "South Sudan", siteCode: SS, want: "Africa/Juba", wantOk: true, wantOffset: "UTC+2"},
		{name: "Sao Tome and Principe", siteCode: ST, want: "Africa/Sao_Tome", wantOk: true, wantOffset: "UTC+0"},
		{name: "El Salvador", siteCode: SV, want: "America/El_Salvador", wantOk: true, wantOffset: "UTC-6"},
		{name: "Sint Maarten", siteCode: SX, want: "America/Curacao", wantOk: true, wantOffset: "UTC-4"},
		{name: "Syrian Arab Republic", siteCode: SY, want: "Asia/Damascus", wantOk: true, wantOffset: "UTC+3"},
		{name: "Eswatini", siteCode: SZ, want: "Africa/Mbabane", wantOk: true, wantOffset: "UTC+2"},

		// T
		{name: "Turks and Caicos Islands", siteCode: TC, want: "America/Grand_Turk", wantOk: true, wantOffset: "UTC-5/-4"},
		{name: "Chad", siteCode: TD, want: "Africa/Ndjamena", wantOk: true, wantOffset: "UTC+1"},
		{name: "French Southern Territories", siteCode: TF, want: "Indian/Kerguelen", wantOk: true, wantOffset: "UTC+5"},
		{name: "Togo", siteCode: TG, want: "Africa/Lome", wantOk: true, wantOffset: "UTC+0"},
		{name: "Thailand", siteCode: TH, want: "Asia/Bangkok", wantOk: true, wantOffset: "UTC+7"},
		{name: "Tajikistan", siteCode: TJ, want: "Asia/Dushanbe", wantOk: true, wantOffset: "UTC+5"},
		{name: "Tokelau", siteCode: TK, want: "Pacific/Fakaofo", wantOk: true, wantOffset: "UTC+13"},
		{name: "Timor-Leste", siteCode: TL, want: "Asia/Dili", wantOk: true, wantOffset: "UTC+9"},
		{name: "Turkmenistan", siteCode: TM, want: "Asia/Ashgabat", wantOk: true, wantOffset: "UTC+5"},
		{name: "Tunisia", siteCode: TN, want: "Africa/Tunis", wantOk: true, wantOffset: "UTC+1"},
		{name: "Tonga", siteCode: TO, want: "Pacific/Tongatapu", wantOk: true, wantOffset: "UTC+13"},
		{name: "Turkey", siteCode: TR, want: "Europe/Istanbul", wantOk: true, wantOffset: "UTC+3"},
		{name: "Trinidad and Tobago", siteCode: TT, want: "America/Port_of_Spain", wantOk: true, wantOffset: "UTC-4"},
		{name: "Tuvalu", siteCode: TV, want: "Pacific/Funafuti", wantOk: true, wantOffset: "UTC+12"},
		{name: "Taiwan", siteCode: TW, want: "Asia/Taipei", wantOk: true, wantOffset: "UTC+8"},
		{name: "Tanzania", siteCode: TZ, want: "Africa/Dar_es_Salaam", wantOk: true, wantOffset: "UTC+3"},

		// U
		{name: "Ukraine", siteCode: UA, want: "Europe/Kiev", wantOk: true, wantOffset: "UTC+2/+3"},
		{name: "Uganda", siteCode: UG, want: "Africa/Kampala", wantOk: true, wantOffset: "UTC+3"},
		{name: "United States Minor Outlying Islands", siteCode: UM, want: "Pacific/Wake", wantOk: true, wantOffset: "UTC+12"},
		{name: "United States", siteCode: US, want: "America/New_York", wantOk: true, wantOffset: "UTC-5/-4"},
		{name: "Uruguay", siteCode: UY, want: "America/Montevideo", wantOk: true, wantOffset: "UTC-3"},
		{name: "Uzbekistan", siteCode: UZ, want: "Asia/Tashkent", wantOk: true, wantOffset: "UTC+5"},

		// V
		{name: "Vatican City", siteCode: VA, want: "Europe/Vatican", wantOk: true, wantOffset: "UTC+1/+2"},
		{name: "Saint Vincent and the Grenadines", siteCode: VC, want: "America/St_Vincent", wantOk: true, wantOffset: "UTC-4"},
		{name: "Venezuela", siteCode: VE, want: "America/Caracas", wantOk: true, wantOffset: "UTC-4"},
		{name: "Virgin Islands, British", siteCode: VG, want: "America/Tortola", wantOk: true, wantOffset: "UTC-4"},
		{name: "Virgin Islands, U.S.", siteCode: VI, want: "America/St_Thomas", wantOk: true, wantOffset: "UTC-4"},
		{name: "Vietnam", siteCode: VN, want: "Asia/Ho_Chi_Minh", wantOk: true, wantOffset: "UTC+7"},
		{name: "Vanuatu", siteCode: VU, want: "Pacific/Efate", wantOk: true, wantOffset: "UTC+11"},

		// W
		{name: "Wallis and Futuna", siteCode: WF, want: "Pacific/Wallis", wantOk: true, wantOffset: "UTC+12"},
		{name: "Samoa", siteCode: WS, want: "Pacific/Apia", wantOk: true, wantOffset: "UTC+13"},

		// Y
		{name: "Yemen", siteCode: YE, want: "Asia/Aden", wantOk: true, wantOffset: "UTC+3"},
		{name: "Mayotte", siteCode: YT, want: "Indian/Mayotte", wantOk: true, wantOffset: "UTC+3"},

		// Z
		{name: "South Africa", siteCode: ZA, want: "Africa/Johannesburg", wantOk: true, wantOffset: "UTC+2"},
		{name: "Zambia", siteCode: ZM, want: "Africa/Lusaka", wantOk: true, wantOffset: "UTC+2"},
		{name: "Zimbabwe", siteCode: ZW, want: "Africa/Harare", wantOk: true, wantOffset: "UTC+2"},

		// Test cases for invalid site codes
		{name: "Invalid site code", siteCode: "XX", want: "", wantOk: false, wantOffset: ""},
		{name: "Empty site code", siteCode: "", want: "", wantOk: false, wantOffset: ""},
		{name: "Lowercase site code", siteCode: "us", want: "America/New_York", wantOk: true, wantOffset: "UTC-5/-4"}, // Should work due to ToUpper
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := TimeZone(tt.siteCode)
			if got != tt.want || ok != tt.wantOk {
				t.Errorf("TimeZone(%s) = %s, %v; want %s, %v", tt.siteCode, got, ok, tt.want, tt.wantOk)
			}
			// verify that the returned timezone is valid and check offset
			if ok {
				loc, err := time.LoadLocation(got)
				if err != nil {
					t.Errorf("TimeZone(%s) returned invalid timezone: %s", tt.siteCode, err)
					return
				}

				// Check both January and July to account for DST
				jan := time.Date(2024, 1, 15, 12, 0, 0, 0, loc)
				jul := time.Date(2024, 7, 15, 12, 0, 0, 0, loc)

				_, janOffset := jan.Zone()
				_, julOffset := jul.Zone()

				janHours := float64(janOffset) / 3600
				julHours := float64(julOffset) / 3600

				// Format the actual offset string
				var actualOffset string
				if janOffset == julOffset {
					// No DST
					if janHours == float64(int(janHours)) {
						actualOffset = fmt.Sprintf("UTC%+d", int(janHours))
					} else {
						actualOffset = fmt.Sprintf("UTC%+g", janHours)
					}
				} else {
					// Has DST
					minHours := janHours
					maxHours := julHours
					if julHours < janHours {
						minHours = julHours
						maxHours = janHours
					}

					minStr := fmt.Sprintf("%+g", minHours)
					maxStr := fmt.Sprintf("%+g", maxHours)
					if minHours == float64(int(minHours)) {
						minStr = fmt.Sprintf("%+d", int(minHours))
					}
					if maxHours == float64(int(maxHours)) {
						maxStr = fmt.Sprintf("%+d", int(maxHours))
					}
					actualOffset = fmt.Sprintf("UTC%s/%s", minStr, maxStr)
				}

				if actualOffset != tt.wantOffset {
					t.Errorf("TimeZone(%s) offset = %s; want %s", tt.siteCode, actualOffset, tt.wantOffset)
				}
			}
		})
	}
}

func TestTimeZone_CaseSensitivity(t *testing.T) {
	testCases := []string{"US", "us", "Us", "uS"}
	expectedTZ := "America/New_York"

	for _, tc := range testCases {
		got, ok := TimeZone(tc)
		if !ok || got != expectedTZ {
			t.Errorf("TimeZone(%s) = %s, %v; want %s, true", tc, got, ok, expectedTZ)
		}
	}
}

func TestTimeZone_AllValidTimezones(t *testing.T) {
	// This test ensures all timezones in siteZones are valid IANA timezone identifiers
	for code, tz := range siteZones {
		_, err := time.LoadLocation(tz)
		if err != nil {
			t.Errorf("Invalid timezone for site code %s: %s - %v", code, tz, err)
		}
	}
}
