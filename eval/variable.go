package eval

import (
	"errors"
	"os"
	"strings"
)

var (
	ErrRoCannotBeSet = errors.New("read-only; cannot be set")
)

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

type roVariable struct {
	value Value
}

func NewRoVariable(v Value) Variable {
	return roVariable{v}
}

func (rv roVariable) Set(val Value) {
	throw(ErrRoCannotBeSet)
}

func (rv roVariable) Get() Value {
	return rv.value
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
	return ev.container.IndexOne(ev.index)
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

type pathEnvVariable struct {
	envVariable
	ppaths *[]string
}

func (pev pathEnvVariable) Set(val Value) {
	s := ToString(val)
	os.Setenv(pev.name, s)
	paths := strings.Split(s, ":")
	*pev.ppaths = paths
}
