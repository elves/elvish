package eval

import (
	"fmt"
	"reflect"
	"strconv"
	"unicode/utf8"

	"github.com/elves/elvish/eval/types"
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

var (
	intType   = reflect.TypeOf(int(0))
	floatType = reflect.TypeOf(float64(0))
	runeType  = reflect.TypeOf(rune(0))
)

// elvToGo converts an Elvish value to a Go value of the specified type.
// Conversion happens for a whitelist of target types; for other types, this
// function just checks whether the source value is already assignable to the
// target type and returns the source value.
func elvToGo(arg interface{}, typ reflect.Type) (interface{}, error) {
	switch typ {
	case intType:
		i, err := elvToInt(arg)
		return i, err
	case floatType:
		f, err := elvToFloat(arg)
		return f, err
	case runeType:
		r, err := elvToRune(arg)
		return r, err
	default:
		if reflect.TypeOf(arg).AssignableTo(typ) {
			return arg, nil
		}
		return nil, fmt.Errorf("need %s, got %s",
			types.Kind(reflect.Zero(typ).Interface()), types.Kind(arg))
	}
}

// scanElvToGo converts an Elvish value to a Go value, putting the result in a
// pointer.
func scanElvToGo(src interface{}, ptr interface{}) error {
	v, err := elvToGo(src, reflect.TypeOf(ptr).Elem())
	if err != nil {
		return err
	}
	reflect.Indirect(reflect.ValueOf(ptr)).Set(reflect.ValueOf(v))
	return nil
}

// mustScanElvToGo is like scanElvToGo, except that errors are turned into
// exceptions.
func mustScanElvToGo(src interface{}, ptr interface{}) {
	maybeThrow(scanElvToGo(src, ptr))
}

// goToElv a Go value to an Elvish value.
func goToElv(a interface{}) interface{} {
	switch a := a.(type) {
	case int:
		return strconv.Itoa(a)
	case float64:
		return strconv.FormatFloat(a, 'g', -1, 64)
	case rune:
		return string(a)
	default:
		return a
	}
}

func elvToFloat(arg interface{}) (float64, error) {
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

func floatToElv(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}

func elvToInt(arg interface{}) (int, error) {
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

func elvToRune(arg interface{}) (rune, error) {
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
