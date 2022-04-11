package testutil

import "testing"

func TestSet(t *testing.T) {
	c := &cleanuper{}
	s := "old"
	Set(c, &s, "new")
	if s != "new" {
		t.Errorf("After Set, s = %q, want %q", s, "new")
	}

	c.runCleanups()
	if s != "old" {
		t.Errorf("After Set, s = %q, want %q", s, "old")
	}
}
