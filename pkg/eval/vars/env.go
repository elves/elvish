package vars

import (
	"errors"
	"os"
)

var errEnvMustBeString = errors.New("environment variable can only be set string values")

type EnvVariable struct {
	name string
}

func (ev EnvVariable) Set(val any) error {
	if s, ok := val.(string); ok {
		os.Setenv(ev.name, s)
		return nil
	}
	return errEnvMustBeString
}

func (ev EnvVariable) Get() any {
	return os.Getenv(ev.name)
}

func (ev EnvVariable) Unset() error {
	return os.Unsetenv(ev.name)
}

func (ev EnvVariable) IsSet() bool {
	_, ok := os.LookupEnv(ev.name)
	return ok
}

// FromEnv returns a Var corresponding to the named environment variable.
func FromEnv(name string) Var {
	return EnvVariable{name}
}
