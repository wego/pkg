package logger_test

import (
	"bytes"
	"encoding/json"
	"net/url"
	"os"
	"testing"

	"github.com/antchfx/xmlquery"
	"github.com/stretchr/testify/assert"
	"github.com/wego/pkg/errors"
	"github.com/wego/pkg/logger"
)

const (
	testFileDir = "./testfiles/mask/"
)

func Test_MaskXML_ReturnErrorText_WithInvalidInput(t *testing.T) {
	assert := assert.New(t)

	output := logger.MaskXML("<wego", "", []logger.MaskData{})
	assert.Contains(output, "invalid XML input")
}

func Test_MaskXML_DoNothing_WhenInputEmptyTags(t *testing.T) {
	assert := assert.New(t)

	input, err := parseXMLToString("mask_input.xml")

	output := logger.MaskXML(input, "", []logger.MaskData{})
	assert.NoError(err)
	assert.Equal(input, output)
}

func Test_MaskXML_Ok(t *testing.T) {
	assert := assert.New(t)
	maskData := []logger.MaskData{
		{
			XMLTag:           "Test1",
			FirstCharsToShow: 4,
			LastCharsToShow:  6,
			KeepSameLength:   true,
		},
		{
			XMLTag:           "Number",
			FirstCharsToShow: 4,
			LastCharsToShow:  6,
			KeepSameLength:   true,
		},
		{
			XMLTag:           "Email",
			FirstCharsToShow: 3,
			LastCharsToShow:  5,
			CharsToIgnore:    []rune{'@'},
			KeepSameLength:   true,
		},
		{
			XMLTag:           "SomethingThatDoesNotExists",
			FirstCharsToShow: 2,
			LastCharsToShow:  6,
			KeepSameLength:   true,
		},
		{
			XMLTag:           "AgentID",
			FirstCharsToShow: 0,
			LastCharsToShow:  0,
			KeepSameLength:   true,
		},
		{
			XMLTag:           "ClientID",
			FirstCharsToShow: 2,
			LastCharsToShow:  0,
			KeepSameLength:   true,
		},
		{
			XMLTag:           "BookerID",
			FirstCharsToShow: 0,
			LastCharsToShow:  2,
			KeepSameLength:   true,
		},
		{
			XMLTag:           "Phone",
			FirstCharsToShow: 4,
			LastCharsToShow:  3,
			CharsToIgnore:    []rune{'$'},
			KeepSameLength:   true,
		},
		{
			XMLTag:           "UserID1",
			FirstCharsToShow: 5,
			LastCharsToShow:  6,
			CharsToIgnore:    []rune{'@'},
			RestrictionType:  logger.MaskRestrictionTypeEmail,
			KeepSameLength:   true,
		},
		{
			XMLTag:           "UserID2",
			FirstCharsToShow: 5,
			LastCharsToShow:  6,
			CharsToIgnore:    []rune{'@'},
			RestrictionType:  logger.MaskRestrictionTypeEmail,
			KeepSameLength:   true,
		},
	}
	input, err := parseXMLToString("mask_input.xml")
	assert.NoError(err)

	for _, testCase := range []struct {
		expectedTestFile string
		replacement      string
	}{
		{
			expectedTestFile: "expected_mask_input.xml",
			replacement:      "|",
		},
		{
			expectedTestFile: "expected_default_mask_input.xml",
		},
	} {
		output := logger.MaskXML(input, testCase.replacement, maskData)
		expected, err := parseXMLToString(testCase.expectedTestFile)
		assert.NoError(err)
		assert.Equal(expected, output)
	}
}

func parseXMLToString(xmlFileName string) (string, error) {
	file, err := os.ReadFile(testFileDir + xmlFileName)
	if err != nil {
		return "", err
	}

	doc, err := xmlquery.Parse(bytes.NewReader(file))
	if err != nil {
		return "", errors.New("invalid XML input", err)
	}

	return doc.OutputXML(true), nil
}

