package vals

import (
	"math"
	"os"
	"testing"

	"github.com/elves/elvish/tt"
	"github.com/xiaq/persistent/hash"
)

type hasher struct{}

func (hasher) Hash() uint32 { return 42 }

type nonHasher struct{}

func TestHash(t *testing.T) {
	tt.Test(t, tt.Fn("Hash", Hash), tt.Table{
		Args(false).Rets(uint32(0)),
		Args(true).Rets(uint32(1)),
		Args(1.0).Rets(hash.UInt64(math.Float64bits(1.0))),
		Args("foo").Rets(hash.String("foo")),
		Args(os.Stdin).Rets(hash.UIntPtr(os.Stdin.Fd())),
		Args(MakeList("foo", "bar")).Rets(hash.DJB(Hash("foo"), Hash("bar"))),
		Args(MakeMap("foo", "bar")).
			Rets(hash.DJB(Hash("foo"), Hash("bar"))),
		Args(testStructMap{"name", 1.0}).
			Rets(hash.DJB(Hash("name"), Hash(1.0))),
		Args(hasher{}).Rets(uint32(42)),
		Args(nonHasher{}).Rets(uint32(0)),
	})
}
