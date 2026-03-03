package example

import (
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

	// Mixed case — no flags (only exact upper/lowercase matched)
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
