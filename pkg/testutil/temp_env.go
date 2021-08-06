package testutil

import "os"

// Setenv sets the value of an environment variable for the duration of a test.
// It returns value.
func Setenv(c Cleanuper, name, value string) string {
	SaveEnv(c, name)
	os.Setenv(name, value)
	return value
}

// Setenv unsets an environment variable for the duration of a test.
func Unsetenv(c Cleanuper, name string) {
	SaveEnv(c, name)
	os.Unsetenv(name)
}

// SaveEnv saves the current value of an environment variable so that it will be
// restored after a test has finished.
func SaveEnv(c Cleanuper, name string) {
	oldValue, existed := os.LookupEnv(name)
	if existed {
		c.Cleanup(func() { os.Setenv(name, oldValue) })
	} else {
		c.Cleanup(func() { os.Unsetenv(name) })
	}
}
