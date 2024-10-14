package logger

import (
	"fmt"
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
	fmt.Println(output)
	// Output:
	// {"first":"fir** value","second":{"first":"1st of second","second":"not-an-email.com","third":{"first":"1st***************d","second":"2nd of second third","third":"3rd of second third"},"fourth":[{"email":{"value":"f****@******com"}},{"email":{"value":"s*****@******com"}}]}}
}
