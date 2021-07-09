package vals

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"unicode/utf8"
)

// Conversion between native and Elvish values.
//
// Elvish uses native Go types most of the time - string, bool, hashmap.Map,
// vector.Vector, etc., and there is no need for any conversions. There are some
// exceptions, for instance int and rune, since Elvish currently lacks integer
// types.
//
// There is a many-to-one relationship between Go types and Elvish types. A
// Go value can always be converted to an Elvish value unambiguously, but to
// convert an Elvish value into a Go value one must know the destination type
// first. For example, all of the Go values int(1), rune('1') and string("1")
// convert to Elvish "1"; conversely, Elvish "1" may be converted to any of the
// aforementioned three possible values, depending on the destination type.
//
// In future, Elvish may gain distinct types for integers and characters, making
// the examples above unnecessary; however, the conversion logic may not
// entirely go away, as there might always be some mismatch between Elvish's
// type system and Go's.

type WrongType struct {
	WantKind string
	GotKind  string
}

func (err WrongType) Error() string {
	return fmt.Sprintf("wrong type: need %s, got %s", err.WantKind, err.GotKind)
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

// ScanToGo converts an Elvish value to a Go value that the pointer refers to. It
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
		n, err := elvToNum(src)
		if err == nil {
			*ptr = ConvertToFloat64(n)
		}
		return err
	case *Num:
		n, err := elvToNum(src)
		if err == nil {
			*ptr = n
		}
		return err
	case *rune:
		r, err := elvToRune(src)
		if err == nil {
			*ptr = r
		}
		return err
	default:
		// Do a generic `*ptr = src` via reflection
		ptrType := TypeOf(ptr)
		if ptrType.Kind() != reflect.Ptr {
			// TODO: Change this to a panic or explain why it shouldn't panic given this should be a
			// "can't happen" situation.
			return fmt.Errorf("internal bug: need pointer to scan to, got %T", ptr)
		}
		dstType := ptrType.Elem()
		if !TypeOf(src).AssignableTo(dstType) {
			return WrongType{Kind(ptr), Kind(src)}
		}
		ValueOf(ptr).Elem().Set(ValueOf(src))
		return nil
	}
}

// FromGo converts a Go value to an Elvish value. Most types are returned as
// is, but exact numerical types are normalized to one of int, *big.Int and
// *big.Rat, using the small representation that can hold the value, and runes
// are converted to strings.
func FromGo(a interface{}) interface{} {
	switch a := a.(type) {
	case *big.Int:
		return NormalizeBigInt(a)
	case *big.Rat:
		return NormalizeBigRat(a)
	case rune:
		return string(a)
	default:
		return a
	}
}

func elvToInt(arg interface{}) (int, error) {
	switch arg := arg.(type) {
	case int:
		return arg, nil
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

func elvToNum(arg interface{}) (Num, error) {
	switch arg := arg.(type) {
	case int, *big.Int, *big.Rat, float64:
		return arg, nil
	case string:
		n := ParseNum(arg)
		if n == nil {
			return 0, cannotParseAs{"number", Repr(arg, -1)}
		}
		return n, nil
	default:
		return 0, errMustBeNumber
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
