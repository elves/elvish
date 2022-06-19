package eval

// Pointers to functions that can be mutated for testing.
var (
	GetHome   = &getHome
	Getwd     = &getwd
	TimeAfter = &timeAfter
)
