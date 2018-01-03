package eval

import (
	"errors"
	"os"

	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/eval/vartypes"
)

// elemVariable is a (arbitrary nested) element.
// XXX(xiaq): This is an ephemeral "variable" and is a bad hack.
type elemVariable struct {
	variable vartypes.Variable
	assocers []types.Assocer
	indices  []types.Value
	setValue types.Value
}

func (ev *elemVariable) Set(v0 types.Value) error {
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

func (ev envVariable) Set(val types.Value) error {
	os.Setenv(ev.name, types.ToString(val))
	return nil
}

func (ev envVariable) Get() types.Value {
	return types.String(os.Getenv(ev.name))
}

// ErrGetBlackhole is raised when attempting to get the value of a blackhole
// variable.
var ErrGetBlackhole = errors.New("cannot get blackhole variable")

// BlackholeVariable represents a blackhole variable. Assignments to a blackhole
// variable will be discarded, and getting a blackhole variable always returns
// an empty string.
type BlackholeVariable struct{}

func (bv BlackholeVariable) Set(types.Value) error {
	return nil
}

func (bv BlackholeVariable) Get() types.Value {
	// TODO: Return a special placeholder value.
	return types.String("")
}
