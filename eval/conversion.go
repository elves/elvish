package eval

import (
	"fmt"
	"reflect"
	"strconv"
	"unicode/utf8"

	"github.com/elves/elvish/eval/types"
)

// Conversion between Go value and Value.

func toFloat(arg types.Value) (float64, error) {
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

func toInt(arg types.Value) (int, error) {
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

func toRune(arg types.Value) (rune, error) {
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
func scanValueToGo(src types.Value, dstPtr interface{}) {
	switch dstPtr := dstPtr.(type) {
	case *string:
		s, ok := src.(string)
		if !ok {
			throwf("cannot convert %T to string", src)
		}
		*dstPtr = s
	case *int:
		i, err := toInt(src)
		maybeThrow(err)
		*dstPtr = i
	case *float64:
		f, err := toFloat(src)
		maybeThrow(err)
		*dstPtr = f
	default:
		ptr := reflect.ValueOf(dstPtr)
		if ptr.Kind() != reflect.Ptr {
			throwf("internal bug: %T to ScanArgs, need pointer", dstPtr)
		}
		dstReflect := reflect.Indirect(ptr)
		if reflect.TypeOf(src).ConvertibleTo(dstReflect.Type()) {
			dstReflect.Set(reflect.ValueOf(src).Convert(dstReflect.Type()))
		} else {
			throwf("need %s argument, got %s", dstReflect.Type().Name(), types.Kind(src))
		}
	}
}
