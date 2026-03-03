package example

import (
	"io"

	"github.com/wego/pkg/currency"
	"github.com/wego/pkg/iso/site"
)

func validCode() {
	// Already using package constants — no flags
	_ = currency.USD
	_ = currency.EUR
	_ = currency.SGD
	_ = site.SG
	_ = site.US
	_ = site.JP

	// Non-ISO strings — no flags
	if "hello" == "world" {
		println("fine")
	}

	x := "not a code"
	_ = x

	// Numbers and booleans — no flags
	y := 42
	_ = y

	// Mixed case — no flags (only exact uppercase is matched)
	_ = "Usd"
	_ = "Sg"
	_ = "uSd"

	// Strings that are too long or too short — no flags
	_ = "USDD"
	_ = "S"
	_ = ""

	// Strings that look like codes but aren't real — no flags
	_ = "XYZ"
	_ = "ZZ"
	_ = "QQ"
}

// Import paths must not be flagged.
// "io" is a 2-char string that matches site.IO (British Indian Ocean Territory),
// but as an import path it is clearly a package name, not a country code.
func useIO() {
	var r io.Reader
	_ = r
}

// Lowercase strings are never flagged — they are common in Go source as
// identifiers, parameter names, and English words. Only uppercase forms
// are considered intentional ISO code references.
func lowercaseStrings() {
	_ = "id" // identifier, not Indonesia
	_ = "to" // preposition, not Tonga
	_ = "no" // negation, not Norway
	_ = "do" // verb, not Dominican Republic
	_ = "io" // I/O, not British Indian Ocean Territory
	_ = "is" // verb, not Iceland
	_ = "sg" // could be anything lowercase
	_ = "us" // pronoun, not United States

	_ = "all" // "everything", not Albanian Lek
	_ = "usd" // lowercase, not linted
	_ = "sgd" // lowercase, not linted
}
