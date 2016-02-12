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

func NewPtrVariable(v Value) Variable {
	return ptrVariable{&v}
}

func (iv ptrVariable) Set(val Value) {
	*iv.valuePtr = val
}

func (iv ptrVariable) Get() Value {
	return *iv.valuePtr
}

// elemVariable is an element of a IndexSetter.
type elemVariable struct {
	container IndexSetter
	index     Value
}

func (ev elemVariable) Set(val Value) {
	ev.container.IndexSet(ev.index, val)
}

func (ev elemVariable) Get() Value {
	return ev.container.Index(ev.index)
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
