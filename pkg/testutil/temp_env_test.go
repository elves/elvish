package testutil

import (
	"os"
	"testing"
)

func TestWithTempEnv_ExistingEnv(t *testing.T) {
	envName := "ELVISH_TEST_ENV"
	os.Setenv(envName, "old value")
	defer os.Unsetenv(envName)

	restore := WithTempEnv(envName, "new value")
	if os.Getenv(envName) != "new value" {
		t.Errorf("did not set to new value")
	}
	restore()
	if os.Getenv(envName) != "old value" {
		t.Errorf("did not restore to old value")
	}
}

func TestWithTempEnv_NewEnv(t *testing.T) {
	envName := "ELVISH_TEST_ENV"
	os.Unsetenv(envName)

	restore := WithTempEnv(envName, "new value")
	if os.Getenv(envName) != "new value" {
		t.Errorf("did not set to new value")
	}
	restore()
	if _, exists := os.LookupEnv(envName); exists {
		t.Errorf("did not remove")
	}
}
