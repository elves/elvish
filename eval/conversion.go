package eval

import (
	"fmt"
	"reflect"
	"strconv"
	"unicode/utf8"

	"github.com/elves/elvish/eval/types"
)

// Conversion between Go value and Value.

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
	switch dstPtr := ptr.(type) {
	case *int:
		i, err := toInt(src)
		maybeThrow(err)
		*dstPtr = i
	case *float64:
		f, err := toFloat(src)
		maybeThrow(err)
		*dstPtr = f
	default:
		ptrReflect := reflect.ValueOf(dstPtr)
		if ptrReflect.Kind() != reflect.Ptr {
			throwf("internal bug: %T to ScanArgs, need pointer", dstPtr)
		}
		dstReflect := reflect.Indirect(ptrReflect)
		if reflect.TypeOf(src).AssignableTo(dstReflect.Type()) {
			dstReflect.Set(reflect.ValueOf(src))
		} else {
			throwf("need %s argument, got %s",
				types.Kind(reflect.Zero(dstReflect.Type()).Interface()),
				types.Kind(src))
		}
	}
}
