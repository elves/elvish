package eval

import (
	"math"
	"math/rand"
)

// Numerical operations.

func init() {
	addToReflectBuiltinFns(map[string]interface{}{
		// Comparison
		"<":  func(a, b float64) bool { return a < b },
		"<=": func(a, b float64) bool { return a <= b },
		"==": func(a, b float64) bool { return a == b },
		"!=": func(a, b float64) bool { return a != b },
		">":  func(a, b float64) bool { return a > b },
		">=": func(a, b float64) bool { return a >= b },

		// Arithmetics
		"+": plus,
		"-": minus,
		"*": times,
		"^": math.Pow,
		"%": func(a, b int) int { return a % b },

		// Random
		"rand":    rand.Float64,
		"randint": randint,
	})
	addToBuiltinFns([]*BuiltinFn{
		{"/", slash},
	})
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

func slash(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoOpt(opts)
	if len(args) == 0 {
		// cd /
		cdInner("/", ec)
		return
	}
	// Division
	divide(ec, args, opts)
}

func divide(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var (
		prod float64
		nums []float64
	)
	ScanArgsVariadic(args, &prod, &nums)
	TakeNoOpt(opts)

	out := ec.ports[1].Chan
	for _, f := range nums {
		prod /= f
	}
	out <- floatToElv(prod)
}

func randint(low, high int) (int, error) {
	if low >= high {
		return 0, ErrArgs
	}
	return low + rand.Intn(high-low), nil
}
