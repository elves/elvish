package vals

import (
	"github.com/xiaq/persistent/hashmap"
)

type mapIterable interface {
	Iterator() hashmap.Iterator
}
type mapAssocable interface {
	Assoc(k, v interface{}) hashmap.Map
}
type mapDissocable interface {
	Dissoc(interface{}) hashmap.Map
}

var (
	_ mapIterable   = hashmap.Map(nil)
	_ Indexer       = hashmap.Map(nil)
	_ mapAssocable  = hashmap.Map(nil)
	_ mapDissocable = hashmap.Map(nil)
)

// EmptyMap is an empty map.
var EmptyMap = hashmap.New(Equal, Hash)

// MakeMap converts a native Go map to Map.
func MakeMap(raw map[interface{}]interface{}) hashmap.Map {
	m := EmptyMap
	for k, v := range raw {
		m = m.Assoc(k, v)
	}
	return m
}

// MakeMapFromKV creates a map from arguments that are alternately keys and
// values. It panics if the number of arguments is odd.
func MakeMapFromKV(a ...interface{}) hashmap.Map {
	if len(a)%2 == 1 {
		panic("Odd number of arguments to MakeMapFromKV")
	}
	m := EmptyMap
	for i := 0; i < len(a); i += 2 {
		m = m.Assoc(a[i], a[i+1])
	}
	return m
}

// MapReprBuilder helps building the Repr of a Map. It is also useful for
// implementing other Map-like values. The zero value of a MapReprBuilder is
// ready to use.
type MapReprBuilder struct {
	ListReprBuilder
}

func (b *MapReprBuilder) WritePair(k string, indent int, v string) {
	if indent > 0 {
		b.WriteElem("&" + k + "=\t" + v)
	} else {
		b.WriteElem("&" + k + "=" + v)
	}
}

func (b *MapReprBuilder) String() string {
	s := b.ListReprBuilder.String()
	if s == "[]" {
		s = "[&]"
	}
	return s
}
