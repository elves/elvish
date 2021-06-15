package vals

import (
	"testing"

	. "src.elv.sh/pkg/tt"
)

func vs(xs ...interface{}) []interface{} { return xs }

type keysIterator struct{ keys []interface{} }

func (k keysIterator) IterateKeys(f func(interface{}) bool) {
	Feed(f, k.keys...)
}

type nonKeysIterator struct{}

func TestIterateKeys(t *testing.T) {
	Test(t, Fn("collectKeys", collectKeys), Table{
		Args(MakeMap("k1", "v1", "k2", "v2")).Rets(vs("k1", "k2"), nil),
		Args(keysIterator{vs("lorem", "ipsum")}).Rets(vs("lorem", "ipsum")),
		Args(nonKeysIterator{}).Rets(
			Any, cannotIterateKeysOf{"!!vals.nonKeysIterator"}),
	})
}

func TestIterateKeys_Map_Break(t *testing.T) {
	var gotKey interface{}
	IterateKeys(MakeMap("k", "v", "k2", "v2"), func(k interface{}) bool {
		if gotKey != nil {
			t.Errorf("callback called again after returning false")
		}
		gotKey = k
		return false
	})
	if gotKey != "k" && gotKey != "k2" {
		t.Errorf("got key %v, want k or k2", gotKey)
	}
}

func TestIterateKeys_StructMap_Break(t *testing.T) {
	var gotKey interface{}
	IterateKeys(testStructMap{}, func(k interface{}) bool {
		if gotKey != nil {
			t.Errorf("callback called again after returning false")
		}
		gotKey = k
		return false
	})
	if gotKey != "name" {
		t.Errorf("got key %v, want name", gotKey)
	}
}

func TestIterateKeys_Unsupported(t *testing.T) {
	err := IterateKeys(1, func(interface{}) bool { return true })
	wantErr := cannotIterateKeysOf{"number"}
	if err != wantErr {
		t.Errorf("got error %v, want %v", err, wantErr)
	}
}
