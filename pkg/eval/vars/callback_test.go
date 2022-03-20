package vars

import "testing"

func TestFromSetGet(t *testing.T) {
	getCalled := false
	get := func() any {
		getCalled = true
		return "cb"
	}
	var setCalledWith any
	set := func(v any) error {
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

func TestFromGet(t *testing.T) {
	getCalled := false
	get := func() any {
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
