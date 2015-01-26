package eval

import "os"

// Variable is the internal representation of a variable.
type Variable interface {
	Set(v Value)
	Get() Value
	StaticType() Type
}

type internalVariable struct {
	valuePtr   *Value
	staticType Type
}

func newInternalVariable(v Value, t Type) Variable {
	return internalVariable{&v, t}
}

func newInternalVariableWithType(v Value) Variable {
	return internalVariable{&v, v.Type()}
}

func (iv internalVariable) Set(val Value) {
	*iv.valuePtr = val
}

func (iv internalVariable) Get() Value {
	return *iv.valuePtr
}

func (iv internalVariable) StaticType() Type {
	return iv.staticType
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

func (ev envVariable) StaticType() Type {
	return stringType{}
}