func Test_MaskJSON_InvalidInput(t *testing.T) {
	assert := assert.New(t)

	output := logger.MaskJSON("<123 invalid }", "", []logger.MaskData{})
	assert.Contains(output, "cannot parse JSON")
}

func Test_MaskJSON_DoNothing_WhenKeysNotFound(t *testing.T) {
	assert := assert.New(t)
	var compactOutput, compactExpectedOutput bytes.Buffer
	input, err := parseJSONToString("mask_input.json")
	assert.NoError(err)

	for _, testCase := range []struct {
		maskData []logger.MaskData
	}{
		{
			maskData: nil,
		},
		{
			maskData: []logger.MaskData{},
		},
		{
			maskData: []logger.MaskData{
				{
					JSONKeys:         []string{"yo"},
					FirstCharsToShow: 4,
					LastCharsToShow:  6,
					KeepSameLength:   true,
				},
				{
					JSONKeys:         []string{"source", "billing_address", "whatsup"},
					FirstCharsToShow: 3,
					LastCharsToShow:  5,
					CharsToIgnore:    []rune{'@'},
					KeepSameLength:   true,
				},
				{
					JSONKeys:         []string{"3ds", "hi"},
					FirstCharsToShow: 2,
					LastCharsToShow:  6,
					KeepSameLength:   true,
				},
			},
		},
	} {

		// not provide MaskData
		output := logger.MaskJSON(input, "", testCase.maskData)
		err = json.Compact(&compactOutput, []byte(output))
		assert.NoError(err)
		err = json.Compact(&compactExpectedOutput, []byte(input))
		assert.NoError(err)
		assert.Equal(compactExpectedOutput, compactOutput)
	}
}

func Test_MaskJSON_Ok(t *testing.T) {
	maskData := []logger.MaskData{
		{
			JSONKeys:         []string{"source", "phone", "number"},
			FirstCharsToShow: 2,
			LastCharsToShow:  4,
			KeepSameLength:   true,
		},
		{
			JSONKeys:         []string{"yo"},
			FirstCharsToShow: 4,
			LastCharsToShow:  6,
			KeepSameLength:   true,
		},
		{
			JSONKeys:         []string{"destination", "phone", "number"},
			FirstCharsToShow: 2,
			LastCharsToShow:  4,
			CharsToIgnore:    []rune{'+'},
			KeepSameLength:   true,
		},
		{
			JSONKeys:         []string{"test1"},
			FirstCharsToShow: 0,
			LastCharsToShow:  0,
			KeepSameLength:   true,
		},
		{
			JSONKeys:         []string{"test2"},
			FirstCharsToShow: 2,
			LastCharsToShow:  0,
			KeepSameLength:   true,
		},
		{
			JSONKeys:         []string{"test3"},
			FirstCharsToShow: 0,
			LastCharsToShow:  3,
			KeepSameLength:   true,
		},
		{
			JSONKeys:         []string{"test4"},
			FirstCharsToShow: 3,
			LastCharsToShow:  3,
			KeepSameLength:   true,
		},
		{
			JSONKeys:         []string{"test5", "[]", "customer", "name"},
			FirstCharsToShow: 2,
			LastCharsToShow:  3,
			KeepSameLength:   false,
		},
		{
			JSONKeys:         []string{"test5", "[]", "customer", "email"},
			FirstCharsToShow: 2,
			LastCharsToShow:  3,
			CharsToIgnore:    []rune{'@'},
			KeepSameLength:   true,
		},
		{
			JSONKeys:         []string{"test6", "[]", "nested", "[]", "value"},
			FirstCharsToShow: 1,
			LastCharsToShow:  1,
			KeepSameLength:   true,
		},
		{
			JSONKeys:         []string{"customer", "email"},
			FirstCharsToShow: 5,
			LastCharsToShow:  3,
			CharsToIgnore:    []rune{'@'},
			KeepSameLength:   true,
		},
		{
			JSONKeys:         []string{"shipping", "phone", "number"},
			FirstCharsToShow: 2,
			LastCharsToShow:  1,
			KeepSameLength:   true,
		},
		{
			JSONKeys:         []string{"risk", "userId1"},
			FirstCharsToShow: 2,
			LastCharsToShow:  7,
			CharsToIgnore:    []rune{'@'},
			KeepSameLength:   true,
		},
		{
			JSONKeys:         []string{"risk", "userId2"},
			FirstCharsToShow: 2,
			LastCharsToShow:  7,
			RestrictionType:  logger.MaskRestrictionTypeEmail,
			CharsToIgnore:    []rune{'@'},
			KeepSameLength:   true,
		},
	}

	assert := assert.New(t)
	var compactOutput, compactExpectedOutput bytes.Buffer

	input, err := parseJSONToString("mask_input.json")
	assert.NoError(err)

	for _, testCase := range []struct {
		expectedTestFile string
		replacement      string
	}{
		{
			expectedTestFile: "expected_mask_input.json",
			replacement:      "|",
		},
		{
			expectedTestFile: "expected_default_mask_input.json",
		},
	} {
		output := logger.MaskJSON(input, testCase.replacement, maskData)
		err := json.Compact(&compactOutput, []byte(output))
		assert.NoError(err)

		expected, err := parseJSONToString(testCase.expectedTestFile)
		assert.NoError(err)
		err = json.Compact(&compactExpectedOutput, []byte(expected))
		assert.NoError(err)
		assert.Equal(compactExpectedOutput, compactOutput)
	}
}

