//go:build unix && !freebsd

package unix

import (
	"math"
	"math/big"
	"strconv"
)

type rlimT = uint64

var rlimTValid = "number between 0 and " + strconv.FormatUint(math.MaxUint64, 10)

const maxInt = uint64(^uint(0) >> 1)

func convertRlimT(x uint64) any {
	if x <= maxInt {
		return int(x)
	}
	if x <= math.MaxInt64 {
		return big.NewInt(int64(x))
	}
	z := big.NewInt(int64(x / 2))
	z.Lsh(z, 1)
	if x%2 == 1 {
		z.Bits()[0] |= 1
	}
	return z

}

func parseRlimT(val any) (uint64, bool) {
	switch val := val.(type) {
	case int:
		if val >= 0 {
			return uint64(val), true
		}
	case *big.Int:
		if val.IsUint64() {
			return val.Uint64(), true
		}
	case string:
		num, err := strconv.ParseUint(val, 0, 64)
		if err == nil {
			return num, true
		}
	}
	return 0, false
}
