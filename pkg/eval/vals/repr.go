package vals

import (
	"fmt"
	"math"
	"reflect"

	"src.elv.sh/pkg/parse"
)

// NoPretty can be passed to Repr to suppress pretty-printing.
const NoPretty = math.MinInt32

// Reprer wraps the Repr method.
type Reprer interface {
	// Repr returns a string that represents a Value. The string either be a
	// literal of that Value that is preferably deep-equal to it (like `[a b c]`
	// for a list), or a string enclosed in "<>" containing the kind and
	// identity of the Value(like `<fn 0xdeadcafe>`).
	//
	// If indent is at least 0, it should be pretty-printed with the current
	// indentation level of indent; the indent of the first line has already
	// been written and shall not be written in Repr. The returned string
	// should never contain a trailing newline.
	Repr(indent int) string
}

// Repr returns the representation for a value, a string that is preferably (but
// not necessarily) an Elvish expression that evaluates to the argument. If
// indent >= 0, the representation is pretty-printed. It is implemented for the
// builtin types nil, bool and string, the File, List and Map types, StructMap
// types, and types satisfying the Reprer interface. For other types, it uses
// fmt.Sprint with the format "<unknown %v>".
func Repr(v interface{}, indent int) string {
	switch v := v.(type) {
	case nil:
		return "$nil"
	case bool:
		if v {
			return "$true"
		}
		return "$false"
	case string:
		return parse.Quote(v)
	case float64:
		return "(float64 " + formatFloat64(v) + ")"
	case []interface{}:
		// This case is to support functions like eval.doFmPrintf which might want to print a Go
		// slice of interface{} types. It obviously mirrors the `case List` below that is used for
		// Elvish lists.
		b := NewSliceReprBuilder(indent)
		for _, val := range v {
			b.WriteElem(Repr(val, indent+1))
		}
		return b.String()
	case Reprer:
		return v.Repr(indent)
	case File:
		return fmt.Sprintf("<file{%s %d}>", parse.Quote(v.Name()), v.Fd())
	case List:
		b := NewListReprBuilder(indent)
		for it := v.Iterator(); it.HasElem(); it.Next() {
			b.WriteElem(Repr(it.Elem(), indent+1))
		}
		return b.String()
	case Map:
		builder := NewMapReprBuilder(indent)
		for it := v.Iterator(); it.HasElem(); it.Next() {
			k, v := it.Elem()
			builder.WritePair(Repr(k, indent+1), indent+2, Repr(v, indent+2))
		}
		return builder.String()
	case StructMap:
		return reprStructMap(v, indent)
	case PseudoStructMap:
		return reprStructMap(v.Fields(), indent)
	default:
		return fmt.Sprintf("<unknown %v>", v)
	}
}

func reprStructMap(v StructMap, indent int) string {
	vValue := reflect.ValueOf(v)
	vType := vValue.Type()
	builder := NewMapReprBuilder(indent)
	it := iterateStructMap(vType)
	for it.Next() {
		k, v := it.Get(vValue)
		builder.WritePair(Repr(k, indent+1), indent+2, Repr(v, indent+2))
	}
	return builder.String()
}