func parseJSONToString(jsonFileName string) (string, error) {
	file, err := os.ReadFile(testFileDir + jsonFileName)
	if err != nil {
		return "", err
	}
	return string(file[:]), nil
}

func BenchmarkMaskJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = logger.MaskJSON(jsonInput, "*", []logger.MaskData{
			{
				JSONKeys:         []string{"test1"},
				FirstCharsToShow: 0,
				LastCharsToShow:  0,
				KeepSameLength:   true,
			},
			{
				JSONKeys:         []string{"test2"},
				FirstCharsToShow: 2,
				LastCharsToShow:  0,
				KeepSameLength:   true,
			},
			{
				JSONKeys:         []string{"test3"},
				FirstCharsToShow: 0,
				LastCharsToShow:  3,
				KeepSameLength:   true,
			},
			{
				JSONKeys:         []string{"test4"},
				FirstCharsToShow: 3,
				LastCharsToShow:  3,
				KeepSameLength:   true,
			},
			{
				JSONKeys:         []string{"test5", "[]", "customer", "name"},
				FirstCharsToShow: 2,
				LastCharsToShow:  3,
				KeepSameLength:   false,
			},
			{
				JSONKeys:         []string{"test5", "[]", "customer", "email"},
				FirstCharsToShow: 2,
				LastCharsToShow:  3,
				CharsToIgnore:    []rune{'@'},
				KeepSameLength:   true,
			},
			{
				JSONKeys:         []string{"test6", "[]", "nested", "[]", "value"},
				FirstCharsToShow: 1,
				LastCharsToShow:  1,
				KeepSameLength:   true,
			},
		})
	}
}

func BenchmarkMaskJSONParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = logger.MaskJSON(jsonInput, "*", []logger.MaskData{
				{
					JSONKeys:         []string{"test1"},
					FirstCharsToShow: 0,
					LastCharsToShow:  0,
					KeepSameLength:   true,
				},
				{
					JSONKeys:         []string{"test2"},
					FirstCharsToShow: 2,
					LastCharsToShow:  0,
					KeepSameLength:   true,
				},
				{
					JSONKeys:         []string{"test3"},
					FirstCharsToShow: 0,
					LastCharsToShow:  3,
					KeepSameLength:   true,
				},
				{
					JSONKeys:         []string{"test4"},
					FirstCharsToShow: 3,
					LastCharsToShow:  3,
					KeepSameLength:   true,
				},
				{
					JSONKeys:         []string{"test5", "[]", "customer", "name"},
					FirstCharsToShow: 2,
					LastCharsToShow:  3,
					KeepSameLength:   false,
				},
				{
					JSONKeys:         []string{"test5", "[]", "customer", "email"},
					FirstCharsToShow: 2,
					LastCharsToShow:  3,
					CharsToIgnore:    []rune{'@'},
					KeepSameLength:   true,
				},
				{
					JSONKeys:         []string{"test6", "[]", "nested", "[]", "value"},
					FirstCharsToShow: 1,
					LastCharsToShow:  1,
					KeepSameLength:   true,
				},
			})
		}
	})
}

