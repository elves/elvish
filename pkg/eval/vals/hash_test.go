package vals

import (
	"math"
	"math/big"
	"os"
	"testing"
	"unsafe"

	"src.elv.sh/pkg/persistent/hash"
	"src.elv.sh/pkg/persistent/hashmap"
	"src.elv.sh/pkg/tt"
)

type hasher struct{}

func (hasher) Hash() uint32 { return 42 }

type nonHasher struct{}

func TestHash(t *testing.T) {
	z := big.NewInt(5)
	z.Lsh(z, 8*uint(unsafe.Sizeof(int(0))))
	z.Add(z, big.NewInt(9))
	// z = 5 << wordSize + 9

	tt.Test(t, Hash,
		Args(false).Rets(uint32(0)),
		Args(true).Rets(uint32(1)),
		Args(1).Rets(uint32(1)),
		Args(z).Rets(hash.DJB(1, 9, 5)),
		Args(big.NewRat(3, 2)).Rets(hash.DJB(Hash(big.NewInt(3)), Hash(big.NewInt(2)))),
		Args(1.0).Rets(hash.UInt64(math.Float64bits(1.0))),
		Args("foo").Rets(hash.String("foo")),
		Args(os.Stdin).Rets(hash.UIntPtr(os.Stdin.Fd())),
		Args(MakeList("foo", "bar")).Rets(hash.DJB(Hash("foo"), Hash("bar"))),
		Args(MakeMap("foo", "bar")).
			Rets(hash.DJB(Hash("foo"), Hash("bar"))),
		Args(hasher{}).Rets(uint32(42)),
		Args(nonHasher{}).Rets(uint32(0)),
	)
}

func TestHash_EqualMapsWithDifferentInternal(t *testing.T) {
	// The internal representation of maps with the same value is not always the
	// same: when some keys of the map have the same hash, their values are
	// stored in the insertion order.
	//
	// To reliably test this case, we construct maps with a custom hashing
	// function.
	m0 := hashmap.New(Equal, func(v any) uint32 { return 0 })
	m1 := m0.Assoc("k1", "v1").Assoc("k2", "v2")
	m2 := m0.Assoc("k2", "v2").Assoc("k1", "v1")
	if h1, h2 := Hash(m1), Hash(m2); h1 != h2 {
		t.Errorf("%v != %v", h1, h2)
	}
}
