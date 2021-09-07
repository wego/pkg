package localization

import (
	"context"
	"strings"
	"time"
)

// I couldn't find any office source. But I try to cover at least common RTL languages.
// One website to verify it: https://www.localeplanet.com/icu/ar/index.html
var (
	rtlLocales map[string]bool = map[string]bool{
		LocaleArabic:       true,
		LocaleFarsiPersian: true,
		LocaleHebrew:       true,
		LocaleKashmiri:     true,
		LocalePashto:       true,
		LocaleUyghur:       true,
		LocaleUrdu:         true,
		LocaleYiddish:      true,
	}

	locations map[string]string = map[string]string{
		SiteCodeAlgeria:               LocationAlgiers,
		SiteCodeAngola:                LocationLuanda,
		SiteCodeArgentina:             LocationBuenosAires,
		SiteCodeAustralian:            LocationSydney,
		SiteCodeBahrain:               LocationBahrain,
		SiteCodeBangladesh:            LocationDhaka,
		SiteCodeBotswana:              LocationGaborone,
		SiteCodeBrazil:                LocationSaoPaulo,
		SiteCodeCameroon:              LocationDouala,
		SiteCodeCanada:                LocationToronto,
		SiteCodeChile:                 LocationSantiago,
		SiteCodeChina:                 LocationShanghai,
		SiteCodeColumbia:              LocationBogota,
		SiteCodeCongo:                 LocationKinshasa,
		SiteCodeEgypt:                 LocationCairo,
		SiteCodeEthiopia:              LocationAddisAbaba,
		SiteCodeFrance:                LocationParis,
		SiteCodeGabon:                 LocationLibreville,
		SiteCodeGermany:               LocationBerlin,
		SiteCodeGhana:                 LocationAccra,
		SiteCodeGreatBritain:          LocationLondon,
		SiteCodeHongKong:              LocationHongKong,
		SiteCodeIndia:                 LocationKolkata,
		SiteCodeIndonesia:             LocationJakarta,
		SiteCodeIran:                  LocationTehran,
		SiteCodeIraq:                  LocationBaghdad,
		SiteCodeIreland:               LocationDublin,
		SiteCodeItaly:                 LocationRome,
		SiteCodeIvoryCoast:            LocationAbidjian,
		SiteCodeJapan:                 LocationTokyo,
		SiteCodeJordan:                LocationAmman,
		SiteCodeKenya:                 LocationNairobi,
		SiteCodeKorea:                 LocationSeoul,
		SiteCodeKuwait:                LocationKuwait,
		SiteCodeLebanon:               LocationBeirut,
		SiteCodeLibya:                 LocationTripoli,
		SiteCodeMacau:                 LocationMacau,
		SiteCodeMalaysia:              LocationKualaLumpur,
		SiteCodeMali:                  LocationBamako,
		SiteCodeMauritius:             LocationMauritius,
		SiteCodeMexio:                 LocationMexicoCity,
		SiteCodeMorocco:               LocationCasablanca,
		SiteCodeMozambique:            LocationMaputo,
		SiteCodeNamibia:               LocationWindhoek,
		SiteCodeNetherlands:           LocationAmsterdam,
		SiteCodeNewZealand:            LocationAuckland,
		SiteCodeNigeria:               LocationLagos,
		SiteCodeOman:                  LocationMuscat,
		SiteCodePakistan:              LocationKarachi,
		SiteCodePalestine:             LocationHebron,
		SiteCodePhilippines:           LocationManila,
		SiteCodePoland:                LocationWarsaw,
		SiteCodePortugal:              LocationLisbon,
		SiteCodeQatar:                 LocationQatar,
		SiteCodeRussia:                LocationMoscow,
		SiteCodeRwanda:                LocationKigali,
		SiteCodeSaudiArabia:           LocationRiyadh,
		SiteCodeSenegal:               LocationDakar,
		SiteCodeSingapore:             LocationSingapore,
		SiteCodeSouthAfrica:           LocationJohannesburg,
		SiteCodeSpain:                 LocationMadrid,
		SiteCodeSriLanka:              LocationColombo,
		SiteCodeSudan:                 LocationKhartoum,
		SiteCodeSweden:                LocationStockholm,
		SiteCodeSwitzerland:           LocationZurich,
		SiteCodeSyria:                 LocationDamascus,
		SiteCodeTaiwan:                LocationTaipei,
		SiteCodeTanzania:              LocationDaresSalaam,
		SiteCodeThailand:              LocationBangkok,
		SiteCodeTunisia:               LocationTunis,
		SiteCodeTurkey:                LocationIstanbul,
		SiteCodeUAE:                   LocationDubai,
		SiteCodeUganda:                LocationKampala,
		SiteCodeUnitedKingdom:         LocationLondon,
		SiteCodeUnitedStatesOfAmerica: LocationLosAngeles,
		SiteCodeVietnam:               LocationHoChiMinh,
		SiteCodeZambia:                LocationLusaka,
		SiteCodeZimbabwe:              LocationHarare,
	}
)

// NewContextWithLocale returns a new context with input locale (or default locale if input is empty)
func NewContextWithLocale(ctx context.Context, locale string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if locale == "" {
		locale = DefaultLocale
	}
	return context.WithValue(ctx, contextKeyLocale, strings.ToLower(locale))
}

// LocaleFromContext tries to get locale from ctx, will return default locale if cannot
func LocaleFromContext(ctx context.Context) string {
	if ctx == nil {
		return DefaultLocale
	}

	locale, ok := ctx.Value(contextKeyLocale).(string)
	if !ok {
		return DefaultLocale
	}
	return locale
}

// TextDirectionFromLocale return text direction from locale
func TextDirectionFromLocale(locale string) TextDirection {
	if len(locale) >= 2 {
		locale = strings.ToLower(locale)[0:2]
		_, ok := rtlLocales[locale]
		if ok {
			return TextDirectionRTL
		}

	}
	return TextDirectionLTR
}

// LoadLocation loads location from siteCode, default to Asia/Dubai
func LoadLocation(siteCode string) *time.Location {
	var loc *time.Location
	name, ok := locations[strings.ToLower(siteCode)]
	if ok {
		loc, _ = time.LoadLocation(name)
	}

	if loc == nil {
		loc, _ = time.LoadLocation(LocationDubai)
	}
	return loc
}
