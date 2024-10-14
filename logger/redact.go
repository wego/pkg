package logger

import (
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/valyala/fastjson"
	"github.com/wego/pkg/errors"
)

// RedactXML replaces inner text of tags from the input XML with replacement or defaultReplacement when replacement is empty
func RedactXML(xml, replacement string, tags []string) string {
	if replacement == "" {
		replacement = defaultReplacement
	}
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

With the following JSON:

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

And the following usage:

	keys := [][]string{{"first"}, {"second", "second"}, {"second", "third", "first"}, {"third", "[]", "value"}}
	result := RedactJSON(input, "Wego", keys)
	fmt.Println(result)

We get the following output:

	{
	  "first": "Wego",
	  "second": {
	    "first": "1st of second",
	    "second": "Wego",
	    "third": {
	      "first": "Wego",
	      "second": "2nd of second third",
	      "third": "3rd of second third"
	    }
	  },
	  "third": [
	    { "value": "Wego" }
	  ]
	}
*/
func RedactJSON(json, replacement string, keys [][]string) string {
	if replacement == "" {
		replacement = defaultReplacement
	}
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
