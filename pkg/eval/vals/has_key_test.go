package vals

import (
	"testing"

	. "src.elv.sh/pkg/tt"
)

type hasKeyer struct{ key any }

func (h hasKeyer) HasKey(k any) bool { return k == h.key }

func TestHasKey(t *testing.T) {
	Test(t, Fn("HasKey", HasKey), Table{
		// Map
		Args(MakeMap("k", "v"), "k").Rets(true),
		Args(MakeMap("k", "v"), "bad").Rets(false),
		// HasKeyer
		Args(hasKeyer{"valid"}, "valid").Rets(true),
		Args(hasKeyer{"valid"}, "invalid").Rets(false),
		// Fallback to IterateKeys
		Args(keysIterator{vs("lorem")}, "lorem").Rets(true),
		Args(keysIterator{vs("lorem")}, "ipsum").Rets(false),
		// Fallback to Len
		Args(MakeList("lorem", "ipsum"), "0").Rets(true),
		Args(MakeList("lorem", "ipsum"), "0..").Rets(true),
		Args(MakeList("lorem", "ipsum"), "2").Rets(false),

		// Non-container
		Args(1, "0").Rets(false),
	})
}
