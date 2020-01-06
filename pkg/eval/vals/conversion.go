package vals

import (
	"errors"
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

type wrongType struct {
	wantKind string
	gotKind  string
}

func (err wrongType) Error() string {
	return fmt.Sprintf("wrong type: need %s, got %s", err.wantKind, err.gotKind)
}

type cannotParseAs struct {
	want string
	repr string
}

func (err cannotParseAs) Error() string {
	return fmt.Sprintf("cannot parse as %s: %s", err.want, err.repr)
}

var (
	errMustBeString       = errors.New("must be string")
	errMustBeValidUTF8    = errors.New("must be valid UTF-8")
	errMustHaveSingleRune = errors.New("must have a single rune")
	errMustBeNumber       = errors.New("must be number")
	errMustBeInteger      = errors.New("must be integer")
)

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
		ptrType := TypeOf(ptr)
		if ptrType.Kind() != reflect.Ptr {
			return fmt.Errorf("internal bug: need pointer to scan to, got %T", ptr)
		}
		dstType := ptrType.Elem()
		if !TypeOf(src).AssignableTo(dstType) {
			return wrongType{Kind(reflect.Zero(dstType).Interface()), Kind(src)}
		}
		ValueOf(ptr).Elem().Set(ValueOf(src))
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
	case rune:
		return string(a)
	default:
		return a
	}
}

func elvToFloat(arg interface{}) (float64, error) {
	switch arg := arg.(type) {
	case float64:
		return arg, nil
	case string:
		f, err := strconv.ParseFloat(arg, 64)
		if err == nil {
			return f, nil
		}
		i, err := strconv.ParseInt(arg, 0, 64)
		if err == nil {
			return float64(i), err
		}
		return 0, cannotParseAs{"number", Repr(arg, -1)}
	default:
		return 0, errMustBeNumber
	}
}

func elvToInt(arg interface{}) (int, error) {
	switch arg := arg.(type) {
	case float64:
		i := int(arg)
		if float64(i) != arg {
			return 0, errMustBeInteger
		}
		return i, nil
	case string:
		num, err := strconv.ParseInt(arg, 0, 0)
		if err == nil {
			return int(num), nil
		}
		return 0, cannotParseAs{"integer", Repr(arg, -1)}
	default:
		return 0, errMustBeInteger
	}
}

func elvToRune(arg interface{}) (rune, error) {
	ss, ok := arg.(string)
	if !ok {
		return -1, errMustBeString
	}
	s := ss
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return -1, errMustBeValidUTF8
	}
	if size != len(s) {
		return -1, errMustHaveSingleRune
	}
	return r, nil
}
