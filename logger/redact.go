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

The value's type does not matter: numbers, booleans, null, objects and arrays are
redacted like strings, at every depth. Every member matching a key is redacted, so
a duplicated key cannot leave one copy behind, and members duplicating a key
collapse into a single redacted member.

Pass CaseInsensitiveKeys when the input was bound with encoding/json, whose field
matching is case-insensitive while the key matching here is not by default.
*/
func RedactJSON(json, replacement string, keys [][]string, opts ...Option) string {
	replacement = replacementCharOrDefault(replacement)
	replacementValue := jsonString(replacement)
	options := buildJSONOptions(opts)

	var p fastjson.Parser
	root, err := p.Parse(json)
	if err != nil {
		return err.Error()
	}

	replace := func(*fastjson.Value) *fastjson.Value {
		return replacementValue
	}
	for _, toRedact := range keys {
		replaceLeaves(root, toRedact, options, replace)
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

/*
RedactFormURLEncoded replaces value of keys from the input form encoded string with replacement or defaultReplacement when replacement is empty.

Since input is form encoded string, keys would just be a simple array/list of keys to be redacted.
*/
func RedactFormURLEncoded(form string, replacement string, keys []string) string {
	replacement = replacementCharOrDefault(replacement)

	formData, err := url.ParseQuery(form)
	if err != nil {
		return form
	}

	for _, key := range keys {
		if len(key) >= 1 {
			if values, exists := formData[key]; exists {
				for i := range values {
					formData[key][i] = replacement
				}
			}
		}
	}

	return formData.Encode()
}

// RedactQueryParams replaces sensitive query parameters in the URL
func RedactQueryParams(rawQueryParams, replacement string, sensitiveParams []string) string {
	replacement = replacementCharOrDefault(replacement)

	queryParams, err := url.ParseQuery(rawQueryParams)
	if err != nil {
		return rawQueryParams
	}

	if len(queryParams) == 0 {
		return rawQueryParams
	}

	masked := false

	for _, param := range sensitiveParams {
		if queryParams.Has(param) {
			queryParams.Set(param, replacement)
			masked = true
		}
	}

	if masked {
		return queryParams.Encode()
	}

	return rawQueryParams
}
