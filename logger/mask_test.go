package logger_test

import (
	"bytes"
	"encoding/json"
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
