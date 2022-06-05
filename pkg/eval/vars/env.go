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
	if s, ok := val.(string); ok {
		os.Setenv(ev.name, s)
		return nil
	}
	return errEnvMustBeString
}

func (ev envVariable) Get() any {
	return os.Getenv(ev.name)
}

func (ev envVariable) Unset() error {
	return os.Unsetenv(ev.name)
}

func (ev envVariable) IsSet() bool {
	_, ok := os.LookupEnv(ev.name)
	return ok
}

// FromEnv returns a Var corresponding to the named environment variable.
func FromEnv(name string) UnsettableVar {
	return envVariable{name}
}
