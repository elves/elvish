package eval

import (
	"errors"
	"os"

	"github.com/elves/elvish/eval/types"
)

var (
	ErrRoCannotBeSet = errors.New("read-only variable; cannot be set")
)

// Variable represents an elvish variable.
type Variable interface {
	Set(v types.Value)
	Get() types.Value
}

type ptrVariable struct {
	valuePtr  *types.Value
	validator func(types.Value) error
}

type invalidValueError struct {
	inner error
}

func (err invalidValueError) Error() string {
	return "invalid value: " + err.inner.Error()
}

func NewPtrVariable(v types.Value) Variable {
	return NewPtrVariableWithValidator(v, nil)
}

func NewPtrVariableWithValidator(v types.Value, vld func(types.Value) error) Variable {
	return ptrVariable{&v, vld}
}

func (iv ptrVariable) Set(val types.Value) {
	if iv.validator != nil {
		if err := iv.validator(val); err != nil {
			throw(invalidValueError{err})
		}
	}
	*iv.valuePtr = val
}

func (iv ptrVariable) Get() types.Value {
	return *iv.valuePtr
}

type roVariable struct {
	value types.Value
}

func NewRoVariable(v types.Value) Variable {
	return roVariable{v}
}

func (rv roVariable) Set(val types.Value) {
	throw(ErrRoCannotBeSet)
}

func (rv roVariable) Get() types.Value {
	return rv.value
}

type cbVariable struct {
	set func(types.Value)
	get func() types.Value
}

// MakeVariableFromCallback makes a variable from a set callback and a get
// callback.
func MakeVariableFromCallback(set func(types.Value), get func() types.Value) Variable {
	return &cbVariable{set, get}
}

func (cv *cbVariable) Set(val types.Value) {
	cv.set(val)
}

func (cv *cbVariable) Get() types.Value {
	return cv.get()
}

type roCbVariable func() types.Value

// MakeRoVariableFromCallback makes a read-only variable from a get callback.
func MakeRoVariableFromCallback(get func() types.Value) Variable {
	return roCbVariable(get)
}

func (cv roCbVariable) Set(types.Value) {
	throw(ErrRoCannotBeSet)
}

func (cv roCbVariable) Get() types.Value {
	return cv()
}

// elemVariable is a (arbitrary nested) element.
// XXX(xiaq): This is an ephemeral "variable" and is a bad hack.
type elemVariable struct {
	variable Variable
	assocers []types.Assocer
	indices  []types.Value
	setValue types.Value
}

func (ev *elemVariable) Set(v0 types.Value) {
	v := v0
	// Evaluate the actual new value from inside out. See comments in
	// compile_lvalue.go for how assignment of indexed variables work.
	for i := len(ev.assocers) - 1; i >= 0; i-- {
		v = ev.assocers[i].Assoc(ev.indices[i], v)
	}
	ev.variable.Set(v)
	// XXX(xiaq): Remember the set value for use in Get.
	ev.setValue = v0
}

func (ev *elemVariable) Get() types.Value {
	// XXX(xiaq): This is only called from fixNilVariables. We don't want to
	// waste time accessing the variable, so we simply return the value that was
	// set.
	return ev.setValue
}

// envVariable represents an environment variable.
type envVariable struct {
	name string
}

func (ev envVariable) Set(val types.Value) {
	os.Setenv(ev.name, types.ToString(val))
}

func (ev envVariable) Get() types.Value {
	return types.String(os.Getenv(ev.name))
}

// ErrGetBlackhole is raised when attempting to get the value of a blackhole
// variable.
var ErrGetBlackhole = errors.New("cannot get blackhole variable")

// BlackholeVariable represents a blackhole variable. Assignments to a blackhole
// variable will be discarded, and getting a blackhole variable raises an error.
type BlackholeVariable struct{}

func (bv BlackholeVariable) Set(types.Value) {}

func (bv BlackholeVariable) Get() types.Value {
	throw(ErrGetBlackhole)
	panic("unreachable")
}
