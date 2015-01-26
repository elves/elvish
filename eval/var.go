package eval

import "os"

// Variable is the internal representation of a variable.
type Variable interface {
	Set(v Value)
	Get() Value
	StaticType() Type
}

type InternalVariable struct {
	valuePtr   *Value
	staticType Type
}

func newInternalVariable(v Value, t Type) Variable {
	return InternalVariable{&v, t}
}

func newInternalVariableWithType(v Value) Variable {
	return InternalVariable{&v, v.Type()}
}

func (iv InternalVariable) Set(val Value) {
	*iv.valuePtr = val
}

func (iv InternalVariable) Get() Value {
	return *iv.valuePtr
}

func (iv InternalVariable) StaticType() Type {
	return iv.staticType
}

type EnvVariable struct {
	name string
}

func newEnvVariable(name string) EnvVariable {
	return EnvVariable{name}
}

func (ev EnvVariable) Set(val Value) {

	os.Setenv(ev.name, val.String())
}

func (ev EnvVariable) Get() Value {
	return String(os.Getenv(ev.name))
}

func (ev EnvVariable) StaticType() Type {
	return StringType{}
}
