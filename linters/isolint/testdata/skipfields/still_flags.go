package skipfields

// The skip-fields setting is targeted: only assignments to the configured
// field names are skipped. Assignments to OTHER fields with ISO-like values
// must still be flagged. This file proves that targeting and runs under
// Settings{SkipFields: ["CardSchemes"]}.

func notInSkipList() {
	// IssuerCountries is not in the skip list — site codes here are real
	// countries and should be flagged.
	a := &cardAvailability{}
	a.IssuerCountries = stringArray{"SG"} // want `use site\.SG instead of "SG"`
	a.IssuerCountries = stringArray{"US"} // want `use site\.US instead of "US"`

	// Same field name in a different shape (KeyValueExpr) — also flagged.
	_ = &cardAvailability{
		IssuerCountries: stringArray{"JP"}, // want `use site\.JP instead of "JP"`
	}
}

func bareLiterals() {
	// Bare assignments without a struct field still flag.
	x := "MY" // want `use site\.MY instead of "MY"`
	_ = x

	// Currency literals are unaffected by skip-fields when they are not
	// inside a configured field assignment.
	y := "USD" // want `use currency\.USD instead of "USD"`
	_ = y
}

// The skip is single-target. Tuple assignments (len(Lhs) > 1) fall through
// to the default flag behavior — even when one of the LHS targets is in
// the skip list, the linter cannot reliably correlate which RHS literal
// belongs to which LHS without type information, so it flags everything.
func tupleAssign() {
	a := &cardAvailability{}
	a.CardSchemes, a.IssuerCountries = stringArray{"MC"}, stringArray{"SG"} // want `use site\.MC instead of "MC"` `use site\.SG instead of "SG"`
}

// The skip is targeted at struct-field assignments. A bare local variable
// that happens to share its name with a configured skip field is NOT a
// struct field — it should still flag.
func localVarShadowing() {
	CardSchemes := "MC" // want `use site\.MC instead of "MC"`
	_ = CardSchemes

	var IssuerCountries = "SG" // want `use site\.SG instead of "SG"`
	_ = IssuerCountries
}
