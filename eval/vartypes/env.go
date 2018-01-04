package vartypes

import (
	"os"

	"github.com/elves/elvish/eval/types"
)

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

func NewEnv(name string) Variable {
	return envVariable{name}
}
