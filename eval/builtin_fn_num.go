package eval

import (
	"math"
	"math/rand"

	"github.com/elves/elvish/eval/vals"
)

// Numerical operations.

func init() {
	addBuiltinFns(map[string]interface{}{
		// Comparison
		"<":  lt,
		"<=": le,
		"==": eqNum,
		"!=": ne,
		">":  gt,
		">=": ge,

		// Arithmetics
		"+": plus,
		"-": minus,
		"*": times,
		"/": slash,
		"^": math.Pow,
		"%": func(a, b int) int { return a % b },

		// Random
		"rand":    rand.Float64,
		"randint": randint,
	})
}

func lt(nums ...float64) bool {
	for i := 0; i < len(nums)-1; i++ {
		if !(nums[i] < nums[i+1]) {
			return false
		}
	}
	return true
}

func le(nums ...float64) bool {
	for i := 0; i < len(nums)-1; i++ {
		if !(nums[i] <= nums[i+1]) {
			return false
		}
	}
	return true
}

func eqNum(nums ...float64) bool {
	for i := 0; i < len(nums)-1; i++ {
		if !(nums[i] == nums[i+1]) {
			return false
		}
	}
	return true
}

func ne(nums ...float64) bool {
	for i := 0; i < len(nums)-1; i++ {
		if !(nums[i] != nums[i+1]) {
			return false
		}
	}
	return true
}

func gt(nums ...float64) bool {
	for i := 0; i < len(nums)-1; i++ {
		if !(nums[i] > nums[i+1]) {
			return false
		}
	}
	return true
}

func ge(nums ...float64) bool {
	for i := 0; i < len(nums)-1; i++ {
		if !(nums[i] >= nums[i+1]) {
			return false
		}
	}
	return true
}

func plus(nums ...float64) float64 {
	sum := 0.0
	for _, f := range nums {
		sum += f
	}
	return sum
}

func minus(sum float64, nums ...float64) float64 {
	if len(nums) == 0 {
		// Unary -
		return -sum
	}
	for _, f := range nums {
		sum -= f
	}
	return sum
}

func times(nums ...float64) float64 {
	prod := 1.0
	for _, f := range nums {
		prod *= f
	}
	return prod
}

func slash(fm *Frame, args ...float64) error {
	if len(args) == 0 {
		// cd /
		return fm.Chdir("/")
	}
	// Division
	divide(fm, args[0], args[1:]...)
	return nil
}

func divide(fm *Frame, prod float64, nums ...float64) {
	out := fm.ports[1].Chan
	for _, f := range nums {
		prod /= f
	}
	out <- vals.FromGo(prod)
}

func randint(low, high int) (int, error) {
	if low >= high {
		return 0, ErrArgs
	}
	return low + rand.Intn(high-low), nil
}
