package util

import (
	"os"
	"testing"
)

func TestWithTempEnv(t *testing.T) {
	envName := "ELVISH_TEST_ENV"
	os.Setenv(envName, "old value")

	restore := WithTempEnv(envName, "new value")
	if os.Getenv(envName) != "new value" {
		t.Errorf("did not set to new value")
	}
	restore()
	if os.Getenv(envName) != "old value" {
		t.Errorf("did not restore to old value")
	}
}
