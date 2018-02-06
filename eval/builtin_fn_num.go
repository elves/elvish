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
		"/": slash,
		"^": math.Pow,
		"%": func(a, b int) int { return a % b },

		// Random
		"rand":    rand.Float64,
		"randint": randint,
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

func slash(fm *Frame, args ...float64) {
	if len(args) == 0 {
		// cd /
		cdInner("/", fm)
		return
	}
	// Division
	divide(fm, args[0], args[1:]...)
}

func divide(fm *Frame, prod float64, nums ...float64) {
	out := fm.ports[1].Chan
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
