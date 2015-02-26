package edit

// CeilDiv computes ceil(float(a)/b) but does not actually use float
// arithmetics.
func CeilDiv(a, b int) int {
	return (a + b - 1) / b
}
