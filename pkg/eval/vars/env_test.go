package vars

import (
	"os"
	"testing"
)

func TestFromEnv(t *testing.T) {
	name := "elvish_test"
	v := FromEnv(name)

	if v.Get() != nil {
		t.Errorf("get non-exist envVariable doesn't return nil")
	}

	os.Setenv(name, "foo")
	if v.Get() != "foo" {
		t.Errorf("envVariable.Get doesn't return env value")
	}

	err := v.Set("bar")
	if err != nil || os.Getenv(name) != "bar" {
		t.Errorf("envVariable.Set doesn't alter env value")
	}

	err = v.Set(true)
	if err != errEnvMustBeString {
		t.Errorf("envVariable.Set to a non-string value didn't return an error")
	}

	err = v.Set(nil)
	if err != nil {
		t.Errorf("envVariable.Set(nil) failed")
	}
	if _, exist := os.LookupEnv(name); exist {
		t.Errorf("envVariable.Set(nil) doesn't unset env value")
	}
}
