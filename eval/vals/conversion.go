package vals

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

// ScanToGo converts an Elvish value to a Go value. the pointer points to. It
// uses the type of the pointer to determine the destination type, and puts the
// converted value in the location the pointer points to. Conversion only
// happens when the destination type is int, float64 or rune; in other cases,
// this function just checks that the source value is already assignable to the
// destination.
func ScanToGo(src interface{}, ptr interface{}) error {
	switch ptr := ptr.(type) {
	case *int:
		i, err := elvToInt(src)
		if err == nil {
			*ptr = i
		}
		return err
	case *float64:
		f, err := elvToFloat(src)
		if err == nil {
			*ptr = f
		}
		return err
	case *rune:
		r, err := elvToRune(src)
		if err == nil {
			*ptr = r
		}
		return err
	case Scanner:
		return ptr.ScanElvish(src)
	default:
		// Do a generic `*ptr = src` via reflection
		ptrType := reflect.TypeOf(ptr)
		if ptrType.Kind() != reflect.Ptr {
			return fmt.Errorf("need pointer to scan into, got %T", ptr)
		}
		dstType := ptrType.Elem()
		if !reflect.TypeOf(src).AssignableTo(dstType) {
			return fmt.Errorf("need %s, got %s",
				Kind(reflect.Zero(dstType).Interface()), Kind(src))
		}
		reflect.ValueOf(ptr).Elem().Set(reflect.ValueOf(src))
		return nil
	}
}

// Scanner is implemented by types that can scan an Elvish value into itself.
type Scanner interface {
	ScanElvish(interface{}) error
}

// FromGo converts a Go value to an Elvish value. Conversion happens when the
// argument is int, float64 or rune (this is consistent with ScanToGo). In other
// cases, this function just returns the argument.
func FromGo(a interface{}) interface{} {
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
