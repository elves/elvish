package util

import "os"

// WithTempEnv sets an environment variable to a temporary value, and returns a
// function for restoring the old value.
func WithTempEnv(name, value string) func() {
	oldValue := os.Getenv(name)
	os.Setenv(name, value)
	return func() { os.Setenv(name, oldValue) }
}
