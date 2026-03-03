package isolint

import (
	"fmt"
	"strings"

	"github.com/wego/pkg/currency"
	"github.com/wego/pkg/iso/site"
)

// IsCurrencyCode reports whether s is a known uppercase ISO 4217 currency
// code (e.g. "USD"). Lowercase ("usd") and mixed case ("Usd") are not
// matched — only uppercase strings are considered intentional ISO references.
func IsCurrencyCode(s string) bool {
	if len(s) != 3 {
		return false
	}
	if !isAllUpper(s) {
		return false
	}
	return currency.IsISO4217(s)
}

// IsSiteCode reports whether s is a known uppercase ISO 3166-1 alpha-2 site
// code (e.g. "SG"). Lowercase ("sg") and mixed case ("Sg") are not matched
// — only uppercase strings are considered intentional ISO references.
func IsSiteCode(s string) bool {
	if len(s) != 2 {
		return false
	}
	if !isAllUpper(s) {
		return false
	}
	_, found := site.Currency(s)
	return found
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
// e.g., "USD" -> "currency.USD".
func CurrencyConstName(code string) string {
	return fmt.Sprintf("currency.%s", NormalizeCurrencyCode(code))
}

// SiteConstName returns the qualified constant name for a site code,
// e.g., "SG" -> "site.SG".
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

