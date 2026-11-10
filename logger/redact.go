package logger

import (
	"net/url"
	"slices"
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
		return invalidJSONText
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

/*
RedactURLQueryParams replaces the values of sensitiveParams in the query of rawURL, and in a query
inside its fragment (e.g. "#/route?key=value" or "#key=value"; the part of a fragment before
its "?" is redacted as params too when it holds a "="), with replacement or defaultReplacement when replacement is empty. The
replacement is query-escaped.

Unlike RedactQueryParams, it takes a whole URL and keeps every other byte as is, including the
order and encoding of the other params. Both "&" and ";" separate params, and keys are compared
after URL-decoding, so "na%6De" matches "name". Keys are case-sensitive. A matched value runs to
the next "&", since a ";" may belong to it; any ";"-separated params after it are redacted too.

A query holding a malformed percent-encoding is replaced whole, so a value that cannot be parsed
is never returned as is. rawURL is returned unchanged, without allocating, when nothing matches;
the replacement is only escaped once a param matches.
*/
func RedactURLQueryParams(rawURL, replacement string, sensitiveParams []string) string {
	if len(sensitiveParams) == 0 {
		return rawURL
	}

	base, fragment, hasFragment := strings.Cut(rawURL, "#")
	path, query, hasQuery := strings.Cut(base, "?")
	var fragmentPath, fragmentQuery string
	var hasFragmentQuery, fragmentIsQuery bool
	if hasFragment {
		fragmentPath, fragmentQuery, hasFragmentQuery = strings.Cut(fragment, "?")
		if !hasFragmentQuery && strings.Contains(fragment, "=") {
			fragmentQuery, hasFragmentQuery, fragmentIsQuery = fragment, true, true
		}
	}
	if !hasQuery && !hasFragmentQuery {
		return rawURL
	}

	replacement = replacementCharOrDefault(replacement)

	var queryChanged, fragmentChanged bool
	if hasQuery {
		query, queryChanged = redactQuery(query, replacement, sensitiveParams)
	}
	if hasFragmentQuery {
		fragmentQuery, fragmentChanged = redactQuery(fragmentQuery, replacement, sensitiveParams)
	}
	var fragmentPathChanged bool
	if hasFragmentQuery && !fragmentIsQuery && strings.Contains(fragmentPath, "=") {
		fragmentPath, fragmentPathChanged = redactQuery(fragmentPath, replacement, sensitiveParams)
	}
	if !queryChanged && !fragmentChanged && !fragmentPathChanged {
		return rawURL
	}

	var sb strings.Builder
	sb.Grow(len(rawURL))
	if hasQuery {
		_, _ = sb.WriteString(path)
		_ = sb.WriteByte('?')
		_, _ = sb.WriteString(query)
	} else {
		_, _ = sb.WriteString(base)
	}
	if hasFragment {
		_ = sb.WriteByte('#')
		switch {
		case fragmentIsQuery:
			_, _ = sb.WriteString(fragmentQuery)
		case hasFragmentQuery:
			_, _ = sb.WriteString(fragmentPath)
			_ = sb.WriteByte('?')
			_, _ = sb.WriteString(fragmentQuery)
		default:
			_, _ = sb.WriteString(fragment)
		}
	}
	return sb.String()
}

// redactQuery replaces the values of params in a raw query with the query-escaped replacement,
// or the whole query when it holds a malformed percent-encoding. It reports whether the query
// changed.
func redactQuery(query, replacement string, params []string) (string, bool) {
	if !validQueryEscapes(query) {
		return url.QueryEscape(replacement), true
	}

	var sb strings.Builder
	copied := 0
	for start := 0; start <= len(query); {
		end := strings.IndexAny(query[start:], "&;")
		if end < 0 {
			end = len(query)
		} else {
			end += start
		}

		key, _, hasValue := strings.Cut(query[start:end], "=")
		if hasValue && isSensitiveQueryKey(key, params) {
			// A ";" may also be part of the value, so a matched value runs to the
			// next "&": over-redacting a ";"-separated param beats logging a tail.
			if amp := strings.IndexByte(query[end:], '&'); amp < 0 {
				end = len(query)
			} else {
				end += amp
			}
			if copied == 0 {
				replacement = url.QueryEscape(replacement)
				sb.Grow(len(query))
			}
			valueStart := start + len(key) + 1
			_, _ = sb.WriteString(query[copied:valueStart])
			_, _ = sb.WriteString(replacement)
			copied = end
		}
		start = end + 1
	}

	if copied == 0 {
		return query, false
	}
	_, _ = sb.WriteString(query[copied:])
	return sb.String(), true
}

// isSensitiveQueryKey reports whether key, once query-unescaped, is one of params. key must hold
// only valid percent-encodings.
func isSensitiveQueryKey(key string, params []string) bool {
	if !strings.ContainsAny(key, "%+") {
		return slices.Contains(params, key)
	}
	for _, param := range params {
		if unescapedQueryKeyEquals(key, param) {
			return true
		}
	}
	return false
}

// unescapedQueryKeyEquals compares a query-escaped key with want without allocating the
// unescaped key.
func unescapedQueryKeyEquals(key, want string) bool {
	n := 0
	for i := 0; i < len(key); i++ {
		c := key[i]
		switch c {
		case '+':
			c = ' '
		case '%':
			c = unhex(key[i+1])<<4 | unhex(key[i+2])
			i += 2
		}
		if n >= len(want) || want[n] != c {
			return false
		}
		n++
	}
	return n == len(want)
}

func unhex(c byte) byte {
	switch {
	case c >= 'a':
		return c - 'a' + 10
	case c >= 'A':
		return c - 'A' + 10
	default:
		return c - '0'
	}
}

// validQueryEscapes reports whether every "%" in s starts a valid percent-encoding, the only
// input url.QueryUnescape rejects.
func validQueryEscapes(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] != '%' {
			continue
		}
		if i+2 >= len(s) || !isHex(s[i+1]) || !isHex(s[i+2]) {
			return false
		}
		i += 2
	}
	return true
}

func isHex(c byte) bool {
	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')
}
