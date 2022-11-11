package to

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	nullValueRegex = regexp.MustCompile(`\"[^'"]*\":(\t+|\s+)?null(,)?(\r|\n|\r\n)?`)
)

// String returns a string representation of the given value.
func String[T any](v T) string {
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Bool:
		if val.Bool() {
			return "true"
		}
		return "false"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(val.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(val.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(val.Float(), 'f', -1, 64)
	case reflect.Complex64, reflect.Complex128:
		c := val.Complex()
		return strconv.FormatFloat(real(c), 'f', -1, 64) + "+" +
			strconv.FormatFloat(imag(c), 'f', -1, 64) + "i"
	case reflect.String:
		return val.String()
	case reflect.Pointer, reflect.Interface:
		if val.IsNil() {
			return "nil"
		}
		return String(val.Elem().Interface())
	case reflect.UnsafePointer:
		return "unsafe.Pointer" + " at 0x" + strconv.FormatUint(uint64(val.Pointer()), 16)
	case reflect.Func:
		return val.Type().String() + " at 0x" + strconv.FormatUint(uint64(val.Pointer()), 16)
	case reflect.Chan:
		return val.Type().String() + " at 0x" + strconv.FormatUint(uint64(val.Pointer()), 16) +
			" with " + strconv.FormatInt(int64(val.Len()), 10) + " elements"
	case reflect.Array, reflect.Slice, reflect.Map, reflect.Struct:
		fallthrough
	default:
		bytes, _ := json.Marshal(v)
		return strings.ReplaceAll(nullValueRegex.ReplaceAllString(string(bytes), ""), `,}`, "}")
	}
}
