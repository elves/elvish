package vals

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"sync"
	"unicode/utf8"

	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/strutil"
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

// ScanToGo converts an Elvish value to a Go value, and stores it in *ptr.
//
//   - If ptr has type *int, *float64, *Num or *rune, it performs a suitable
//     conversion, and returns an error if the conversion fails.
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
func ScanToGo(src any, ptr any) error {
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
		return assignPtr(src, ptr)
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

// Does "*ptr = src" via reflection.
func assignPtr(src, ptr any) error {
	ptrType := TypeOf(ptr)
	if ptrType.Kind() != reflect.Ptr {
		return fmt.Errorf("internal bug: need pointer to scan to, got %T", ptr)
	}
	dstType := ptrType.Elem()
	// Allow using any(nil) as T(nil) for any T whose zero value is spelt
	// nil.
	if src == nil {
		switch dstType.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
			ValueOf(ptr).Elem().SetZero()
			return nil
		}
	}
	if !TypeOf(src).AssignableTo(dstType) {
		var dstKind string
		if dstType.Kind() == reflect.Interface {
			dstKind = "!!" + dstType.String()
		} else {
			dstKind = Kind(reflect.Zero(dstType).Interface())
		}
		return WrongType{dstKind, Kind(src)}
	}
	ValueOf(ptr).Elem().Set(ValueOf(src))
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

// ScanMapToGo scans map elements into ptr, which must be a pointer to a struct.
// Struct field names are converted to map keys with CamelToDashed.
//
// The map may contains keys that don't correspond to struct fields, and it
// doesn't have to contain all keys that correspond to struct fields.
func ScanMapToGo(src Map, ptr any) error {
	// Iterate over the struct keys instead of the map: since extra keys are
	// allowed, the map may be very big, while the size of the struct is bound.
	keys, _ := StructFieldsInfo(reflect.TypeOf(ptr).Elem())
	structValue := reflect.ValueOf(ptr).Elem()
	for i, key := range keys {
		if key == "" {
			continue
		}
		val, ok := src.Index(key)
		if !ok {
			continue
		}
		err := ScanToGo(val, structValue.Field(i).Addr().Interface())
		if err != nil {
			return err
		}
	}
	return nil
}

// StructFieldsInfo takes a type for a struct, and returns a slice for each
// field name, converted with CamelToDashed, and a reverse index. Unexported
// fields result in an empty string in the slice, and is omitted from the
// reverse index.
func StructFieldsInfo(t reflect.Type) ([]string, map[string]int) {
	if info, ok := structFieldsInfoCache.Load(t); ok {
		info := info.(structFieldsInfo)
		return info.keys, info.keyIdx
	}
	info := makeStructFieldsInfo(t)
	structFieldsInfoCache.Store(t, info)
	return info.keys, info.keyIdx
}

var structFieldsInfoCache sync.Map

type structFieldsInfo struct {
	keys   []string
	keyIdx map[string]int
}

func makeStructFieldsInfo(t reflect.Type) structFieldsInfo {
	keys := make([]string, t.NumField())
	keyIdx := make(map[string]int)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		key := strutil.CamelToDashed(field.Name)
		keyIdx[key] = i
		keys[i] = key
	}
	return structFieldsInfo{keys, keyIdx}
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
