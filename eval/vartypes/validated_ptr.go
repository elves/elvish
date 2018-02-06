package vartypes

type validatedPtr struct {
	valuePtr  *interface{}
	validator func(interface{}) error
}

type invalidValueError struct {
	inner error
}

func (err invalidValueError) Error() string {
	return "invalid value: " + err.inner.Error()
}

func NewValidatedPtr(v interface{}, vld func(interface{}) error) Variable {
	return validatedPtr{&v, vld}
}

func (iv validatedPtr) Set(val interface{}) error {
	if err := iv.validator(val); err != nil {
		return invalidValueError{err}
	}
	*iv.valuePtr = val
	return nil
}

func (iv validatedPtr) Get() interface{} {
	return *iv.valuePtr
}
