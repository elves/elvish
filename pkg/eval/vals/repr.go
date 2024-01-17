package vals

import (
	"fmt"
	"math"
	"math/big"
	"sort"
	"strconv"

	"github.com/mattn/go-runewidth"

	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/persistent/hashmap"
)

// Reprer wraps the Repr method.
type Reprer interface {
	// Repr returns a string that represents a Value. The string either be a
	// literal of that Value that is preferably deep-equal to it (like `[a b c]`
	// for a list), or a string enclosed in "<>" containing the kind and
	// identity of the Value(like `<fn 0xdeadcafe>`).
	//
	// If indent is not negative the value should be pretty-printed with the
	// current indentation level of indent; the indent of the first line has
	// already been written and shall not be written in Repr. The returned
	// string should never contain a trailing newline.
	Repr(indent int) string
}

// ReprPlain is like Repr, but without pretty-printing.
func ReprPlain(v any) string {
	return Repr(v, math.MinInt)
}

// Repr returns the representation for a value, a string that is preferably
// (but not necessarily) an Elvish expression that evaluates to the argument.
// The representation is pretty-printed, using indent as the initial level of
// indentation. It is implemented for the builtin types nil, bool and string,
// the File, List and Map types, StructMap types, and types satisfying the
// Reprer interface. For other types, it uses fmt.Sprint with the format
// "<unknown %v>".
func Repr(v any, indent int) string {
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
	case int:
		return "(num " + strconv.Itoa(v) + ")"
	case *big.Int:
		return "(num " + v.String() + ")"
	case *big.Rat:
		return "(num " + v.String() + ")"
	case float64:
		return "(num " + formatFloat64(v) + ")"
	case File:
		return fmt.Sprintf("<file{%s %d}>", parse.Quote(v.Name()), v.Fd())
	case List:
		b := NewListReprBuilder(indent)
		for it := v.Iterator(); it.HasElem(); it.Next() {
			b.WriteElem(Repr(it.Elem(), indent+1))
		}
		return b.String()
	case Map:
		return reprMap(v.Iterator(), v.Len(), indent)
	case StructMap:
		return reprMap(iterateStructMap(v), lenStructMap(v), indent)
	case Reprer:
		return v.Repr(indent)
	case PseudoMap:
		m := v.Fields()
		s := reprMap(iterateStructMap(m), lenStructMap(m), indent)
		// Add a tag immediately after [.
		return "[^" + Kind(v) + " " + s[1:]
	default:
		return fmt.Sprintf("<unknown %v>", v)
	}
}

func reprMap(it hashmap.Iterator, n, indent int) string {
	builder := NewMapReprBuilder(indent)
	maxKeyWidth := 0
	// Collect all the key-value pairs.
	pairs := make([][2]any, 0, n)
	for ; it.HasElem(); it.Next() {
		k, v := it.Elem()
		pairs = append(pairs, [2]any{k, v})
		if n := runewidth.StringWidth(Repr(k, 0)); n > maxKeyWidth {
			maxKeyWidth = n
		}
	}
	// Sort the pairs. See the godoc of CmpTotal for the sorting algorithm.
	sort.Slice(pairs, func(i, j int) bool {
		return CmpTotal(pairs[i][0], pairs[j][0]) == CmpLess
	})
	// Print the pairs.
	for _, pair := range pairs {
		k, v := pair[0], pair[1]
		builder.WritePair(maxKeyWidth, Repr(k, indent+1), indent+2, Repr(v, indent+2))
	}
	return builder.String()
}
