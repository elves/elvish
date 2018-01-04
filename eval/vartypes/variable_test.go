package vartypes

import (
	"os"
	"testing"

	"github.com/elves/elvish/eval/types"
)

func TestPtrVariable(t *testing.T) {
	v := NewPtr(types.Bool(true))
	if v.Get() != types.Bool(true) {
		t.Errorf("PtrVariable.Get doesn't return initial value")
	}
	if v.Set(types.String("233")) != nil {
		t.Errorf("PtrVariable.Set errors")
	}
	if v.Get() != types.String("233") {
		t.Errorf("PtrVariable.Get doesn't return altered value")
	}
}

func TestValidatedPtrVariable(t *testing.T) {
	v := NewValidatedPtr(types.Bool(true), ShouldBeBool)
	if v.Set(types.String("233")) == nil {
		t.Errorf("ValidatedPtrVariable.Set doesn't error when setting incompatible value")
	}
}

func TestRoVariable(t *testing.T) {
	v := NewRo(types.String("haha"))
	if v.Get() != types.String("haha") {
		t.Errorf("RoVariable.Get doesn't return initial value")
	}
	if v.Set(types.String("lala")) == nil {
		t.Errorf("RoVariable.Set doesn't error")
	}
}

func TestCbVariable(t *testing.T) {
	getCalled := false
	get := func() types.Value {
		getCalled = true
		return types.String("cb")
	}
	var setCalledWith types.Value
	set := func(v types.Value) error {
		setCalledWith = v
		return nil
	}

	v := NewCallback(set, get)
	if v.Get() != types.String("cb") {
		t.Errorf("cbVariable doesn't return value from callback")
	}
	if !getCalled {
		t.Errorf("cbVariable doesn't call callback")
	}
	v.Set(types.String("setting"))
	if setCalledWith != types.String("setting") {
		t.Errorf("cbVariable.Set doesn't call setter with value")
	}
}

func TestRoCbVariable(t *testing.T) {
	getCalled := false
	get := func() types.Value {
		getCalled = true
		return types.String("cb")
	}
	v := NewRoCallback(get)
	if v.Get() != types.String("cb") {
		t.Errorf("roCbVariable doesn't return value from callback")
	}
	if !getCalled {
		t.Errorf("roCbVariable doesn't call callback")
	}
	if v.Set(types.String("lala")) == nil {
		t.Errorf("roCbVariable.Set doesn't error")
	}
}

func TestEnvVariable(t *testing.T) {
	name := "elvish_test"
	v := envVariable{name}
	os.Setenv(name, "foo")
	if v.Get() != types.String("foo") {
		t.Errorf("envVariable.Get doesn't return env value")
	}
	v.Set(types.String("bar"))
	if os.Getenv(name) != "bar" {
		t.Errorf("envVariable.Set doesn't alter env value")
	}
}