func TestMaskFormURLEncoded(t *testing.T) {
	assert := assert.New(t)

	maskData := []logger.MaskData{
		{
			JSONKeys:         []string{"field1"},
			FirstCharsToShow: 2,
			LastCharsToShow:  2,
			KeepSameLength:   false,
		},
		{
			JSONKeys:         []string{"field3"},
			FirstCharsToShow: 0,
			LastCharsToShow:  0,
			KeepSameLength:   true,
		},
		{
			JSONKeys:         []string{"field4", "but-nested-doesnt-matter"},
			FirstCharsToShow: 0,
			LastCharsToShow:  0,
			KeepSameLength:   true,
		},
	}

	formData := url.Values{
		"field1": []string{"field1value1", "field1value2"},
		"field2": []string{"field2value1"},
		"field3": []string{"sensitive_data"},
		"field4": []string{"data"},
	}
	input := formData.Encode()

	output := logger.MaskFormURLEncoded(input, "*", maskData)

	expected := "field1=fi*e1&field1=fi*e2&field2=field2value1&field3=**************&field4=****"
	expectedFormData, err := url.ParseQuery(expected)
	assert.NoError(err)
	assert.Equal(expectedFormData.Encode(), output)
}

func BenchmarkMaskFormURLEncoded(b *testing.B) {
	maskData := []logger.MaskData{
		{
			JSONKeys:         []string{"field1"},
			FirstCharsToShow: 2,
			LastCharsToShow:  2,
			KeepSameLength:   false,
		},
		{
			JSONKeys:         []string{"field3"},
			FirstCharsToShow: 0,
			LastCharsToShow:  0,
			KeepSameLength:   true,
		},
	}

	formData := url.Values{
		"field1": []string{"field1value1", "field1value2"},
		"field2": []string{"field2value1"},
		"field3": []string{"sensitive_data"},
	}
	input := formData.Encode()

	for i := 0; i < b.N; i++ {
		_ = logger.MaskFormURLEncoded(input, "*", maskData)
	}
}

func BenchmarkMaskFormURLEncodedParallel(b *testing.B) {
	maskData := []logger.MaskData{
		{
			JSONKeys:         []string{"field1"},
			FirstCharsToShow: 2,
			LastCharsToShow:  2,
			KeepSameLength:   false,
		},
		{
			JSONKeys:         []string{"field3"},
			FirstCharsToShow: 0,
			LastCharsToShow:  0,
			KeepSameLength:   true,
		},
	}

	formData := url.Values{
		"field1": []string{"field1value1", "field1value2"},
		"field2": []string{"field2value1"},
		"field3": []string{"sensitive_data"},
	}
	input := formData.Encode()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = logger.MaskFormURLEncoded(input, "*", maskData)
		}
	})
}

