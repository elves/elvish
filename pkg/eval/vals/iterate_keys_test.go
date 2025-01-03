package vals

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

func vs(xs ...any) []any { return xs }

type keysIterator struct{ keys []any }

func (k keysIterator) IterateKeys(f func(any) bool) {
	Feed(f, k.keys...)
}

type nonKeysIterator struct{}

func TestIterateKeys(t *testing.T) {
	tt.Test(t, tt.Fn(collectKeys).Named("collectKeys"),
		Args(MakeMap("k1", "v1", "k2", "v2")).Rets(vs("k1", "k2"), nil),
		Args(keysIterator{vs("lorem", "ipsum")}).Rets(vs("lorem", "ipsum")),
		Args(nonKeysIterator{}).Rets(
			tt.Any, cannotIterateKeysOf{"!!vals.nonKeysIterator"}),
	)
}

func TestIterateKeys_Map_Break(t *testing.T) {
	var gotKey any
	IterateKeys(MakeMap("k", "v", "k2", "v2"), func(k any) bool {
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

func TestIterateKeys_FieldMap_Break(t *testing.T) {
	var gotKey any
	IterateKeys(fieldMap{}, func(k any) bool {
		if gotKey != nil {
			t.Errorf("callback called again after returning false")
		}
		gotKey = k
		return false
	})
	if gotKey != "foo" {
		t.Errorf("got key %v, want name", gotKey)
	}
}

func TestIterateKeys_Unsupported(t *testing.T) {
	err := IterateKeys(1, func(any) bool { return true })
	wantErr := cannotIterateKeysOf{"number"}
	if err != wantErr {
		t.Errorf("got error %v, want %v", err, wantErr)
	}
}
