package mode

import (
	"errors"
	"testing"

	"src.elv.sh/pkg/cli"
	. "src.elv.sh/pkg/cli/clitest"
	"src.elv.sh/pkg/cli/lscolors"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/testutil"
	"src.elv.sh/pkg/ui"
)

var testDir = testutil.Dir{
	"a": "",
	"d": testutil.Dir{
		"d1": "content\td1\nline 2",
		"d2": testutil.Dir{
			"d21":     "content d21",
			"d22":     "content d22",
			"d23.png": "",
		},
		"d3":  testutil.Dir{},
		".dh": "hidden",
	},
	"f": "",
}

func TestErrorInAscend(t *testing.T) {
	f := Setup()
	defer f.Stop()

	c := getTestCursor()
	c.ascendErr = errors.New("cannot ascend")
	startNavigation(f.App, NavigationSpec{Cursor: c})

	f.TTY.Inject(term.K(ui.Left))
	f.TestTTYNotes(t, "cannot ascend")
}

func TestErrorInDescend(t *testing.T) {
	f := Setup()
	defer f.Stop()

	c := getTestCursor()
	c.descendErr = errors.New("cannot descend")
	startNavigation(f.App, NavigationSpec{Cursor: c})

	f.TTY.Inject(term.K(ui.Down))
	f.TTY.Inject(term.K(ui.Right))
	f.TestTTYNotes(t, "cannot descend")
}

func TestErrorInCurrent(t *testing.T) {
	f, cleanup := setup()
	defer cleanup()
	defer f.Stop()

	c := getTestCursor()
	c.currentErr = errors.New("ERR")
	startNavigation(f.App, NavigationSpec{Cursor: c})

	buf := f.MakeBuffer(
		"", term.DotHere, "\n",
		" NAVIGATING  \n", Styles,
		"************ ",
		" a   ERR            \n", Styles,
		"     !!!",
		" d  \n", Styles,
		"////",
		" f  ",
	)

	f.TTY.TestBuffer(t, buf)

	// Test that Right does nothing.
	f.TTY.Inject(term.K(ui.Right))
	f.TTY.TestBuffer(t, buf)
}

func TestErrorInParent(t *testing.T) {
	f, cleanup := setup()
	defer cleanup()
	defer f.Stop()

	c := getTestCursor()
	c.parentErr = errors.New("ERR")
	startNavigation(f.App, NavigationSpec{Cursor: c})

	f.TestTTY(t,
		"", term.DotHere, "\n",
		" NAVIGATING  \n", Styles,
		"************ ",
		"ERR   d1            content    d1\n", Styles,
		"!!!  ++++++++++++++",
		"      d2            line 2\n", Styles,
		"     //////////////",
		"      d3           ", Styles,
		"     //////////////",
	)
}

func TestWidthRatio(t *testing.T) {
	f, cleanup := setup()
	defer cleanup()
	defer f.Stop()

	c := getTestCursor()
	startNavigation(f.App, NavigationSpec{
		Cursor:     c,
		WidthRatio: func() [3]int { return [3]int{1, 1, 1} },
	})

	f.TestTTY(t,
		"", term.DotHere, "\n",
		" NAVIGATING  \n", Styles,
		"************ ",
		" a            d1           content    d1\n", Styles,
		"             +++++++++++++",
		" d            d2           line 2\n", Styles,
		"############ /////////////",
		" f            d3          ", Styles,
		"             /////////////",
	)
}

func TestGetSelectedName(t *testing.T) {
	f := Setup()
	defer f.Stop()

	w := startNavigation(f.App, NavigationSpec{Cursor: getTestCursor()})

	wantName := "d1"
	if name := w.SelectedName(); name != wantName {
		t.Errorf("Got name %q, want %q", name, wantName)
	}
}

func TestNavigation_FakeFS(t *testing.T) {
	cursor := getTestCursor()
	testNavigation(t, cursor)
}

func TestNavigation_RealFS(t *testing.T) {
	_, cleanupFs := testutil.InTestDir()
	defer cleanupFs()
	testutil.ApplyDir(testDir)

	testutil.MustChdir("d")
	testNavigation(t, nil)
}

