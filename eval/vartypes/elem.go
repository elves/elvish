package vartypes

import (
	"github.com/elves/elvish/eval/types"
)

type elem struct {
	variable Variable
	assocers []types.Assocer
	indices  []types.Value
	setValue types.Value
}

func (ev *elem) Set(v0 types.Value) error {
	v := v0
	// Evaluate the actual new value from inside out. See comments in
	// compile_lvalue.go for how assignment of indexed variables work.
	for i := len(ev.assocers) - 1; i >= 0; i-- {
		v = ev.assocers[i].Assoc(ev.indices[i], v)
	}
	err := ev.variable.Set(v)
	// XXX(xiaq): Remember the set value for use in Get.
	ev.setValue = v0
	return err
}

func (ev *elem) Get() types.Value {
	// XXX(xiaq): This is only called from fixNilVariables. We don't want to
	// waste time accessing the variable, so we simply return the value that was
	// set.
	return ev.setValue
}

// NewEleme represents an ephemeral variable that represents arbitrary nested
// elements.
func NewElem(v Variable, a []types.Assocer, i []types.Value) Variable {
	return &elem{v, a, i, types.String("")}
}

// GetUnderlyingOfEleme gets the underlying variable from an element variable,
// or nil if the argument is not an element variable.
func GetUnderlyingOfElem(v Variable) Variable {
	if ev, ok := v.(*elem); ok {
		return ev.variable
	}
	return nil
}
