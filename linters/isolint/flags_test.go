package isolint

import (
	"testing"
)

// TestSkipFieldsFlag verifies that the -skip-fields flag exposed on the
// analyzer's FlagSet accepts a comma-separated list and updates the
// closure-captured state used by the run function.
func TestSkipFieldsFlag(t *testing.T) {
	a := NewAnalyzer(Settings{})

	flag := a.Flags.Lookup("skip-fields")
	if flag == nil {
		t.Fatal("expected -skip-fields flag to be registered on analyzer")
	}

	if err := flag.Value.Set("CardSchemes,IssuerCountries"); err != nil {
		t.Fatalf("flag.Value.Set returned error: %v", err)
	}

	got := flag.Value.String()
	want := "CardSchemes,IssuerCountries"
	if got != want {
		t.Errorf("flag.Value.String() = %q, want %q", got, want)
	}
}

// TestSkipFieldsFlagSeededFromSettings verifies that constructor-provided
// settings populate the flag's default value, so plugin and analystest
// callers (which do not parse flags) still get the configured behavior.
func TestSkipFieldsFlagSeededFromSettings(t *testing.T) {
	a := NewAnalyzer(Settings{SkipFields: []string{"CardSchemes"}})

	flag := a.Flags.Lookup("skip-fields")
	if flag == nil {
		t.Fatal("expected -skip-fields flag to be registered on analyzer")
	}

	got := flag.Value.String()
	want := "CardSchemes"
	if got != want {
		t.Errorf("seeded flag.Value.String() = %q, want %q", got, want)
	}
}
