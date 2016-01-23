package eval

import "os"

// Variable is the internal representation of a variable.
type Variable interface {
	Set(v Value)
	Get() Value
}

type internalVariable struct {
	valuePtr *Value
}

func newInternalVariable(v Value) Variable {
	return internalVariable{&v}
}

func (iv internalVariable) Set(val Value) {
	*iv.valuePtr = val
}

func (iv internalVariable) Get() Value {
	return *iv.valuePtr
}

type envVariable struct {
	name string
}

func newEnvVariable(name string) envVariable {
	return envVariable{name}
}

func (ev envVariable) Set(val Value) {
	os.Setenv(ev.name, toString(val))
}

func (ev envVariable) Get() Value {
	return str(os.Getenv(ev.name))
}
