package vals

import (
	"testing"

	"src.elv.sh/pkg/persistent/hash"
	"src.elv.sh/pkg/tt"
)

type fieldMap struct {
	Foo    string
	Bar    string
	FooBar int
}

type fieldMap2 fieldMap

func TestFieldMap(t *testing.T) {
	TestValue(t, fieldMap{"lorem", "ipsum", 23}).
		Kind("map").
		Bool(true).
		Hash(
			hash.DJB(Hash("foo"), Hash("lorem"))+
				hash.DJB(Hash("bar"), Hash("ipsum"))+
				hash.DJB(Hash("foo-bar"), Hash(23))).
		Repr(`[&bar=ipsum &foo=lorem &foo-bar=(num 23)]`).
		Len(3).
		Equal(
			// Field maps behave like maps, so they are equal to normal maps
			// and other field maps with the same entries.
			MakeMap("foo", "lorem", "bar", "ipsum", "foo-bar", 23),
			fieldMap{"lorem", "ipsum", 23},
			fieldMap2{"lorem", "ipsum", 23}).
		NotEqual("a", MakeMap(), fieldMap{"lorem", "ipsum", 2}).
		HasKey("foo", "bar", "foo-bar").
		HasNoKey("bad", 1.0).
		IndexError("bad", NoSuchKey("bad")).
		IndexError(1.0, NoSuchKey(1.0)).
		AllKeys("foo", "bar", "foo-bar").
		Index("foo", "lorem").
		Index("bar", "ipsum").
		Index("foo-bar", 23)
}

type notFieldMap1 struct{ foo string }
type notFieldMap2 struct{ Embedded }
type Embedded struct{ Foo string }

func TestIsFieldMap(t *testing.T) {
	tt.Test(t, IsFieldMap,
		Args(fieldMap{}).Rets(true),
		Args(fieldMap2{}).Rets(true),
		Args(notFieldMap1{""}).Rets(false),
		Args(notFieldMap2{}).Rets(false),
	)
}

// Compare reflection-based operations on a field map compared to manual
// implementations.

func BenchmarkFieldMap_Index(b *testing.B) {
	benchmarkFieldMap_Index(b, fieldMap{})
}

type fieldMapIndexer fieldMap

var _ Indexer = fieldMapIndexer{}

func (m fieldMapIndexer) Index(k any) (any, bool) {
	switch k {
	case "foo":
		return m.Foo, true
	case "bar":
		return m.Bar, true
	case "foo-bar":
		return m.FooBar, true
	}
	return nil, false
}

func BenchmarkFieldMap_Index_Manual(b *testing.B) {
	benchmarkFieldMap_Index(b, fieldMapIndexer{})
}

func benchmarkFieldMap_Index(b *testing.B, value any) {
	for range b.N {
		Index(value, "foo")
		Index(value, "bar")
		Index(value, "lorem")
		Index(value, "ipsum")
	}
}

func BenchmarkFieldMap_HasKey(b *testing.B) {
	benchmarkFieldMap_HasKey(b, fieldMap{})
}

type fieldMapHasKeyer fieldMap

var _ HasKeyer = fieldMapHasKeyer{}

func (m fieldMapHasKeyer) HasKey(k any) bool {
	return k == "foo" || k == "bar" || k == "lorem"
}

func BenchmarkFieldMap_HasKey_Manual(b *testing.B) {
	benchmarkFieldMap_HasKey(b, fieldMapHasKeyer{})
}

func benchmarkFieldMap_HasKey(b *testing.B, value any) {
	for range b.N {
		HasKey(value, "foo")
		HasKey(value, "bar")
		HasKey(value, "lorem")
		HasKey(value, "ipsum")
	}
}

func BenchmarkFieldMap_IterateKeys(b *testing.B) {
	benchmarkFieldMap_IterateKeys(b, fieldMap{})
}

type fieldMapKeysIterator fieldMap

var _ KeysIterator = fieldMapKeysIterator{}

func (m fieldMapKeysIterator) IterateKeys(f func(any) bool) {
	if f("foo") {
		if f("bar") {
			f("lorem")
		}
	}
}

func BenchmarkFieldMap_IterateKeys_Manual(b *testing.B) {
	benchmarkFieldMap_IterateKeys(b, fieldMapKeysIterator{})
}

func benchmarkFieldMap_IterateKeys(b *testing.B, value any) {
	for range b.N {
		IterateKeys(value, func(any) bool { return true })
	}
}

func BenchmarkFieldMap_Len(b *testing.B) {
	benchmarkFieldMap_Len(b, fieldMap{})
}

type fieldMapLener fieldMap

var _ Lener = fieldMapLener{}

func (m fieldMapLener) Len() int { return 3 }

func BenchmarkFieldMap_Len_Manual(b *testing.B) {
	benchmarkFieldMap_Len(b, fieldMapLener{})
}

func benchmarkFieldMap_Len(b *testing.B, value any) {
	for range b.N {
		Len(value)
	}
}
