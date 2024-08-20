package vals

import (
	"testing"

	"src.elv.sh/pkg/persistent/hash"
	"src.elv.sh/pkg/tt"
)

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