func testNavigation(t *testing.T, c NavigationCursor) {
	f, cleanup := setup()
	defer cleanup()
	defer f.Stop()

	w := startNavigation(f.App, NavigationSpec{Cursor: c})

	// Test initial UI and file preview.
	// NOTE: Buffers are named after the file that is now being selected.
	d1Buf := f.MakeBuffer(
		"", term.DotHere, "\n",
		" NAVIGATING  \n", Styles,
		"************ ",
		" a    d1            content    d1\n", Styles,
		"     ++++++++++++++",
		" d    d2            line 2\n", Styles,
		"#### //////////////",
		" f    d3           ", Styles,
		"     //////////////",
	)
	f.TTY.TestBuffer(t, d1Buf)

	// Test scrolling of preview.
	w.ScrollPreview(1)
	f.App.Redraw()
	d1Buf2 := f.MakeBuffer(
		"", term.DotHere, "\n",
		" NAVIGATING  \n", Styles,
		"************ ",
		" a    d1            line 2             │\n", Styles,
		"     ++++++++++++++                    -",
		" d    d2                               │\n", Styles,
		"#### //////////////                    -",
		" f    d3                                \n", Styles,
		"     //////////////                    X",
		"                                        ", Styles,
		"                                       X",
	)
	f.TTY.TestBuffer(t, d1Buf2)

	// Test handling of selection change and directory preview. Also test
	// LS_COLORS.
	w.Select(tk.Next)
	f.App.Redraw()
	d2Buf := f.MakeBuffer(
		"", term.DotHere, "\n",
		" NAVIGATING  \n", Styles,
		"************ ",
		" a    d1             d21                \n", Styles,
		"                    ++++++++++++++++++++",
		" d    d2             d22                \n", Styles,
		"#### ##############",
		" f    d3             d23.png            ", Styles,
		"     ////////////// !!!!!!!!!!!!!!!!!!!!",
	)
	f.TTY.TestBuffer(t, d2Buf)

	// Test handling of Descend.
	w.Descend()
	f.App.Redraw()
	d21Buf := f.MakeBuffer(
		"", term.DotHere, "\n",
		" NAVIGATING  \n", Styles,
		"************ ",
		" d1   d21           content d21\n", Styles,
		"     ++++++++++++++",
		" d2   d22          \n", Styles,
		"####",
		" d3   d23.png      ", Styles,
		"//// !!!!!!!!!!!!!!",
	)
	f.TTY.TestBuffer(t, d21Buf)

	// Test handling of Ascend, and that the current column selects the
	// directory we just ascended from, thus reverting to wantBuf1.
	w.Ascend()
	f.App.Redraw()
	f.TTY.TestBuffer(t, d2Buf)

	// Test handling of Descend on a regular file, i.e. do nothing. First move
	// the cursor to d1, which is a regular file.
	w.Select(tk.Prev)
	f.App.Redraw()
	f.TTY.TestBuffer(t, d1Buf)
	// Now descend, and verify that the buffer has not changed.
	w.Descend()
	f.App.Redraw()
	f.TTY.TestBuffer(t, d1Buf)

	// Test showing hidden.
	w.MutateShowHidden(func(bool) bool { return true })
	f.App.Redraw()
	f.TestTTY(t,
		"", term.DotHere, "\n",
		" NAVIGATING (show hidden)  \n", Styles,
		"************************** ",
		" a    .dh           content    d1\n",
		" d    d1            line 2\n", Styles,
		"#### ++++++++++++++",
		" f    d2           \n", Styles,
		"     //////////////",
		"      d3           ", Styles,
		"     //////////////",
	)
	w.MutateShowHidden(func(bool) bool { return false })

	// Test filtering; current column shows d1, d2, d3 before filtering.
	w.MutateFiltering(func(bool) bool { return true })
	f.TTY.Inject(term.K('3'))
	f.TestTTY(t,
		"\n",
		" NAVIGATING  3", Styles,
		"************  ", term.DotHere, "\n",
		" a    d3            \n", Styles,
		"     ##############",
		" d  \n", Styles,
		"####",
		" f  ",
	)
	w.MutateFiltering(func(bool) bool { return false })

	// Now move into d3, an empty directory. Test that the filter has been
	// cleared.
	w.Select(tk.Next)
	w.Select(tk.Next)
	w.Descend()
	f.App.Redraw()
	d3NoneBuf := f.MakeBuffer(
		"", term.DotHere, "\n",
		" NAVIGATING  \n", Styles,
		"************ ",
		" d1                 \n",
		" d2 \n", Styles,
		"////",
		" d3 ", Styles,
		"####",
	)
	f.TTY.TestBuffer(t, d3NoneBuf)
	// Test that selecting the previous does nothing in an empty directory.
	w.Select(tk.Prev)
	f.App.Redraw()
	f.TTY.TestBuffer(t, d3NoneBuf)
	// Test that selecting the next does nothing in an empty directory.
	w.Select(tk.Next)
	f.App.Redraw()
	f.TTY.TestBuffer(t, d3NoneBuf)
	// Test that Descend does nothing in an empty directory.
	w.Descend()
	f.App.Redraw()
	f.TTY.TestBuffer(t, d3NoneBuf)
}

func setup() (*Fixture, func()) {
	restore := lscolors.WithTestLsColors()
	return Setup(WithTTY(func(tty TTYCtrl) { tty.SetSize(6, 40) })), restore
}

func startNavigation(app cli.App, spec NavigationSpec) Navigation {
	w := NewNavigation(app, spec)
	startMode(app, w, nil)
	return w
}

func getTestCursor() *testCursor {
	return &testCursor{root: testDir, pwd: []string{"d"}}
}
