package store

import "testing"

func TestSharedVar(t *testing.T) {
	varname := "foo"
	value1 := "lorem ipsum"
	value2 := "o mores, o tempora"

	// Getting an nonexistent variable should return ErrNoSharedVar.
	_, err := tStore.SharedVar(varname)
	if err != ErrNoSharedVar {
		t.Error("want ErrNoSharedVar, got", err)
	}

	// Setting a variable for the first time creates it.
	err = tStore.SetSharedVar(varname, value1)
	if err != nil {
		t.Error("want no error, got", err)
	}
	v, err := tStore.SharedVar(varname)
	if v != value1 || err != nil {
		t.Errorf("want %q and no error, got %q and %v", value1, v, err)
	}

	// Setting an existing variable updates its value.
	err = tStore.SetSharedVar(varname, value2)
	if err != nil {
		t.Error("want no error, got", err)
	}
	v, err = tStore.SharedVar(varname)
	if v != value2 || err != nil {
		t.Errorf("want %q and no error, got %q and %v", value2, v, err)
	}

	// After deleting a variable, access to it cause ErrNoSharedVar.
	err = tStore.DelSharedVar(varname)
	if err != nil {
		t.Error("want no error, got", err)
	}
	_, err = tStore.SharedVar(varname)
	if err != ErrNoSharedVar {
		t.Error("want ErrNoSharedVar, got", err)
	}
}
