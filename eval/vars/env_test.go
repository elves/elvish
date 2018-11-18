package vars

import (
	"os"
	"testing"
)

func TestFromEnv(t *testing.T) {
	name := "elvish_test"
	v := FromEnv(name)
	os.Setenv(name, "foo")
	if v.Get() != "foo" {
		t.Errorf("envVariable.Get doesn't return env value")
	}
	v.Set("bar")
	if os.Getenv(name) != "bar" {
		t.Errorf("envVariable.Set doesn't alter env value")
	}
}
