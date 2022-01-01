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

// Conversion between "Go values" (those expected by native Go functions) and
// "Elvish values" (those participating in the Elvish runtime).
//
// Among the conversion functions, ScanToGo and FromGo implement the implicit
// conversion used when calling native Go functions from Elvish. The API is
// asymmetric; this has to do with two characteristics of Elvish's type system:
//
// - Elvish doesn't have a dedicated rune type and uses strings to represent
//   them.
//
// - Elvish permits using strings that look like numbers in place of numbers.
//
// As a result, while FromGo can always convert a "Go value" to an "Elvish
// value" unambiguously, ScanToGo can't do that in the opposite direction.
// For example, "1" may be converted into "1", '1' or 1, depending on what
// the destination type is, and the process may fail. Thus ScanToGo takes the
// pointer to the destination as an argument, and returns an error.
//
// The rest of the conversion functions need to explicitly invoked.

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

// ScanToGo converts an Elvish value, and stores it in the destination of ptr,
// which must be a pointer.
//
// If ptr has type *int, *float64, *Num or *rune, it performs a suitable
// conversion, and returns an error if the conversion fails. In other cases,
// this function just tries to perform "*ptr = src" via reflection and returns
// an error if the assignment can't be done.
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
			return fmt.Errorf("internal bug: need pointer to scan to, got %T", ptr)
		}
		dstType := ptrType.Elem()
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
		return 0, cannotParseAs{"integer", ReprPlain(arg)}
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
			return 0, cannotParseAs{"number", ReprPlain(arg)}
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

// ScanListToGo converts a List to a slice, using ScanToGo to convert each
// element.
func ScanListToGo(src List, ptr interface{}) error {
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
func Optional(ptr interface{}) interface{} { return optional{ptr} }

type optional struct{ ptr interface{} }

// ScanListElementsToGo unpacks elements from a list, storing the each element
// in the given pointers with ScanToGo.
//
// The last pointer may be wrapped with Optional to indicate that it is
// optional.
func ScanListElementsToGo(src List, ptrs ...interface{}) error {
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
func ScanMapToGo(src Map, ptr interface{}) error {
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
