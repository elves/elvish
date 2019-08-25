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

/*
func TestHistlist_Start(t *testing.T) {
	ed := &fakeApp{}
	ev := eval.NewEvaler()
	lsMode := listing.Mode{}
	lsBinding := emptyBindingMap
	// TODO: Move this into common setup.
	histFuser, err := histutil.NewFuser(testStore)
	if err != nil {
		panic(err)
	}

	ns := initHistlist(ed, ev, histFuser.AllCmds, &lsMode, &lsBinding)

	// Call <edit:histlist>:start.
	fm := eval.NewTopFrame(ev, eval.NewInternalSource("[test]"), nil)
	fm.Call(getFn(ns, "start"), eval.NoArgs, eval.NoOpts)

	// Verify that the current mode supports listing.
	lister, ok := ed.state.Mode().(clitypes.Lister)
	if !ok {
		t.Errorf("Mode is not Lister after <edit:histlist>:start")
	}
	// Verify the actual listing.
	buf := ui.Render(lister.List(10), 30)
	wantBuf := ui.NewBufferBuilder(30).
		WriteString("   0 echo hello world", "7").Buffer()
	if !reflect.DeepEqual(buf, wantBuf) {
		t.Errorf("Rendered listing is %v, want %v", buf, wantBuf)
	}
}

func TestInitLastCmd_Start(t *testing.T) {
	ed := &fakeApp{}
	ev := eval.NewEvaler()
	lsMode := listing.Mode{}
	lsBinding := emptyBindingMap

	ns := initLastcmd(ed, ev, testStore, &lsMode, &lsBinding)

	// Call <edit:listing>:start.
	fm := eval.NewTopFrame(ev, eval.NewInternalSource("[test]"), nil)
	fm.Call(getFn(ns, "start"), nil, eval.NoOpts)

	// Verify that the current mode supports listing.
	lister, ok := ed.state.Mode().(clitypes.Lister)
	if !ok {
		t.Errorf("Mode is not Lister after <edit:lastcmd>:start")
	}
	// Verify the listing.
	buf := ui.Render(lister.List(10), 20)
	wantBuf := ui.NewBufferBuilder(20).
		WriteString("    echo hello world", "7").Newline().
		WriteUnstyled("  0 echo").Newline().
		WriteUnstyled("  1 hello").Newline().
		WriteUnstyled("  2 world").Buffer()
	if !reflect.DeepEqual(buf, wantBuf) {
		t.Errorf("Rendered listing is %v, want %v", buf, wantBuf)
	}
}

func TestLocation_Start(t *testing.T) {
	ed := &fakeApp{}
	ev := eval.NewEvaler()
	lsMode := listing.Mode{}
	lsBinding := emptyBindingMap
	getDirs := func() ([]storedefs.Dir, error) {
		return []storedefs.Dir{
			{Path: "/usr/bin", Score: 20},
			{Path: "/home/elf", Score: 10},
		}, nil
	}
	cd := func(string) error { return nil }

	ns := initLocation(ed, ev, getDirs, cd, &lsMode, &lsBinding)

	// Call <edit:location>:start.
	fm := eval.NewTopFrame(ev, eval.NewInternalSource("[test]"), nil)
	fm.Call(getFn(ns, "start"), eval.NoArgs, eval.NoOpts)

	// Verify that the current mode supports listing.
	lister, ok := ed.state.Mode().(clitypes.Lister)
	if !ok {
		t.Errorf("Mode is not Lister after <edit:location>:start")
	}
	// Verify the actual listing.
	buf := ui.Render(lister.List(10), 30)
	wantBuf := ui.NewBufferBuilder(30).
		WriteString(" 20 /usr/bin", "7").Newline().
		WriteString(" 10 /home/elf", "").Buffer()
	if !reflect.DeepEqual(buf, wantBuf) {
		t.Errorf("Rendered listing is %v, want %v", buf, wantBuf)
	}
}
*/
