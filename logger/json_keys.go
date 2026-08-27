package logger

import (
	"encoding/json"
	"strings"

	"github.com/valyala/fastjson"
)

// Option configures how RedactJSON and MaskJSON resolve JSON key paths.
type Option func(*jsonOptions)

type jsonOptions struct {
	caseInsensitiveKeys bool
}

/*
CaseInsensitiveKeys matches JSON object keys without regard to case, so one
lower-case entry in the key path covers every casing a caller can send.

encoding/json binds struct fields case-insensitively: a body of
`{"Customer":{"PAN":"..."}}` reaches the handler and is persisted exactly like
`{"customer":{"pan":"..."}}`. Masking runs on the raw body through fastjson,
which compares keys byte for byte, so without this option only the exact casing
is redacted and the log silently diverges from what the handler accepted.

Pass it whenever the input is a request or response body bound with encoding/json.
*/
func CaseInsensitiveKeys() Option {
	return func(o *jsonOptions) { o.caseInsensitiveKeys = true }
}

func buildJSONOptions(opts []Option) jsonOptions {
	var o jsonOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	return o
}

/*
replaceLeaves walks keys from v and offers every position the path resolves to
to replace, which returns the value to store there or nil to leave it as it is.

Three things differ from a plain fastjson Get/Set walk, each of them a silent
no-op before:

  - Every member matching a key is visited, not only the first. A JSON object may
    repeat a key; fastjson keeps both copies, and Get and Set only ever reach the
    first, so the duplicate survives into the logged bytes.
  - The value's type is not consulted. Numbers, booleans, null, objects and arrays
    are offered to replace exactly like strings.
  - `[]` fans out over array items at any position including the last, where every
    item of the array is offered to replace.
*/
func replaceLeaves(v *fastjson.Value, keys []string, o jsonOptions, replace func(*fastjson.Value) *fastjson.Value) {
	if v == nil || len(keys) == 0 {
		return
	}

	if keys[0] == arrayKey {
		items := v.GetArray()
		if len(keys) == 1 {
			for i, item := range items {
				if newValue := replace(item); newValue != nil {
					v.SetArrayItem(i, newValue)
				}
			}
			return
		}
		for _, item := range items {
			replaceLeaves(item, keys[1:], o, replace)
		}
		return
	}

	obj, err := v.Object()
	if err != nil {
		// Not an object, so it holds no members to match. Arrays only take part
		// through an explicit `[]` in the key path, as before.
		return
	}

	// Members are counted before anything is written: replacing one while Visit
	// walks the same object is not safe, and duplicates have to be counted before
	// the first of them is removed. The overwhelmingly common answer is one
	// member, which is handled without collecting anything.
	var (
		matches   int
		firstKey  string
		firstItem *fastjson.Value
	)
	obj.Visit(func(k []byte, member *fastjson.Value) {
		if !keyMatches(k, keys[0], o) {
			return
		}
		matches++
		if matches == 1 {
			firstKey, firstItem = keys[0], member
			if o.caseInsensitiveKeys {
				firstKey = string(k)
			}
		}
	})

	switch matches {
	case 0:
		return
	case 1:
		if len(keys) > 1 {
			replaceLeaves(firstItem, keys[1:], o, replace)
			return
		}
		if newValue := replace(firstItem); newValue != nil {
			obj.Set(firstKey, newValue)
		}
		return
	}

	replaceDuplicatedLeaves(obj, keys, o, replace)
}

// replaceDuplicatedLeaves is replaceLeaves' path for an object carrying more than
// one member matching a key: the same key repeated, or several spellings of it
// under CaseInsensitiveKeys. Split out because it is the rare case and the only
// one that needs to hold the matches in memory.
func replaceDuplicatedLeaves(obj *fastjson.Object, keys []string, o jsonOptions, replace func(*fastjson.Value) *fastjson.Value) {
	var spellings []string
	members := map[string][]*fastjson.Value{}
	obj.Visit(func(k []byte, member *fastjson.Value) {
		if !keyMatches(k, keys[0], o) {
			return
		}
		key := string(k)
		if _, seen := members[key]; !seen {
			spellings = append(spellings, key)
		}
		members[key] = append(members[key], member)
	})

	for _, key := range spellings {
		matched := members[key]

		if len(keys) > 1 {
			for _, member := range matched {
				replaceLeaves(member, keys[1:], o, replace)
			}
			continue
		}

		// Members duplicating a key collapse into one. encoding/json binds the
		// last of them, so that is the copy whose value decides the replacement.
		newValue := replace(matched[len(matched)-1])
		if newValue == nil {
			continue
		}
		if len(matched) > 1 {
			// Set replaces only the first member carrying the key, so every copy
			// has to be deleted first. Del removes one occurrence per call. This
			// moves the key to the end of the object; a key present once keeps
			// its position.
			for range matched {
				obj.Del(key)
			}
		}
		obj.Set(key, newValue)
	}
}

func keyMatches(key []byte, want string, o jsonOptions) bool {
	if o.caseInsensitiveKeys {
		return strings.EqualFold(string(key), want)
	}
	// The compiler turns this comparison into a byte compare, with no conversion.
	return string(key) == want
}

// jsonString builds a JSON string value holding s.
//
// fastjson.MustParse on a hand-quoted string panics on any value carrying a quote
// or a backslash, which a partially masked name or address can still hold: masking
// keeps FirstCharsToShow and LastCharsToShow verbatim.
func jsonString(s string) *fastjson.Value {
	quoted, err := json.Marshal(s)
	if err != nil {
		return fastjson.MustParse(`""`)
	}
	return fastjson.MustParse(string(quoted))
}

// scalarText renders a JSON scalar the way masking needs to read it: a string as
// its unquoted content, a number or a boolean as its literal. Objects, arrays and
// null carry no text to mask and report false.
func scalarText(v *fastjson.Value) (string, bool) {
	if v == nil {
		return "", false
	}

	switch v.Type() {
	case fastjson.TypeString:
		return string(v.GetStringBytes()), true
	case fastjson.TypeNumber, fastjson.TypeTrue, fastjson.TypeFalse:
		return v.String(), true
	default:
		return "", false
	}
}
