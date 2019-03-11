package newedit

import (
	"testing"
)

func TestInitListing_Binding(t *testing.T) {
	// Test that the binding variable in the returned namespace indeed refers to
	// the BindingMap returned.
	_, binding, ns := initListing(&fakeEditor{})
	if ns["binding"].Get() != *binding {
		t.Errorf("The binding var in the ns is not the same as the BindingMap")
	}
}

// TODO: Test the builtin functions
