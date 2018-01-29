package types

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
	Without(interface{}) hashmap.Map
}

var (
	_ mapIterable   = hashmap.Map(nil)
	_ Getter        = hashmap.Map(nil)
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
