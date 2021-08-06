package testutil

import (
	"os"
	"testing"
)

const envName = "ELVISH_TEST_ENV"

func TestSetenv_ExistingEnv(t *testing.T) {
	os.Setenv(envName, "old value")
	defer os.Unsetenv(envName)

	c := &cleanuper{}
	v := Setenv(c, envName, "new value")

	if v != "new value" {
		t.Errorf("did not return new value")
	}
	if os.Getenv(envName) != "new value" {
		t.Errorf("did not set to new value")
	}
	c.runCleanups()
	if os.Getenv(envName) != "old value" {
		t.Errorf("did not restore to old value")
	}
}

func TestSetenv_NewEnv(t *testing.T) {
	os.Unsetenv(envName)

	c := &cleanuper{}
	v := Setenv(c, envName, "new value")

	if v != "new value" {
		t.Errorf("did not return new value")
	}
	if os.Getenv(envName) != "new value" {
		t.Errorf("did not set to new value")
	}
	c.runCleanups()
	if _, exists := os.LookupEnv(envName); exists {
		t.Errorf("did not remove")
	}
}

func TestUnsetenv_ExistingEnv(t *testing.T) {
	os.Setenv(envName, "old value")
	defer os.Unsetenv(envName)

	c := &cleanuper{}
	Unsetenv(c, envName)

	if _, exists := os.LookupEnv(envName); exists {
		t.Errorf("did not unset")
	}
	c.runCleanups()
	if os.Getenv(envName) != "old value" {
		t.Errorf("did not restore to old value")
	}
}

func TestUnsetenv_NewEnv(t *testing.T) {
	os.Unsetenv(envName)

	c := &cleanuper{}
	Unsetenv(c, envName)

	if _, exists := os.LookupEnv(envName); exists {
		t.Errorf("did not unset")
	}
	c.runCleanups()
	if _, exists := os.LookupEnv(envName); exists {
		t.Errorf("did not remove")
	}
}

// SaveEnv tested as a dependency of Setenv and Unsetenv
