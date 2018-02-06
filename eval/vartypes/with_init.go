package vartypes

type any struct {
	ptr *interface{}
}

func (v any) Set(val interface{}) error {
	*v.ptr = val
	return nil
}

func (v any) Get() interface{} {
	return *v.ptr
}

// NewAny creates a variable with an initial value. The variable created can be
// assigned values of any type.
func NewAny(v interface{}) Variable {
	return any{&v}
}
