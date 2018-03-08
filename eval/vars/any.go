package vars

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

// NewAnyWithInit creates a variable with an initial value. The variable created
// can be assigned values of any type.
func NewAnyWithInit(v interface{}) Var {
	return any{&v}
}
