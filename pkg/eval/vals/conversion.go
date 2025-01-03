package vals

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"

	"src.elv.sh/pkg/eval/errs"
)

// WrongType is returned by ScanToGo if the source value doesn't have a
// compatible type.
type WrongType struct {
	wantKind string
	gotKind  string
}

// Error implements the error interface.
func (err WrongType) Error() string {
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

// ScanToGo converts an Elvish value to a Go value, and stores it in *ptr. It
// panics if ptr is not a pointer.
//
//   - If ptr has type *int, *float64, *Num or *rune, it performs a suitable
//     conversion, and returns an error if the conversion fails.
//
//   - If ptr is a pointer to a field map (*M) and src is a map or another field
//     map, it converts each individual field recursively, failing if the
//     scanning of any field fails.
//
//     The set of keys in src must match the set of keys in M exactly. This
//     behavior can be changed by using [ScanToGoOpts] instead.
//
//   - In other cases, it tries to perform "*ptr = src" via reflection and
//     returns an error if the assignment can't be done.
//
// Strictly speaking, an Elvish value is always a Go values, and any Go value
// can potentially participate in Elvish. The distinction between "Go value" and
// "Elvish value" is a matter of how they are used:
//
//   - Go values are those returned by idiomatic Go functions, expected as
//     arguments, or in general more convenient to manipulate in
//     non-Elvish-specific Go code.
//
//   - Elvish values are those fully supported by the Elvish runtime, defined to
//     a large extent by standard operations such as [Index] in this package.
//
// The API of ScanToGo and the related [FromGo] is asymmetric. This is because
// the Elvish value system makes fewer distinctions:
//
//   - Elvish doesn't have a dedicated rune type and uses strings to represent
//     them.
//
//   - Elvish doesn't distinguish maps and field maps, but they have different
//     underlying Go types.
//
//   - Moreover, exact numbers in Elvish are normalized to the smallest type
//     that can represent it.
//
// As a result, while [FromGo] can always convert a Go value to an Elvish value
// unambiguously and successfully, ScanToGo can't do that in the opposite
// direction. For example, "1" may be converted into "1" or '1', depending on
// what the destination type is, and the process may fail. Thus ScanToGo takes
// the pointer to the destination as an argument, and returns an error.
func ScanToGo(src, ptr any) error {
	return ScanToGoOpts(src, ptr, 0)
}

// Options for [ScanToGoOpts].
type ScanOpt uint32

const (
	// When scanning a map into a field map, allow the map to not have some keys
	// of the field map.
	AllowMissingMapKey ScanOpt = 1 << iota
	// When scanning a map into a field map, allow the map to have keys that are
	// not in the field map.
	AllowExtraMapKey
)

// ScanToGoOpts is like [ScanToGo], but allows customization the behavior with
// the flag argument.
func ScanToGoOpts(src, ptr any, opt ScanOpt) error {
	switch ptr := ptr.(type) {
	case *int:
		return convAndStore(elvToInt, src, ptr)
	case *float64:
		n, err := elvToNum(src)
		if err == nil {
			*ptr = ConvertToFloat64(n)
		}
		return err
	case *Num:
		return convAndStore(elvToNum, src, ptr)
	case *rune:
		return convAndStore(elvToRune, src, ptr)
	default:
		dstType := reflect.TypeOf(ptr).Elem()
		// Attempt a simple assignment (*ptr = src) via reflection.
		if TypeOf(src).AssignableTo(dstType) {
			ValueOf(ptr).Elem().Set(ValueOf(src))
			return nil
		}
		// Allow using any(nil) (which the value of Elvish's $nil) as T(nil) for
		// any T whose zero value is spelt nil.
		if src == nil {
			switch dstType.Kind() {
			case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
				ValueOf(ptr).Elem().SetZero()
				return nil
			}
		}
		// Try to scan a field map.
		if keys := getFieldMapKeysT(reflect.TypeOf(ptr).Elem()); keys != nil {
			if _, ok := src.(Map); ok || IsFieldMap(src) {
				return scanFieldMapFromMap(src, ptr, keys, opt)
			}
		}
		// Return a suitable error.
		var dstKind string
		if dstType.Kind() == reflect.Interface {
			dstKind = "!!" + dstType.String()
		} else {
			dstKind = Kind(reflect.Zero(dstType).Interface())
		}
		return WrongType{dstKind, Kind(src)}
	}
}

func convAndStore[T any](conv func(any) (T, error), src any, ptr *T) error {
	v, err := conv(src)
	if err == nil {
		*ptr = v
	}
	return err
}

func elvToInt(arg any) (int, error) {
	switch arg := arg.(type) {
	case int:
		return arg, nil
	case string:
		num, err := strconv.ParseInt(arg, 0, 0)
		if err == nil {
			return int(num), nil
		}
		return 0, cannotParseAs{"integer", ReprPlain(arg)}
	default:
		return 0, errMustBeInteger
	}
}

func elvToNum(arg any) (Num, error) {
	switch arg := arg.(type) {
	case int, *big.Int, *big.Rat, float64:
		return arg, nil
	case string:
		n := ParseNum(arg)
		if n == nil {
			return 0, cannotParseAs{"number", ReprPlain(arg)}
		}
		return n, nil
	default:
		return 0, errMustBeNumber
	}
}

func elvToRune(arg any) (rune, error) {
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

func scanFieldMapFromMap(src any, ptr any, dstKeys FieldMapKeys, opt ScanOpt) error {
	makeErr := func(keysDescription string) error {
		return errs.BadValue{
			// TODO: Add path information in error messages.
			What:   "value",
			Valid:  fmt.Sprintf("map with keys %s [%s]", keysDescription, strings.Join(dstKeys, " ")),
			Actual: ReprPlain(src),
		}
	}

	switch opt & (AllowMissingMapKey | AllowExtraMapKey) {
	case 0:
		if Len(src) != len(dstKeys) {
			return makeErr("being exactly")
		}
	case AllowMissingMapKey:
		if Len(src) > len(dstKeys) {
			return makeErr("constrained to")
		}
	case AllowExtraMapKey:
		if Len(src) < len(dstKeys) {
			return makeErr("containing at least")
		}
	}
	dst := reflect.ValueOf(ptr).Elem()
	usedSrcKeys := 0
	for i, key := range dstKeys {
		srcValue, err := Index(src, key)
		if err != nil {
			if opt&AllowMissingMapKey == 0 {
				return makeErr("containing at least")
			}
			continue
		}
		err = ScanToGoOpts(srcValue, dst.Field(i).Addr().Interface(), opt)
		if err != nil {
			return err
		}
		usedSrcKeys++
	}
	if opt&AllowExtraMapKey == 0 && usedSrcKeys < Len(src) {
		return makeErr("constrained to")
	}
	return nil
}

// ScanListToGo converts a List to a slice, using ScanToGo to convert each
// element.
func ScanListToGo(src List, ptr any) error {
	n := src.Len()
	values := reflect.MakeSlice(reflect.TypeOf(ptr).Elem(), n, n)
	i := 0
	for it := src.Iterator(); it.HasElem(); it.Next() {
		err := ScanToGo(it.Elem(), values.Index(i).Addr().Interface())
		if err != nil {
			return err
		}
		i++
	}
	reflect.ValueOf(ptr).Elem().Set(values)
	return nil
}

// Optional wraps the last pointer passed to ScanListElementsToGo, to indicate
// that it is optional.
func Optional(ptr any) any { return optional{ptr} }

type optional struct{ ptr any }

// ScanListElementsToGo unpacks elements from a list, storing the each element
// in the given pointers with ScanToGo.
//
// The last pointer may be wrapped with Optional to indicate that it is
// optional.
func ScanListElementsToGo(src List, ptrs ...any) error {
	if o, ok := ptrs[len(ptrs)-1].(optional); ok {
		switch src.Len() {
		case len(ptrs) - 1:
			ptrs = ptrs[:len(ptrs)-1]
		case len(ptrs):
			ptrs[len(ptrs)-1] = o.ptr
		default:
			return errs.ArityMismatch{What: "list elements",
				ValidLow: len(ptrs) - 1, ValidHigh: len(ptrs), Actual: src.Len()}
		}
	} else if src.Len() != len(ptrs) {
		return errs.ArityMismatch{What: "list elements",
			ValidLow: len(ptrs), ValidHigh: len(ptrs), Actual: src.Len()}
	}

	i := 0
	for it := src.Iterator(); it.HasElem(); it.Next() {
		err := ScanToGo(it.Elem(), ptrs[i])
		if err != nil {
			return err
		}
		i++
	}
	return nil
}

// FromGo converts a Go value to an Elvish value.
//
// Exact numbers are normalized to the smallest types that can hold them, and
// runes are converted to strings. Values of other types are returned unchanged.
//
// See the related [ScanToGo] for the concepts of "Go value" and "Elvish value".
func FromGo(a any) any {
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
