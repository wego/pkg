package logger

import (
	"fmt"
	"net/url"
)

func ExampleMaskJSON() {
	maskData := []MaskData{
		{
			JSONKeys:         []string{"first"},
			FirstCharsToShow: 3,
			LastCharsToShow:  6,
			KeepSameLength:   true,
		},
		{
			JSONKeys:         []string{"second", "second"},
			FirstCharsToShow: 2,
			LastCharsToShow:  3,
			CharsToIgnore:    []rune{'@'},
			RestrictionType:  MaskRestrictionTypeEmail,
			KeepSameLength:   true,
		},
		{
			JSONKeys:         []string{"second", "third", "first"},
			FirstCharsToShow: 3,
			LastCharsToShow:  1,
			KeepSameLength:   true,
		},
		{
			JSONKeys:         []string{"second", "fourth", "[]", "email", "value"},
			FirstCharsToShow: 1,
			LastCharsToShow:  3,
			CharsToIgnore:    []rune{'@'},
			KeepSameLength:   true,
		},
	}

	input := `
{
  "first": "first value",
  "second": {
    "first": "1st of second",
    "second": "not-an-email.com",
    "third": {
      "first": "1st of second third",
      "second": "2nd of second third",
      "third": "3rd of second third"
    },
    "fourth": [
      { "email": { "value": "first@email.com" } },
      { "email": { "value": "second@email.com" } }
    ]
  }
}	
`

	output := MaskJSON(input, "*", maskData)
	_, _ = fmt.Println(output)
	// Output:
	// {"first":"fir** value","second":{"first":"1st of second","second":"not-an-email.com","third":{"first":"1st***************d","second":"2nd of second third","third":"3rd of second third"},"fourth":[{"email":{"value":"f****@******com"}},{"email":{"value":"s*****@******com"}}]}}
}

func ExampleRedactJSON() {
	keys := [][]string{
		{"first"},
		{"second", "second"},
		{"second", "third", "first"},
		{"third", "[]", "value"},
	}

	input := `
{
  "first": "first value",
  "second": {
    "first": "1st of second",
    "second": "2nd of second",
    "third": {
      "first": "1st of second third",
      "second": "2nd of second third",
      "third": "3rd of second third"
    }
  },
  "third": [
    { "value": "third value" }
  ]
}
`

	output := RedactJSON(input, "Wego", keys)
	_, _ = fmt.Println(output)
	// Output:
	// {"first":"Wego","second":{"first":"1st of second","second":"Wego","third":{"first":"Wego","second":"2nd of second third","third":"3rd of second third"}},"third":[{"value":"Wego"}]}
}

func ExampleRedactFormURLEncoded() {
	keys := []string{
		"field1",
		"field3",
		"field4.nested.data",
	}

	formData := url.Values{
		"field1":             []string{"field1value1", "field1value2"},
		"field2":             []string{"field2value1"},
		"field3":             []string{"sensitive_data"},
		"field4.nested.data": []string{"data"},
	}
	input := formData.Encode()

	output := RedactFormURLEncoded(input, "Wego", keys)
	_, _ = fmt.Println(output)
	// Output:
	// field1=Wego&field1=Wego&field2=field2value1&field3=Wego&field4.nested.data=Wego
}
