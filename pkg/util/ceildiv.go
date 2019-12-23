package util

// CeilDiv computes ceil(float(a)/b) without using float arithmetics.
func CeilDiv(a, b int) int {
	return (a + b - 1) / b
}