func TestMaskURLQueryParams(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		name     string
		rawURL   string
		maskChar string
		toMasks  []logger.MaskData
		expected string
	}{
		{
			name:     "mask single parameter with default mask character",
			rawURL:   "https://example.com/webhook?email_address=test@example.com&amount=100",
			maskChar: "",
			toMasks: []logger.MaskData{
				{
					JSONKeys:         []string{"email_address"},
					FirstCharsToShow: 0,
					LastCharsToShow:  0,
					KeepSameLength:   true,
				},
			},
			expected: "https://example.com/webhook?amount=100&email_address=%2A%2A%2A%2A%2A%2A%2A%2A%2A%2A%2A%2A%2A%2A%2A%2A",
		},
		{
			name:     "mask multiple parameters with custom mask character",
			rawURL:   "https://example.com/webhook?email_address=test@example.com&mobile_no=1234567890&amount=100",
			maskChar: "X",
			toMasks: []logger.MaskData{
				{
					JSONKeys:         []string{"email_address"},
					FirstCharsToShow: 0,
					LastCharsToShow:  0,
					KeepSameLength:   true,
				},
				{
					JSONKeys:         []string{"mobile_no"},
					FirstCharsToShow: 0,
					LastCharsToShow:  0,
					KeepSameLength:   true,
				},
			},
			expected: "https://example.com/webhook?amount=100&email_address=XXXXXXXXXXXXXXXX&mobile_no=XXXXXXXXXX",
		},
		{
			name:     "mask with partial showing - first and last chars",
			rawURL:   "https://example.com/webhook?email_address=test@example.com&mobile_no=1234567890",
			maskChar: "~",
			toMasks: []logger.MaskData{
				{
					JSONKeys:         []string{"email_address"},
					FirstCharsToShow: 2,
					LastCharsToShow:  4,
					CharsToIgnore:    []rune{'@'},
					KeepSameLength:   true,
				},
				{
					JSONKeys:         []string{"mobile_no"},
					FirstCharsToShow: 3,
					LastCharsToShow:  2,
					KeepSameLength:   true,
				},
			},
			expected: "https://example.com/webhook?email_address=te~~%40~~~~~~~.com&mobile_no=123~~~~~90",
		},
		{
			name:     "mask with different lengths - keep same length false",
			rawURL:   "https://example.com/webhook?field1=verylongvalue&field2=short",
			maskChar: "#",
			toMasks: []logger.MaskData{
				{
					JSONKeys:         []string{"field1"},
					FirstCharsToShow: 2,
					LastCharsToShow:  2,
					KeepSameLength:   false,
				},
				{
					JSONKeys:         []string{"field2"},
					FirstCharsToShow: 1,
					LastCharsToShow:  1,
					KeepSameLength:   false,
				},
			},
			expected: "https://example.com/webhook?field1=ve%23ue&field2=s%23t",
		},
		{
			name:     "no parameters to mask",
			rawURL:   "https://example.com/webhook?amount=100&currency=USD",
			maskChar: "*",
			toMasks: []logger.MaskData{
				{
					JSONKeys:         []string{"email_address"},
					FirstCharsToShow: 0,
					LastCharsToShow:  0,
					KeepSameLength:   true,
				},
			},
			expected: "https://example.com/webhook?amount=100&currency=USD",
		},
		{
			name:     "parameter not found",
			rawURL:   "https://example.com/webhook?name=john&age=30",
			maskChar: "*",
			toMasks: []logger.MaskData{
				{
					JSONKeys:         []string{"email_address"},
					FirstCharsToShow: 0,
					LastCharsToShow:  0,
					KeepSameLength:   true,
				},
			},
			expected: "https://example.com/webhook?name=john&age=30",
		},
		{
			name:     "invalid URL returns original",
			rawURL:   "not-a-valid-url",
			maskChar: "*",
			toMasks: []logger.MaskData{
				{
					JSONKeys:         []string{"email_address"},
					FirstCharsToShow: 0,
					LastCharsToShow:  0,
					KeepSameLength:   true,
				},
			},
			expected: "not-a-valid-url",
		},
		{
			name:     "URL without query parameters",
			rawURL:   "https://example.com/webhook",
			maskChar: "*",
			toMasks: []logger.MaskData{
				{
					JSONKeys:         []string{"email_address"},
					FirstCharsToShow: 0,
					LastCharsToShow:  0,
					KeepSameLength:   true,
				},
			},
			expected: "https://example.com/webhook",
		},
		{
			name:     "empty toMasks list",
			rawURL:   "https://example.com/webhook?email_address=test@example.com",
			maskChar: "*",
			toMasks:  []logger.MaskData{},
			expected: "https://example.com/webhook?email_address=test@example.com",
		},
		{
			name:     "mask parameter with special characters and URL encoding",
			rawURL:   "https://example.com/webhook?email_address=test%40example.com&mobile_no=%2B1234567890",
			maskChar: "#",
			toMasks: []logger.MaskData{
				{
					JSONKeys:         []string{"email_address"},
					FirstCharsToShow: 2,
					LastCharsToShow:  4,
					CharsToIgnore:    []rune{'@'},
					KeepSameLength:   true,
				},
				{
					JSONKeys:         []string{"mobile_no"},
					FirstCharsToShow: 1,
					LastCharsToShow:  3,
					KeepSameLength:   true,
				},
			},
			expected: "https://example.com/webhook?email_address=te%23%23%40%23%23%23%23%23%23%23.com&mobile_no=%2B%23%23%23%23%23%23%23890",
		},
		{
			name:     "mask multiple values for same parameter",
			rawURL:   "https://example.com/webhook?tags=sensitive1&tags=sensitive2&tags=public",
			maskChar: "~",
			toMasks: []logger.MaskData{
				{
					JSONKeys:         []string{"tags"},
					FirstCharsToShow: 2,
					LastCharsToShow:  1,
					KeepSameLength:   true,
				},
			},
			expected: "https://example.com/webhook?tags=se~~~~~~~1&tags=se~~~~~~~2&tags=pu~~~c",
		},
		{
			name:     "email restriction type test",
			rawURL:   "https://example.com/webhook?user_id=test@example.com&account_id=12345",
			maskChar: "~",
			toMasks: []logger.MaskData{
				{
					JSONKeys:         []string{"user_id"},
					FirstCharsToShow: 2,
					LastCharsToShow:  4,
					CharsToIgnore:    []rune{'@'},
					RestrictionType:  logger.MaskRestrictionTypeEmail,
					KeepSameLength:   true,
				},
				{
					JSONKeys:         []string{"account_id"},
					FirstCharsToShow: 2,
					LastCharsToShow:  4,
					RestrictionType:  logger.MaskRestrictionTypeEmail,
					KeepSameLength:   true,
				},
			},
			expected: "https://example.com/webhook?account_id=12345&user_id=te~~%40~~~~~~~.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := logger.MaskURLQueryParams(tc.rawURL, tc.maskChar, tc.toMasks)
			assert.Equal(tc.expected, result)
		})
	}
}

