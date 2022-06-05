package vars

import (
	"os"
	"testing"

	"src.elv.sh/pkg/testutil"
)

func TestEnvVariable(t *testing.T) {
	name := "elvish_test"
	testutil.Unsetenv(t, name)

	v := FromEnv(name).(envVariable)

	if set := v.IsSet(); set != false {
		t.Errorf("EnvVariable.Set returns true for unset env variable")
	}

	err := v.Set("foo")
	if err != nil || os.Getenv(name) != "foo" {
		t.Errorf("EnvVariable.Set doesn't alter env value")
	}

	if set := v.IsSet(); set != true {
		t.Errorf("EnvVariable.Set returns false for set env variable")
	}

	err = v.Set(true)
	if err != errEnvMustBeString {
		t.Errorf("envVariable.Set to a non-string value didn't return an error")
	}

	os.Setenv(name, "bar")
	if v.Get() != "bar" {
		t.Errorf("EnvVariable.Get doesn't return value set elsewhere")
	}
}
