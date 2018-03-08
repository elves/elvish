package vars

import (
	"os"
	"testing"
)

func TestAnyVariable(t *testing.T) {
	v := NewAnyWithInit(true)
	if v.Get() != true {
		t.Errorf("PtrVariable.Get doesn't return initial value")
	}
	if v.Set("233") != nil {
		t.Errorf("PtrVariable.Set errors")
	}
	if v.Get() != "233" {
		t.Errorf("PtrVariable.Get doesn't return altered value")
	}
}

func TestRoVariable(t *testing.T) {
	v := NewRo("haha")
	if v.Get() != "haha" {
		t.Errorf("RoVariable.Get doesn't return initial value")
	}
	if v.Set("lala") == nil {
		t.Errorf("RoVariable.Set doesn't error")
	}
}

func TestCbVariable(t *testing.T) {
	getCalled := false
	get := func() interface{} {
		getCalled = true
		return "cb"
	}
	var setCalledWith interface{}
	set := func(v interface{}) error {
		setCalledWith = v
		return nil
	}

	v := FromSetGet(set, get)
	if v.Get() != "cb" {
		t.Errorf("cbVariable doesn't return value from callback")
	}
	if !getCalled {
		t.Errorf("cbVariable doesn't call callback")
	}
	v.Set("setting")
	if setCalledWith != "setting" {
		t.Errorf("cbVariable.Set doesn't call setter with value")
	}
}

func TestRoCbVariable(t *testing.T) {
	getCalled := false
	get := func() interface{} {
		getCalled = true
		return "cb"
	}
	v := FromGet(get)
	if v.Get() != "cb" {
		t.Errorf("roCbVariable doesn't return value from callback")
	}
	if !getCalled {
		t.Errorf("roCbVariable doesn't call callback")
	}
	if v.Set("lala") == nil {
		t.Errorf("roCbVariable.Set doesn't error")
	}
}

func TestEnvVariable(t *testing.T) {
	name := "elvish_test"
	v := envVariable{name}
	os.Setenv(name, "foo")
	if v.Get() != "foo" {
		t.Errorf("envVariable.Get doesn't return env value")
	}
	v.Set("bar")
	if os.Getenv(name) != "bar" {
		t.Errorf("envVariable.Set doesn't alter env value")
	}
}
