package eval

import (
	"errors"
	"os"
)

// ErrNonExistentEnvVar is raised by the get-env command when the environment
// variable does not exist.
var ErrNonExistentEnvVar = errors.New("non-existent environment variable")

func init() {
	addBuiltinFns(map[string]any{
		"has-env":   hasEnv,
		"get-env":   getEnv,
		"set-env":   os.Setenv,
		"unset-env": os.Unsetenv,
	})
}

func hasEnv(key string) bool {
	_, ok := os.LookupEnv(key)
	return ok
}

func getEnv(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", ErrNonExistentEnvVar
	}
	return value, nil
}
