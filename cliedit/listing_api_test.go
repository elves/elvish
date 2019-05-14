package cliedit

/*
func TestInitListing_Binding(t *testing.T) {
	// Test that the binding variable in the returned namespace indeed refers to
	// the BindingMap returned.
	_, binding, ns := initListing(&fakeApp{})
	if ns["binding"].Get() != *binding {
		t.Errorf("The binding var in the ns is not the same as the BindingMap")
	}
}
*/

// TODO: Test the builtin functions. As a prerequisite, we need to make listing
// mode's state observable, and expose fakeItems and fakeAcceptableItems of the
// listing package.
