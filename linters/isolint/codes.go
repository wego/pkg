package isolint

import (
	"fmt"
	"strings"

	"github.com/wego/pkg/currency"
	"github.com/wego/pkg/iso/site"
)

// IsCurrencyCode reports whether s is a known ISO 4217 currency code,
// matching either exact uppercase ("USD") or exact lowercase ("usd").
// Mixed case ("Usd") is not matched.
func IsCurrencyCode(s string) bool {
	if len(s) != 3 {
		return false
	}
	if isAllUpper(s) {
		return currency.IsISO4217(s)
	}
	if isAllLower(s) {
		return currency.IsISO4217(strings.ToUpper(s))
	}
	return false
}

// IsSiteCode reports whether s is a known ISO 3166-1 alpha-2 site code,
// matching either exact uppercase ("SG") or exact lowercase ("sg").
// Mixed case ("Sg") is not matched.
func IsSiteCode(s string) bool {
	if len(s) != 2 {
		return false
	}
	if isAllUpper(s) {
		_, found := site.Currency(s)
		return found
	}
	if isAllLower(s) {
		_, found := site.Currency(strings.ToUpper(s))
		return found
	}
	return false
}

// NormalizeCurrencyCode returns the canonical uppercase form of a currency code.
func NormalizeCurrencyCode(code string) string {
	return strings.ToUpper(code)
}

// NormalizeSiteCode returns the canonical uppercase form of a site code.
func NormalizeSiteCode(code string) string {
	return strings.ToUpper(code)
}

// CurrencyConstName returns the qualified constant name for a currency code,
// e.g., "USD" or "usd" -> "currency.USD".
func CurrencyConstName(code string) string {
	return fmt.Sprintf("currency.%s", NormalizeCurrencyCode(code))
}

// SiteConstName returns the qualified constant name for a site code,
// e.g., "SG" or "sg" -> "site.SG".
func SiteConstName(code string) string {
	return fmt.Sprintf("site.%s", NormalizeSiteCode(code))
}

// isAllUpper reports whether s contains only uppercase ASCII letters.
func isAllUpper(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < 'A' || s[i] > 'Z' {
			return false
		}
	}
	return len(s) > 0
}

// isAllLower reports whether s contains only lowercase ASCII letters.
func isAllLower(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < 'a' || s[i] > 'z' {
			return false
		}
	}
	return len(s) > 0
}
