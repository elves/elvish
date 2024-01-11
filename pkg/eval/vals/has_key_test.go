package vals

import (
	"testing"

	"src.elv.sh/pkg/tt"
)

type hasKeyer struct{ key any }

func (h hasKeyer) HasKey(k any) bool { return k == h.key }

func TestHasKey(t *testing.T) {
	tt.Test(t, HasKey,
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
		Args(MakeList("lorem", "ipsum"), "0..=").Rets(true),
		Args(MakeList("lorem", "ipsum"), "..2").Rets(true),
		Args(MakeList("lorem", "ipsum"), "..=2").Rets(false),
		Args(MakeList("lorem", "ipsum"), "2").Rets(false),
		Args(MakeList("lorem", "ipsum", "dolor", "sit"), "0..4").Rets(true),
		Args(MakeList("lorem", "ipsum", "dolor", "sit"), "0..=4").Rets(false),
		Args(MakeList("lorem", "ipsum", "dolor", "sit"), "1..3").Rets(true),
		Args(MakeList("lorem", "ipsum", "dolor", "sit"), "1..5").Rets(false),
		Args(MakeList("lorem", "ipsum", "dolor", "sit"), "-2..=-1").Rets(true),

		// Non-container
		Args(1, "0").Rets(false),
	)
}
