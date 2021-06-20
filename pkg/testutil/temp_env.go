package testutil

import "os"

// WithTempEnv sets an environment variable to a temporary value, and returns a
// function for restoring the old value.
func WithTempEnv(name, value string) func() {
	oldValue, existed := os.LookupEnv(name)
	os.Setenv(name, value)
	return func() {
		if existed {
			os.Setenv(name, oldValue)
		} else {
			os.Unsetenv(name)
		}
	}
}
