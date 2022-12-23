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
			XMLTag:          "Number",
			FirstCharsToShow: 4,
			LastCharsToShow: 6,
		},
		{
			XMLTag:          "Email",
			FirstCharsToShow: 3,
			LastCharsToShow: 5,
			CharsToIgnore:   []rune{'@'},
		},
		{
			XMLTag:          "SomethingThatDoesNotExists",
			FirstCharsToShow: 2,
			LastCharsToShow: 6,
		},
		{
			XMLTag:          "AgentID",
			FirstCharsToShow: 0,
			LastCharsToShow: 0,
		},
		{
			XMLTag:          "ClientID",
			FirstCharsToShow: 2,
			LastCharsToShow: 0,
		},
		{
			XMLTag:          "BookerID",
			FirstCharsToShow: 0,
			LastCharsToShow: 2,
		},
		{
			XMLTag:          "Phone",
			FirstCharsToShow: 4,
			LastCharsToShow: 3,
			CharsToIgnore:   []rune{'$'},
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
					JSONKeys:        []string{"yo"},
					FirstCharsToShow: 4,
					LastCharsToShow: 6,
				},
				{
					JSONKeys:        []string{"source", "billing_address", "whatsup"},
					FirstCharsToShow: 3,
					LastCharsToShow: 5,
					CharsToIgnore:   []rune{'@'},
				},
				{
					JSONKeys:        []string{"3ds", "hi"},
					FirstCharsToShow: 2,
					LastCharsToShow: 6,
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
			JSONKeys:        []string{"source", "phone", "number"},
			FirstCharsToShow: 2,
			LastCharsToShow: 4,
		},
		{
			JSONKeys:        []string{"yo"},
			FirstCharsToShow: 4,
			LastCharsToShow: 6,
		},
		{
			JSONKeys:        []string{"destination", "phone", "number"},
			FirstCharsToShow: 2,
			LastCharsToShow: 4,
			CharsToIgnore:   []rune{'+'},
		},
		{
			JSONKeys:        []string{"test1"},
			FirstCharsToShow: 0,
			LastCharsToShow: 0,
		},
		{
			JSONKeys:        []string{"test2"},
			FirstCharsToShow: 2,
			LastCharsToShow: 0,
		},
		{
			JSONKeys:        []string{"test3"},
			FirstCharsToShow: 0,
			LastCharsToShow: 3,
		},
		{
			JSONKeys:        []string{"customer", "email"},
			FirstCharsToShow: 5,
			LastCharsToShow: 3,
			CharsToIgnore:   []rune{'@'},
		},
		{
			JSONKeys:        []string{"shipping", "phone", "number"},
			FirstCharsToShow: 2,
			LastCharsToShow: 1,
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
