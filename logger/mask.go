package logger

import (
	"net/mail"
	"net/url"
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

/*
MaskJSON mask parts of the json key paths value from the input json with replacement

For nested arrays, use `[]` as the key.
*/
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
			arrIndices := []int{}
			for i, key := range toMask.JSONKeys {
				if key == arrayKey {
					arrIndices = append(arrIndices, i)
				}
			}
			// `root.Exists(toMask.JSONKeys...)` will not work when there are array indices (more than 1 "[]"), so we
			// should also try to set `exist` to `true` if the caller inputs array indices.
			exist := root.Exists(toMask.JSONKeys...) || len(arrIndices) > 0

			if exist {
				if len(arrIndices) > 0 {
					maskArrayRecursive(root, toMask.JSONKeys, maskChar, toMask)
				} else {
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
	}

	out := root.MarshalTo([]byte{})
	return string(out)
}

func maskArrayRecursive(obj *fastjson.Value, keys []string, maskChar string, toMask MaskData) {
	if len(keys) == 0 || obj == nil {
		return
	}

	arrayAtStart := keys[0] == arrayKey
	singleProperty := len(keys) == 1
	arrayAtEnd := len(keys) >= 2 && keys[len(keys)-1] == arrayKey

	switch {
	case arrayAtStart:
		maskArrayItems(obj, keys[1:], maskChar, toMask)
	case singleProperty:
		maskSingleProperty(obj, keys[0], maskChar, toMask)
	case arrayAtEnd:
		maskArrayAtPath(obj, keys, maskChar, toMask)
	default:
		navigateDeeper(obj, keys, maskChar, toMask)
	}
}

// maskArrayItems handles the pattern ["[]", "property", ...] where array key comes first.
// It iterates through array items and recursively masks each item with remaining keys.
func maskArrayItems(obj *fastjson.Value, remainingKeys []string, maskChar string, toMask MaskData) {
	arr := obj.GetArray()
	for _, item := range arr {
		maskArrayRecursive(item, remainingKeys, maskChar, toMask)
	}
}

// maskSingleProperty handles the pattern ["property"] for final property masking.
// It masks the specified property directly on the object.
func maskSingleProperty(obj *fastjson.Value, key string, maskChar string, toMask MaskData) {
	value := getJSONValue(obj.Get(key))
	if value != "" {
		maskedVal := getMaskedValue(maskChar, value, toMask)
		setMaskedValue(obj, key, maskedVal)
	}
}

// maskArrayAtPath handles patterns ["property", "[]"] or ["prop1", "prop2", "[]"]
// where array key comes at the end. It navigates to the property path then masks
// all contents of the array found there.
func maskArrayAtPath(obj *fastjson.Value, keys []string, maskChar string, toMask MaskData) {
	propertyKeys := keys[:len(keys)-1]
	nestedObj := obj.Get(propertyKeys...)

	if nestedObj != nil && nestedObj.Type() == fastjson.TypeArray {
		maskArrayContents(nestedObj, maskChar, toMask)
	}
}

// navigateDeeper handles the pattern ["property", ...] for navigating deeper
// into object structure. It gets the nested object and continues recursively.
func navigateDeeper(obj *fastjson.Value, keys []string, maskChar string, toMask MaskData) {
	nestedObj := obj.Get(keys[0])
	if nestedObj != nil {
		maskArrayRecursive(nestedObj, keys[1:], maskChar, toMask)
	}
}

// maskArrayContents masks all string values within an array, handling both
// direct string values and nested objects/arrays recursively.
func maskArrayContents(arrayObj *fastjson.Value, maskChar string, toMask MaskData) {
	arr := arrayObj.GetArray()
	for i, item := range arr {
		if item.Type() == fastjson.TypeString {
			value := getJSONValue(item)
			if value != "" {
				maskedVal := getMaskedValue(maskChar, value, toMask)
				setArrayItemMaskedValue(arrayObj, i, maskedVal)
			}
		} else {
			maskAllStringValues(item, maskChar, toMask)
		}
	}
}

// setMaskedValue sets a masked string value on an object property.
func setMaskedValue(obj *fastjson.Value, key string, maskedVal string) {
	replacement := fastjson.MustParse(`"` + maskedVal + `"`)
	obj.Set(key, replacement)
}

// setArrayItemMaskedValue sets a masked string value at a specific array index.
func setArrayItemMaskedValue(arrayObj *fastjson.Value, index int, maskedVal string) {
	replacement := fastjson.MustParse(`"` + maskedVal + `"`)
	arrayObj.SetArrayItem(index, replacement)
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

func maskAllStringValues(obj *fastjson.Value, maskChar string, toMask MaskData) {
	if obj == nil || obj.Type() != fastjson.TypeArray {
		return
	}

	arr := obj.GetArray()
	for i, item := range arr {
		if item.Type() == fastjson.TypeString {
			value := getJSONValue(item)
			if value != "" {
				maskedVal := getMaskedValue(maskChar, value, toMask)
				replacement := fastjson.MustParse(`"` + maskedVal + `"`)
				obj.SetArrayItem(i, replacement)
			}
		} else {
			// Recursively handle nested arrays
			maskAllStringValues(item, maskChar, toMask)
		}
	}
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

// MaskFormURLEncoded mask parts of the key paths values from the input form encoded string with replacement
func MaskFormURLEncoded(form string, maskChar string, toMasks []MaskData) string {
	maskChar = maskCharOrDefault(maskChar)

	r, err := url.ParseQuery(form)
	if err != nil {
		return form
	}

	for _, toMask := range toMasks {
		for _, key := range toMask.JSONKeys {
			if values, exists := r[key]; exists {
				for i := range values {
					r[key][i] = getMaskedValue(maskChar, values[i], toMask)
				}
			}
		}
	}

	return r.Encode()
}

// MaskQueryParams masks sensitive query parameters in the URL with replacement
func MaskQueryParams(rawQueryParams, maskChar string, toMasks []MaskData) string {
	maskChar = maskCharOrDefault(maskChar)

	queryParams, err := url.ParseQuery(rawQueryParams)
	if err != nil {
		return rawQueryParams
	}

	if len(queryParams) == 0 {
		return rawQueryParams
	}

	masked := false

	for _, toMask := range toMasks {
		for _, key := range toMask.JSONKeys {
			if queryParams.Has(key) {
				values := queryParams[key]
				for i := range values {
					queryParams[key][i] = getMaskedValue(maskChar, values[i], toMask)
				}
				masked = true
			}
		}
	}

	if masked {
		return queryParams.Encode()
	}

	return rawQueryParams
}
