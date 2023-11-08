package logger

import (
	"net/mail"
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/valyala/fastjson"
	"github.com/wego/pkg/collection"
	"github.com/wego/pkg/errors"
)

// MaskRestrictionType the type of text to mask, will only mask if text is of the specified type
type MaskRestrictionType string

const (
	// MaskRestrictionTypeEmail will only mask email text
	MaskRestrictionTypeEmail MaskRestrictionType = "email"
)

// MaskData the data as well as the information on how to mask the data
type MaskData struct {
	FirstCharsToShow int
	LastCharsToShow  int
	RestrictionType  MaskRestrictionType
	CharsToIgnore    []rune
	XMLTag           string
	JSONKeys         []string
	KeepSameLength   bool
	prefixesToSkip   []string
}

// MaskXML masks parts of the inner text of tags from the input XML with replacement
func MaskXML(xml, maskChar string, toMasks []MaskData) string {
	maskChar = maskCharOrDefault(maskChar)

	doc, err := xmlquery.Parse(strings.NewReader(xml))
	if err != nil {
		return errors.New("invalid XML input", err).Error()
	}
	out := findTagAndMaskMulti(doc, maskChar, toMasks)

	return out
}

func findTagAndMaskMulti(doc *xmlquery.Node, maskChar string, toMasks []MaskData) string {
	for _, toMask := range toMasks {
		findTagAndMask(doc, maskChar, toMask)
	}
	return doc.OutputXML(true)
}

func findTagAndMask(doc *xmlquery.Node, maskChar string, toMask MaskData) {
	nodes := xmlquery.Find(doc, "//"+toMask.XMLTag)
	for _, node := range nodes {
		tagValue := node.InnerText()
		if strings.TrimSpace(tagValue) != "" {
			var nodeToUpdate *xmlquery.Node
			// get specific node that contains the value to update
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				nodeToUpdate = child
			}
			maskedVal := getMaskedValue(maskChar, tagValue, toMask)
			nodeToUpdate.Data = maskedVal
		}
	}
}

// MaskJSON mask parts of the json key paths value from the input json with replacement
//
// Example:
//
//	input = `
//	{
//		"first": "first value",
//		"second": {
//			"first": "1st of second",
//			"second": "second@wego.com",
//			"third": {
//				"first": "1st of second third",
//				"second": "2nd of second third",
//				"third": "3rd of second third",
//			}
//		}
//	}`
//	maskData := []logger.MaskData{
//		{
//			JSONKey:         []string{"first"},
//			FistCharsToShow: 3,
//			LastCharsToShow: 6,
//			KeepSameLength: true,
//		},
//		{
//			JSONKey:         []string{"second", "second"},
//			FistCharsToShow: 2,
//			LastCharsToShow: 3,
//			CharsToIgnore:   []rune{'@'},
//			KeepSameLength: true,
//		},
//		{
//			JSONKey:         []string{"second", "third", "first"},
//			FistCharsToShow: 3,
//			LastCharsToShow: 1,
//			KeepSameLength: true,
//		},
//	}
//	MaskJSON(input, "!", maskData) will return
//	{
//		"first": "fir!! value",
//		"second": {
//			"first": "1st of second",
//			"second": "se!!!!@!!!!!com",
//			"third": {
//				"first": "1st!!!!!!!!!!!!!!!d",
//				"second": "2nd of second third",
//				"third": "3rd of second third",
//			}
//		}
//	}
func MaskJSON(json, maskChar string, toMasks []MaskData) string {
	maskChar = maskCharOrDefault(maskChar)

	var p fastjson.Parser
	root, err := p.Parse(json)
	if err != nil {
		return err.Error()
	}

	for _, toMask := range toMasks {
		l := len(toMask.JSONKeys)
		switch {
		case l == 1:
			if exist := root.Exists(toMask.JSONKeys[0]); exist {
				// currently do not support masking for non-string values
				value := getJSONValue(root.Get(toMask.JSONKeys[0]))
				if value != "" {
					maskedVal := getMaskedValue(maskChar, value, toMask)
					replacement := fastjson.MustParse(`"` + maskedVal + `"`)
					root.Set(toMask.JSONKeys[0], replacement)
				}
			}
		case l > 1:
			if exist := root.Exists(toMask.JSONKeys...); exist {
				// get the parent obj then replace the value
				v := root.Get(toMask.JSONKeys[:l-1]...)

				// currently do not support masking for non-string values
				value := getJSONValue(v.Get(toMask.JSONKeys[l-1]))
				if value != "" {
					maskedVal := getMaskedValue(maskChar, value, toMask)
					replacement := fastjson.MustParse(`"` + maskedVal + `"`)
					v.Set(toMask.JSONKeys[l-1], replacement)
				}
			}
		}
	}

	out := root.MarshalTo([]byte{})
	return string(out)
}

func getJSONValue(jsonVal *fastjson.Value) string {
	val := ""
	if jsonVal != nil {
		bytes := jsonVal.GetStringBytes()
		if len(bytes) > 0 {
			val = string(bytes[:])
		}
	}

	return val
}

func maskCharOrDefault(maskChar string) string {
	if maskChar == "" {
		maskChar = defaultMaskChar
	}
	return maskChar
}

func getMaskedValue(maskChar, valueToReplace string, toMask MaskData) string {
	negativeCharsToShow := toMask.FirstCharsToShow < 0 || toMask.LastCharsToShow < 0
	if negativeCharsToShow || !willMask(valueToReplace, toMask.RestrictionType) {
		return valueToReplace
	}

	firstCharsToShow := toMask.FirstCharsToShow
	for _, prefix := range toMask.prefixesToSkip {
		if strings.HasPrefix(valueToReplace, prefix) {
			firstCharsToShow = len(prefix) + toMask.FirstCharsToShow
			break
		}
	}

	totalCharsToShow := firstCharsToShow + toMask.LastCharsToShow
	valueToReplaceLen := len(valueToReplace)
	lastIndexToShowStart := valueToReplaceLen - toMask.LastCharsToShow

	// check if need to mask
	repeatTimes := valueToReplaceLen - totalCharsToShow
	if repeatTimes <= 0 {
		return valueToReplace
	}

	if !toMask.KeepSameLength {
		return valueToReplace[:firstCharsToShow] + maskChar + valueToReplace[lastIndexToShowStart:]
	}

	valToMask := valueToReplace[firstCharsToShow:lastIndexToShowStart]
	var sb strings.Builder
	for _, c := range valToMask {
		// do not mask characters that should be ignored like '@'
		if collection.Contains(toMask.CharsToIgnore, c) {
			_, _ = sb.WriteRune(c)
		} else {
			_, _ = sb.WriteString(maskChar)
		}
	}

	replacement := sb.String()
	maskedVal := valueToReplace[:firstCharsToShow] + replacement + valueToReplace[lastIndexToShowStart:]

	return maskedVal
}

func willMask(valueToReplace string, restrictionType MaskRestrictionType) bool {
	if restrictionType == "" || valueToReplace == "" {
		return true
	}

	switch restrictionType {
	case MaskRestrictionTypeEmail:
		_, err := mail.ParseAddress(valueToReplace)
		return err == nil
	}

	return true
}
