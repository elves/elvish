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
	if v.Set(string("233")) != nil {
		t.Errorf("PtrVariable.Set errors")
	}
	if v.Get() != string("233") {
		t.Errorf("PtrVariable.Get doesn't return altered value")
	}
}

func TestValidatedPtrVariable(t *testing.T) {
	v := NewValidatedPtr(types.Bool(true), ShouldBeBool)
	if v.Set(string("233")) == nil {
		t.Errorf("ValidatedPtrVariable.Set doesn't error when setting incompatible value")
	}
}

func TestRoVariable(t *testing.T) {
	v := NewRo(string("haha"))
	if v.Get() != string("haha") {
		t.Errorf("RoVariable.Get doesn't return initial value")
	}
	if v.Set(string("lala")) == nil {
		t.Errorf("RoVariable.Set doesn't error")
	}
}

func TestCbVariable(t *testing.T) {
	getCalled := false
	get := func() types.Value {
		getCalled = true
		return string("cb")
	}
	var setCalledWith types.Value
	set := func(v types.Value) error {
		setCalledWith = v
		return nil
	}

	v := NewCallback(set, get)
	if v.Get() != string("cb") {
		t.Errorf("cbVariable doesn't return value from callback")
	}
	if !getCalled {
		t.Errorf("cbVariable doesn't call callback")
	}
	v.Set(string("setting"))
	if setCalledWith != string("setting") {
		t.Errorf("cbVariable.Set doesn't call setter with value")
	}
}

func TestRoCbVariable(t *testing.T) {
	getCalled := false
	get := func() types.Value {
		getCalled = true
		return string("cb")
	}
	v := NewRoCallback(get)
	if v.Get() != string("cb") {
		t.Errorf("roCbVariable doesn't return value from callback")
	}
	if !getCalled {
		t.Errorf("roCbVariable doesn't call callback")
	}
	if v.Set(string("lala")) == nil {
		t.Errorf("roCbVariable.Set doesn't error")
	}
}

func TestEnvVariable(t *testing.T) {
	name := "elvish_test"
	v := envVariable{name}
	os.Setenv(name, "foo")
	if v.Get() != string("foo") {
		t.Errorf("envVariable.Get doesn't return env value")
	}
	v.Set(string("bar"))
	if os.Getenv(name) != "bar" {
		t.Errorf("envVariable.Set doesn't alter env value")
	}
}
