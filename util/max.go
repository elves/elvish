package util

func MaxInt(x0 int, xs ...int) int {
	m := x0
	for _, x := range xs {
		if m < x {
			m = x
		}
	}
	return m
}
