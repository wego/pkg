package skipfields

// String literals assigned to fields whose names appear in the configured
// skip-fields list must not be flagged, even when the value matches a valid
// ISO code. This file asserts zero diagnostics under
// Settings{SkipFields: ["CardSchemes"]}.
//
// Two assignment shapes are exercised:
//   - Pattern A: receiver-field assignment (a.CardSchemes = ...)
//   - Pattern B: composite-literal struct field (Foo{CardSchemes: ...})
//
// "MC" (Monaco) and "AX" (Aland Islands) are valid ISO 3166-1 alpha-2
// site codes — but here they denote MasterCard and American Express card
// schemes, not countries.

type stringArray []string

type cardAvailability struct {
	CardSchemes     stringArray
	IssuerCountries stringArray
}

// Pattern A — receiver-field assignment.
func patternA() {
	a := &cardAvailability{}
	a.CardSchemes = stringArray{"MC"}
	a.CardSchemes = stringArray{"AX"}
	a.CardSchemes = stringArray{"VISA", "MC"}
	a.CardSchemes = stringArray{"MC", "AX", "JCB"}
}

// Pattern B — composite-literal struct field.
func patternB() {
	_ = &cardAvailability{
		CardSchemes: stringArray{"MC"},
	}
	_ = &cardAvailability{
		CardSchemes: stringArray{"MC", "AX"},
	}
	_ = cardAvailability{CardSchemes: stringArray{"AX"}}
}
