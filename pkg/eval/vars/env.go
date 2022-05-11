package vars

import (
	"errors"
	"os"
)

var errEnvMustBeString = errors.New("environment variable can only be set string values")

type envVariable struct {
	name string
}

func (ev envVariable) Set(val any) error {
	if val == nil {
		return os.Unsetenv(ev.name)
	}

	if s, ok := val.(string); ok {
		os.Setenv(ev.name, s)
		return nil
	}
	return errEnvMustBeString
}

func (ev envVariable) Get() any {
	if v, exist := os.LookupEnv(ev.name); exist {
		return v
	}

	return nil
}

// FromEnv returns a Var corresponding to the named environment variable.
func FromEnv(name string) Var {
	return envVariable{name}
}
