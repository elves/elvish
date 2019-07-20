package vars

import (
	"errors"
	"os"
)

var errEnvMustBeString = errors.New("environment variable can only be set string values")

type envVariable struct {
	name string
}

func (ev envVariable) Set(val interface{}) error {
	if s, ok := val.(string); ok {
		os.Setenv(ev.name, s)
		return nil
	}
	return errEnvMustBeString
}

func (ev envVariable) Get() interface{} {
	return os.Getenv(ev.name)
}

// FromEnv returns a Var corresponding to the named environment variable.
func FromEnv(name string) Var {
	return envVariable{name}
}
