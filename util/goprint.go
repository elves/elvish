package util

import (
	"bytes"
	"fmt"
	"reflect"
)

// GoPrint is like printing with the %#v formatter of fmt, but it prints
// pointer fields recursively.
func GoPrint(x interface{}) string {
	b := &bytes.Buffer{}
	goPrint(b, reflect.ValueOf(x))
	return b.String()
}

func goPrint(b *bytes.Buffer, v reflect.Value) {
	i := v.Interface()
	t := v.Type()
	if g, ok := i.(fmt.GoStringer); ok {
		b.WriteString(g.GoString())
		return
	}
	switch v.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.Struct:

	case reflect.Ptr:
		if v.IsNil() {
			b.WriteString(t.String())
			b.WriteString("(nil)")
		} else {
			b.WriteRune('&')
			goPrint(b, reflect.Indirect(v))
		}
		return
	default:
		fmt.Fprintf(b, "%#v", i)
		return
	}

	b.WriteString(t.String())

	if (v.Kind() == reflect.Map || v.Kind() == reflect.Slice) && v.IsNil() {
		b.WriteString("(nil)")
		return
	}

	b.WriteRune('{')
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				b.WriteString(", ")
			}
			goPrint(b, v.Index(i))
		}
	case reflect.Map:
		keys := v.MapKeys()
		for i, k := range keys {
			if i > 0 {
				b.WriteString(", ")
			}
			goPrint(b, k)
			b.WriteString(": ")
			goPrint(b, v.MapIndex(k))
		}
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(t.Field(i).Name)
			b.WriteString(": ")
			goPrint(b, v.Field(i))
		}
	}
	b.WriteRune('}')
}
