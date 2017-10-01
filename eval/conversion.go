package eval

import (
	"fmt"
	"reflect"
	"strconv"
	"unicode/utf8"
)

// Conversion between Go value and Value.

func toFloat(arg Value) (float64, error) {
	if _, ok := arg.(String); !ok {
		return 0, fmt.Errorf("must be string")
	}
	s := string(arg.(String))
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

func floatToString(f float64) String {
	return String(strconv.FormatFloat(f, 'g', -1, 64))
}

func toInt(arg Value) (int, error) {
	arg, ok := arg.(String)
	if !ok {
		return 0, fmt.Errorf("must be string")
	}
	num, err := strconv.ParseInt(string(arg.(String)), 0, 0)
	if err != nil {
		return 0, err
	}
	return int(num), nil
}

func toRune(arg Value) (rune, error) {
	ss, ok := arg.(String)
	if !ok {
		return -1, fmt.Errorf("must be string")
	}
	s := string(ss)
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return -1, fmt.Errorf("string is not valid UTF-8")
	}
	if size != len(s) {
		return -1, fmt.Errorf("string has multiple runes")
	}
	return r, nil
}

func scanValueToGo(src Value, dstPtr interface{}) {
	ptr := reflect.ValueOf(dstPtr)
	if ptr.Kind() != reflect.Ptr {
		throwf("internal bug: %T to ScanArgs, need pointer", dstPtr)
	}
	dst := reflect.Indirect(ptr)
	switch dst.Kind() {
	case reflect.Int:
		i, err := toInt(src)
		maybeThrow(err)
		dst.Set(reflect.ValueOf(i))
	case reflect.Float64:
		f, err := toFloat(src)
		maybeThrow(err)
		dst.Set(reflect.ValueOf(f))
	default:
		if reflect.TypeOf(src).ConvertibleTo(dst.Type()) {
			dst.Set(reflect.ValueOf(src).Convert(dst.Type()))
		} else {
			throwf("need %T argument, got %s", dst.Interface(), src.Kind())
		}
	}
}
