package vartypes

import (
	"errors"
	"os"

	"github.com/elves/elvish/eval/types"
)

var errEnvMustBeString = errors.New("environment variable can only be set string values")

// envVariable represents an environment variable.
type envVariable struct {
	name string
}

func (ev envVariable) Set(val types.Value) error {
	if s, ok := val.(types.String); ok {
		os.Setenv(ev.name, string(s))
		return nil
	}
	return errEnvMustBeString
}

func (ev envVariable) Get() types.Value {
	return types.String(os.Getenv(ev.name))
}

// NewEnv returns an environment variable.
func NewEnv(name string) Variable {
	return envVariable{name}
}
