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

// RedactJSON replaces value of key paths from the input JSON with replacement or defaultReplacement when replacement is empty
//
// Example:
//
// 		input = `
// 		{
// 			"first": "first value",
// 			"second": {
// 				"first": "1st of second",
// 				"second": "2nd of second",
// 				"third": {
// 					"first": "1st of second third",
// 					"second": "2nd of second third",
// 					"third": "3rd of second third",
// 				}
// 			}
// 		}`
// 		keys := [][]string{{"first"}, {"second", "second"}, {"second", "third", "first"}}
// 		RedactJSON(input, "Wego", keys) will return
// 		{
// 			"first": "Wego",
// 			"second": {
// 				"first": "1st of second",
// 				"second": "Wego",
// 				"third": {
// 					"first": "Wego",
// 					"second": "2nd of second third",
// 					"third": "3rd of second third",
// 				}
// 			}
// 		}
func RedactJSON(json, replacement string, keys [][]string) string {
	if replacement == "" {
		replacement = defaultReplacement
	}
	r := fastjson.MustParse(`"` + replacement + `"`)
	var p fastjson.Parser
	root, err := p.Parse(json)
	if err != nil {
		return err.Error()
	}
	for _, k := range keys {
		l := len(k)
		switch {
		case l == 1:
			if exist := root.Exists(k[0]); exist {
				root.Set(k[0], r)
			}
		case l > 1:
			if exist := root.Exists(k...); exist {
				// get the parent obj then replace the value
				v := root.Get(k[:l-1]...)
				v.Set(k[l-1], r)
			}
		}
	}

	out := root.MarshalTo([]byte{})
	return string(out)
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
