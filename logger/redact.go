package logger

import (
	"net/url"
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/valyala/fastjson"
	"github.com/wego/pkg/errors"
)

// RedactXML replaces inner text of tags from the input XML with replacement or defaultReplacement when replacement is empty
func RedactXML(xml, replacement string, tags []string) string {
	replacement = replacementCharOrDefault(replacement)
	doc, err := xmlquery.Parse(strings.NewReader(xml))
	if err != nil {
		return errors.New("invalid XML input", err).Error()
	}
	text := findTextMulti(doc, tags)
	out := xml
	for _, t := range text {
		out = strings.ReplaceAll(out, t, replacement)
	}
	return out
}

/*
RedactJSON replaces value of key paths from the input JSON with replacement or defaultReplacement when replacement is empty.

For nested arrays, use `[]` as the key.
*/
func RedactJSON(json, replacement string, keys [][]string) string {
	replacement = replacementCharOrDefault(replacement)
	replacementValue := fastjson.MustParse(`"` + replacement + `"`)
	var p fastjson.Parser
	root, err := p.Parse(json)
	if err != nil {
		return err.Error()
	}
	for _, toRedact := range keys {
		l := len(toRedact)
		switch {
		case l == 1:
			if exist := root.Exists(toRedact[0]); exist {
				root.Set(toRedact[0], replacementValue)
			}
		case l > 1:
			arrIndices := []int{}
			for i, key := range toRedact {
				if key == arrayKey {
					arrIndices = append(arrIndices, i)
				}
			}
			// `root.Exists(toMask.JSONKeys...)` will not work when there are array indices (more than 1 "[]"), so we
			// should also try to set `exist` to `true` if the caller inputs array indices.
			exist := root.Exists(toRedact...) || len(arrIndices) > 0

			if exist {
				if len(arrIndices) > 0 {
					redactArrayRecursive(root, toRedact, replacementValue)
				} else {
					// get the parent obj then replace the value
					v := root.Get(toRedact[:l-1]...)

					// currently do not support masking for non-string values
					value := getJSONValue(v.Get(toRedact[l-1]))
					if value != "" {
						v.Set(toRedact[l-1], replacementValue)
					}
				}
			}
		}
	}

	out := root.MarshalTo([]byte{})
	return string(out)
}

func replacementCharOrDefault(replacement string) string {
	if replacement == "" {
		return defaultReplacement
	}
	return replacement
}

func redactArrayRecursive(obj *fastjson.Value, keys []string, replacementValue *fastjson.Value) {
	if len(keys) == 0 || obj == nil || replacementValue == nil {
		return
	}

	if keys[0] == arrayKey {
		arr := obj.GetArray()
		for _, item := range arr {
			redactArrayRecursive(item, keys[1:], replacementValue)
		}
	} else if len(keys) == 1 {
		value := getJSONValue(obj.Get(keys[0]))
		if value != "" {
			obj.Set(keys[0], replacementValue)
		}
	} else {
		nestedObj := obj.Get(keys[0])
		if nestedObj != nil {
			redactArrayRecursive(nestedObj, keys[1:], replacementValue)
		}
	}
}

func findText(doc *xmlquery.Node, tag string) []string {
	text := []string{}
	nodes := xmlquery.Find(doc, "//"+tag)
	for _, node := range nodes {
		it := node.InnerText()
		if strings.TrimSpace(it) != "" {
			text = append(text, node.InnerText())
		}
	}
	return text
}

func findTextMulti(doc *xmlquery.Node, tags []string) []string {
	text := []string{}
	for _, tag := range tags {
		text = append(text, findText(doc, tag)...)
	}
	return text
}

// RedactFormURLEncoded replaces value of keys from the input form encoded string with replacement or defaultReplacement when replacement is empty.
func RedactFormURLEncoded(form string, replacement string, keys [][]string) string {
	if form == "" || len(keys) == 0 {
		return form
	}

	r, err := url.ParseQuery(form)
	if err != nil {
		return form
	}

	replacement = replacementCharOrDefault(replacement)

	keyMap := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		if len(key) >= 1 {
			keyMap[key[0]] = struct{}{}
		}
	}

	// Single pass through values
	for key, values := range r {
		if _, exists := keyMap[key]; exists {
			for i := range values {
				r[key][i] = replacement
			}
		}
	}

	return r.Encode()
}
