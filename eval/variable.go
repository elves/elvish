package eval

import "os"

// Variable represents an elvish variable.
type Variable interface {
	Set(v Value)
	Get() Value
}

type ptrVariable struct {
	valuePtr *Value
}

func newPtrVariable(v Value) Variable {
	return ptrVariable{&v}
}

func (iv ptrVariable) Set(val Value) {
	*iv.valuePtr = val
}

func (iv ptrVariable) Get() Value {
	return *iv.valuePtr
}

type envVariable struct {
	name string
}

func newEnvVariable(name string) envVariable {
	return envVariable{name}
}

func (ev envVariable) Set(val Value) {
	os.Setenv(ev.name, ToString(val))
}

func (ev envVariable) Get() Value {
	return String(os.Getenv(ev.name))
}
