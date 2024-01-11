package vals

import (
	"math"
	"math/big"
	"testing"

	"src.elv.sh/pkg/testutil"
	"src.elv.sh/pkg/tt"
)

// Test utilities.

const (
	zeros = "0000000000000000000"
	// Values that exceed the range of int64, used for testing BigInt.
	z   = "1" + zeros + "0"
	z1  = "1" + zeros + "1" // z+1
	z2  = "1" + zeros + "2" // z+2
	z3  = "1" + zeros + "3" // z+3
	zz  = "2" + zeros + "0" // 2z
	zz1 = "2" + zeros + "1" // 2z+1
	zz2 = "2" + zeros + "2" // 2z+2
	zz3 = "2" + zeros + "3" // 2z+3
)

func TestParseNum(t *testing.T) {
	tt.Test(t, ParseNum,
		Args("1").Rets(1),

		Args(z).Rets(bigInt(z)),

		Args("1/2").Rets(big.NewRat(1, 2)),
		Args("2/1").Rets(2),
		Args(z+"/1").Rets(bigInt(z)),

		Args("1.0").Rets(1.0),
		Args("1e-5").Rets(1e-5),

		Args("x").Rets(nil),
		Args("x/y").Rets(nil),
	)
}

func TestUnifyNums(t *testing.T) {
	tt.Test(t, UnifyNums,
		Args([]Num{1, 2, 3, 4}, Int).
			Rets([]int{1, 2, 3, 4}),

		Args([]Num{1, 2, 3, bigInt(z)}, Int).
			Rets([]*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3), bigInt(z)}),

		Args([]Num{1, 2, 3, big.NewRat(1, 2)}, Int).
			Rets([]*big.Rat{
				big.NewRat(1, 1), big.NewRat(2, 1),
				big.NewRat(3, 1), big.NewRat(1, 2)}),
		Args([]Num{1, 2, bigInt(z), big.NewRat(1, 2)}, Int).
			Rets([]*big.Rat{
				big.NewRat(1, 1), big.NewRat(2, 1), bigRat(z), big.NewRat(1, 2)}),

		Args([]Num{1, 2, 3, 4.0}, Int).
			Rets([]float64{1, 2, 3, 4}),
		Args([]Num{1, 2, big.NewRat(1, 2), 4.0}, Int).
			Rets([]float64{1, 2, 0.5, 4}),
		Args([]Num{1, 2, big.NewInt(3), 4.0}, Int).
			Rets([]float64{1, 2, 3, 4}),
		Args([]Num{1, 2, bigInt(z), 4.0}, Int).
			Rets([]float64{1, 2, math.Inf(1), 4}),

		Args([]Num{1, 2, 3, 4}, BigInt).
			Rets([]*big.Int{
				big.NewInt(1), big.NewInt(2), big.NewInt(3), big.NewInt(4)}),
	)
}

func TestUnifyNums2(t *testing.T) {
	tt.Test(t, UnifyNums2,
		Args(1, 2, Int).Rets(1, 2),
		Args(1, bigInt(z), Int).Rets(big.NewInt(1), bigInt(z)),
		Args(1, big.NewRat(1, 2), Int).Rets(big.NewRat(1, 1), big.NewRat(1, 2)),
		Args(1, 2.0, Int).Rets(1.0, 2.0),
		Args(1, 2, BigInt).Rets(big.NewInt(1), big.NewInt(2)),
	)
}

func TestInvalidNumType(t *testing.T) {
	tt.Test(t, testutil.Recover,
		Args(func() { UnifyNums([]Num{int32(0)}, 0) }).Rets("invalid num type int32"),
		Args(func() { PromoteToBigInt(int32(0)) }).Rets("invalid num type int32"),
		Args(func() { PromoteToBigRat(int32(0)) }).Rets("invalid num type int32"),
		Args(func() { ConvertToFloat64(int32(0)) }).Rets("invalid num type int32"),
	)
}

func TestInt64ToNum(t *testing.T) {
	n := Int64ToNum(1)
	if _, isInt := n.(int); !isInt {
		t.Errorf("got %T, want int", n)
	}

	if math.MaxInt != math.MaxInt64 {
		n = Int64ToNum(math.MaxInt64)
		if _, isBigInt := n.(*big.Int); !isBigInt {
			t.Errorf("got %T, want *big.Int", n)
		}
	}
}

func TestUint64ToNum(t *testing.T) {
	n := Uint64ToNum(1)
	if _, isInt := n.(int); !isInt {
		t.Errorf("got %T, want int", n)
	}

	n = Uint64ToNum(math.MaxUint64)
	if _, isBigInt := n.(*big.Int); !isBigInt {
		t.Errorf("got %T, want *big.Int", n)
	}
}

func bigInt(s string) *big.Int {
	z, ok := new(big.Int).SetString(s, 0)
	if !ok {
		panic("cannot parse as big int: " + s)
	}
	return z
}

func bigRat(s string) *big.Rat {
	z, ok := new(big.Rat).SetString(s)
	if !ok {
		panic("cannot parse as big rat: " + s)
	}
	return z
}
