package types

import (
	"encoding/json"

	"github.com/xiaq/persistent/hashmap"
)

// Map is a map from string to Value.
type Map struct {
	inner hashmap.Map
}

var _ MapLike = Map{}

type noSuchKeyError struct {
	key Value
}

// NoSuchKey returns an error indicating that a key is not found in a map-like
// value.
func NoSuchKey(k Value) error {
	return noSuchKeyError{k}
}

func (err noSuchKeyError) Error() string {
	return "no such key: " + Repr(err.key, NoPretty)
}

var EmptyMapInner = hashmap.New(Equal, Hash)

// EmptyMap is an empty Map.
var EmptyMap = Map{EmptyMapInner}

// NewMap creates a new Map from an inner HashMap.
func NewMap(inner hashmap.Map) Map {
	return Map{inner}
}

// MakeMap converts a native Go map to Map.
func MakeMap(m map[Value]Value) Map {
	inner := EmptyMapInner
	for k, v := range m {
		inner = inner.Assoc(k, v)
	}
	return NewMap(inner)
}

func (Map) Kind() string {
	return "map"
}

func (m Map) Equal(a interface{}) bool {
	return m == a || EqMapLike(m, a)
}

func (m Map) Hash() uint32 {
	return HashMapLike(m)
}

func (m Map) MarshalJSON() ([]byte, error) {
	// TODO(xiaq): Replace with a more efficient implementation.
	mm := map[string]Value{}
	for it := m.inner.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		mm[ToString(k.(Value))] = v.(Value)
	}
	return json.Marshal(mm)
}

func (m Map) Repr(indent int) string {
	var builder MapReprBuilder
	builder.Indent = indent
	for it := m.inner.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		builder.WritePair(Repr(k, indent+1), indent+2, Repr(v, indent+2))
	}
	return builder.String()
}

func (m Map) Len() int {
	return m.inner.Len()
}

func (m Map) Index(idx Value) (Value, error) {
	v, ok := m.inner.Get(idx)
	if !ok {
		return nil, NoSuchKey(idx)
	}
	return v.(Value), nil
}

func (m Map) Assoc(k, v Value) (Value, error) {
	return Map{m.inner.Assoc(k, v)}, nil
}

func (m Map) Dissoc(k Value) Value {
	return Map{m.inner.Without(k)}
}

func (m Map) IterateKey(f func(Value) bool) {
	for it := m.inner.Iterator(); it.HasElem(); it.Next() {
		k, _ := it.Elem()
		if !f(k.(Value)) {
			break
		}
	}
}

func (m Map) IteratePair(f func(Value, Value) bool) {
	for it := m.inner.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		if !f(k.(Value), v.(Value)) {
			break
		}
	}
}

func (m Map) HasKey(k Value) bool {
	_, ok := m.inner.Get(k)
	return ok
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
