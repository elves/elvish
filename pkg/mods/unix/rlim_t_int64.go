//go:build freebsd

package unix

import (
	"math"
	"math/big"
	"strconv"
)

type rlimT = int64

var rlimTValid = "number between 0 and " + strconv.FormatInt(math.MaxInt64, 10)

const maxInt = int64(^uint(0) >> 1)

func convertRlimT(x int64) any {
	if x <= maxInt {
		return int(x)
	}
	return big.NewInt(int64(x))

}

func parseRlimT(val any) (int64, bool) {
	switch val := val.(type) {
	case int:
		return int64(val), true
	case *big.Int:
		if val.IsInt64() {
			return val.Int64(), true
		}
	case string:
		num, err := strconv.ParseInt(val, 0, 64)
		if err == nil {
			return num, true
		}
	}
	return 0, false
}
