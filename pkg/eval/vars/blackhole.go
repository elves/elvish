package vars

type blackhole struct{}

func (blackhole) Set(any) error {
	return nil
}

func (blackhole) Get() any {
	return nil
}

// NewBlackhole returns a blackhole variable. Assignments to a blackhole
// variable will be discarded, and getting a blackhole variable always returns
// nil.
func NewBlackhole() Var {
	return blackhole{}
}

// IsBlackhole returns whether the variable is a blackhole variable.
func IsBlackhole(v Var) bool {
	_, ok := v.(blackhole)
	return ok
}
