package vals

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

type hasKeyer struct{ key any }

func (h hasKeyer) HasKey(k any) bool { return k == h.key }

func TestHasKey(t *testing.T) {
	tt.Test(t, tt.Fn("HasKey", HasKey), tt.Table{
		// Map
		tt.Args(MakeMap("k", "v"), "k").Rets(true),
		tt.Args(MakeMap("k", "v"), "bad").Rets(false),
		// HasKeyer
		tt.Args(hasKeyer{"valid"}, "valid").Rets(true),
		tt.Args(hasKeyer{"valid"}, "invalid").Rets(false),
		// Fallback to IterateKeys
		tt.Args(keysIterator{vs("lorem")}, "lorem").Rets(true),
		tt.Args(keysIterator{vs("lorem")}, "ipsum").Rets(false),
		// Fallback to Len
		tt.Args(MakeList("lorem", "ipsum"), "0").Rets(true),
		tt.Args(MakeList("lorem", "ipsum"), "0..").Rets(true),
		tt.Args(MakeList("lorem", "ipsum"), "2").Rets(false),

		// Non-container
		tt.Args(1, "0").Rets(false),
	})
}
