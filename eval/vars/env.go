package vars

import (
	"errors"
	"os"
)

var errEnvMustBeString = errors.New("environment variable can only be set string values")

// envVariable represents an environment variable.
type envVariable struct {
	name string
}

func (ev envVariable) Set(val interface{}) error {
	if s, ok := val.(string); ok {
		os.Setenv(ev.name, string(s))
		return nil
	}
	return errEnvMustBeString
}

func (ev envVariable) Get() interface{} {
	return string(os.Getenv(ev.name))
}

// NewEnv returns an environment variable.
func NewEnv(name string) Var {
	return envVariable{name}
}