func BenchmarkMaskURLQueryParams(b *testing.B) {
	rawURL := "https://example.com/webhook?email_address=test@example.com&mobile_no=1234567890&amount=100&currency=USD"
	toMasks := []logger.MaskData{
		{
			JSONKeys:         []string{"email_address"},
			FirstCharsToShow: 2,
			LastCharsToShow:  4,
			CharsToIgnore:    []rune{'@'},
			KeepSameLength:   true,
		},
		{
			JSONKeys:         []string{"mobile_no"},
			FirstCharsToShow: 3,
			LastCharsToShow:  2,
			KeepSameLength:   true,
		},
	}

	for i := 0; i < b.N; i++ {
		_ = logger.MaskURLQueryParams(rawURL, "~", toMasks)
	}
}

func BenchmarkMaskURLQueryParamsParallel(b *testing.B) {
	rawURL := "https://example.com/webhook?email_address=test@example.com&mobile_no=1234567890&amount=100&currency=USD"
	toMasks := []logger.MaskData{
		{
			JSONKeys:         []string{"email_address"},
			FirstCharsToShow: 2,
			LastCharsToShow:  4,
			CharsToIgnore:    []rune{'@'},
			KeepSameLength:   true,
		},
		{
			JSONKeys:         []string{"mobile_no"},
			FirstCharsToShow: 3,
			LastCharsToShow:  2,
			KeepSameLength:   true,
		},
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = logger.MaskURLQueryParams(rawURL, "~", toMasks)
		}
	})
}
