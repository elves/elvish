package eval

import (
	"fmt"
	"reflect"
	"strconv"
	"unicode/utf8"
)

// Conversion between native and Elvish values.
//
// Elvish uses native Go types most of the time - string, bool, hashmap.Map,
// vector.Vector, etc., and there is no need for any conversions. There are some
// exceptions, for instance the numerical types int and float64: in native Go
// code they are distinct types, but Elvish uses string for all numbers.
// Similarly, Elvish uses string to represent runes. In all cases, there is a
// many-to-one relationship between Go types and Elvish types.
//
// Conversion from Go value to Elvish value can happen without knowing the
// destination type: int, float64 and rune values are converted to strings, and
// values of other types remain unchanged. The opposite is not true: the native
// Go value corresponding to the Elvish value "1" can be string("1"), int(1),
// float64(1.0) or rune('1'). Conversion in this direction depends on the
// destination type.

func toFloat(arg interface{}) (float64, error) {
	if _, ok := arg.(string); !ok {
		return 0, fmt.Errorf("must be string")
	}
	s := arg.(string)
	num, err := strconv.ParseFloat(s, 64)
	if err != nil {
		num, err2 := strconv.ParseInt(s, 0, 64)
		if err2 != nil {
			return 0, err
		}
		return float64(num), nil
	}
	return num, nil
}

func floatToString(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}

func toInt(arg interface{}) (int, error) {
	arg, ok := arg.(string)
	if !ok {
		return 0, fmt.Errorf("must be string")
	}
	num, err := strconv.ParseInt(arg.(string), 0, 0)
	if err != nil {
		return 0, err
	}
	return int(num), nil
}

func toRune(arg interface{}) (rune, error) {
	ss, ok := arg.(string)
	if !ok {
		return -1, fmt.Errorf("must be string")
	}
	s := ss
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return -1, fmt.Errorf("string is not valid UTF-8")
	}
	if size != len(s) {
		return -1, fmt.Errorf("string has multiple runes")
	}
	return r, nil
}

// scanValueToGo converts Value to Go data, depending on the type of the
// destination.
func scanValueToGo(src interface{}, ptr interface{}) {
	v, err := convertArg(src, reflect.TypeOf(ptr).Elem())
	maybeThrow(err)
	reflect.Indirect(reflect.ValueOf(ptr)).Set(reflect.ValueOf(v))
}
