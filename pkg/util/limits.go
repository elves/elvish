package util

// Limit values for uint and int.
//
// NOTE: The math package contains similar constants for explicitly sized
// integer types, but lack those for uint and int.
const (
	MaxUint = ^uint(0)
	MinUint = 0
	MaxInt  = int(MaxUint >> 1)
	MinInt  = -MaxInt - 1
)
