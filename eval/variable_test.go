package eval

import (
	"os"
	"testing"

	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/util"
)

func TestPtrVariable(t *testing.T) {
	v := NewPtrVariable(types.Bool(true))
	if v.Get() != types.Bool(true) {
		t.Errorf("PtrVariable.Get doesn't return initial value")
	}
	v.Set(String("233"))
	if v.Get() != String("233") {
		t.Errorf("PtrVariable.Get doesn't return altered value")
	}

	v = NewPtrVariableWithValidator(types.Bool(true), ShouldBeBool)
	if util.DoesntThrow(func() { v.Set(String("233")) }) {
		t.Errorf("PtrVariable.Set doesn't error when setting incompatible value")
	}
}

func TestRoVariable(t *testing.T) {
	v := NewRoVariable(String("haha"))
	if v.Get() != String("haha") {
		t.Errorf("RoVariable.Get doesn't return initial value")
	}
	if util.DoesntThrow(func() { v.Set(String("lala")) }) {
		t.Errorf("RoVariable.Set doesn't error")
	}
}

func TestCbVariable(t *testing.T) {
	getCalled := false
	get := func() types.Value {
		getCalled = true
		return String("cb")
	}
	var setCalledWith types.Value
	set := func(v types.Value) {
		setCalledWith = v
	}

	v := MakeVariableFromCallback(set, get)
	if v.Get() != String("cb") {
		t.Errorf("cbVariable doesn't return value from callback")
	}
	if !getCalled {
		t.Errorf("cbVariable doesn't call callback")
	}
	v.Set(String("setting"))
	if setCalledWith != String("setting") {
		t.Errorf("cbVariable.Set doesn't call setter with value")
	}
}

func TestRoCbVariable(t *testing.T) {
	getCalled := false
	get := func() types.Value {
		getCalled = true
		return String("cb")
	}
	v := MakeRoVariableFromCallback(get)
	if v.Get() != String("cb") {
		t.Errorf("roCbVariable doesn't return value from callback")
	}
	if !getCalled {
		t.Errorf("roCbVariable doesn't call callback")
	}
	if util.DoesntThrow(func() { v.Set(String("lala")) }) {
		t.Errorf("roCbVariable.Set doesn't error")
	}
}

func TestEnvVariable(t *testing.T) {
	name := "elvish_test"
	v := envVariable{name}
	os.Setenv(name, "foo")
	if v.Get() != String("foo") {
		t.Errorf("envVariable.Get doesn't return env value")
	}
	v.Set(String("bar"))
	if os.Getenv(name) != "bar" {
		t.Errorf("envVariable.Set doesn't alter env value")
	}
}
