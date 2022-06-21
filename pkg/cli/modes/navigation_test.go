package modes

import (
	"errors"
	"testing"

	"src.elv.sh/pkg/cli"
	. "src.elv.sh/pkg/cli/clitest"
	"src.elv.sh/pkg/cli/lscolors"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/testutil"
	"src.elv.sh/pkg/ui"
)

var testDir = testutil.Dir{
	"a": "",
	"d": testutil.Dir{
		"d1": "content\td1\nline 2",
		"d2": testutil.Dir{
			"d21":       "content d21",
			"d22":       "content d22",
			"other.png": "",
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
	f.TestTTYNotes(t,
		"error: cannot ascend", Styles,
		"!!!!!!")
}

func TestErrorInDescend(t *testing.T) {
	f := Setup()
	defer f.Stop()

	c := getTestCursor()
	c.descendErr = errors.New("cannot descend")
	startNavigation(f.App, NavigationSpec{Cursor: c})

	f.TTY.Inject(term.K(ui.Down))
	f.TTY.Inject(term.K(ui.Right))
	f.TestTTYNotes(t,
		"error: cannot descend", Styles,
		"!!!!!!")
}

func TestErrorInCurrent(t *testing.T) {
	f := setupNav(t)
	defer f.Stop()

	c := getTestCursor()
	c.currentErr = errors.New("ERR")
	startNavigation(f.App, NavigationSpec{Cursor: c})

	f.TestTTY(t,
		"", term.DotHere, "\n",
		" NAVIGATING  \n", Styles,
		"************ ",
		" a   ERR            \n", Styles,
		"     !!!",
		" d  \n", Styles,
		"////",
		" f  ",
	)

	// Test that Right does nothing.
	f.TTY.Inject(term.K(ui.Right))
	// We can't just test that the buffer hasn't changed, because that might
	// capture the state of the buffer before the Right key is handled. Instead
	// we inject a key and test the result of that instead, to ensure that the
	// Right key had no effect.
	f.TTY.Inject(term.K('x'))
	f.TestTTY(t,
		"x", term.DotHere, "\n",
		" NAVIGATING  \n", Styles,
		"************ ",
		" a   ERR            \n", Styles,
		"     !!!",
		" d  \n", Styles,
		"////",
		" f  ",
	)
}

func TestErrorInParent(t *testing.T) {
	f := setupNav(t)
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
	f := setupNav(t)
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

func TestNavigation_SelectedName(t *testing.T) {
	f := Setup()
	defer f.Stop()

	w := startNavigation(f.App, NavigationSpec{Cursor: getTestCursor()})

	wantName := "d1"
	if name := w.SelectedName(); name != wantName {
		t.Errorf("Got name %q, want %q", name, wantName)
	}
}

func TestNavigation_SelectedName_EmptyDirectory(t *testing.T) {
	f := Setup()
	defer f.Stop()

	cursor := &testCursor{
		root: testutil.Dir{"d": testutil.Dir{}},
		pwd:  []string{"d"}}
	w := startNavigation(f.App, NavigationSpec{Cursor: cursor})

	wantName := ""
	if name := w.SelectedName(); name != wantName {
		t.Errorf("Got name %q, want %q", name, wantName)
	}
}

func TestNavigation_FakeFS(t *testing.T) {
	cursor := getTestCursor()
	testNavigation(t, cursor)
}

func TestNavigation_RealFS(t *testing.T) {
	testutil.InTempDir(t)
	testutil.ApplyDir(testDir)

	must.Chdir("d")
	testNavigation(t, nil)
}

func testNavigation(t *testing.T, c NavigationCursor) {
	f := setupNav(t)
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
		" f    d3                                ", Styles,
		"     //////////////                    X",
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
		" f    d3             other.png          ", Styles,
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
		" d3   other.png    ", Styles,
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

	// Test filtering; current column shows d1, d2, d3 before filtering, and
	// only shows d2 after filtering.
	w.MutateFiltering(func(bool) bool { return true })
	f.TTY.Inject(term.K('2'))
	dFilter2Buf := f.MakeBuffer(
		"\n",
		" NAVIGATING  2", Styles,
		"************  ", term.DotHere, "\n",
		" a    d2             d21                \n", Styles,
		"     ############## ++++++++++++++++++++",
		" d                   d22                \n", Styles,
		"####",
		" f                   other.png          ", Styles,
		"                    !!!!!!!!!!!!!!!!!!!!",
	)
	f.TTY.TestBuffer(t, dFilter2Buf)

	// Unbound key while filtering is ignored.
	f.TTY.Inject(term.K('a', ui.Alt))
	f.TTY.TestBuffer(t, dFilter2Buf)
	w.MutateFiltering(func(bool) bool { return false })

	// Now move into d2, and test that the filter has been cleared when
	// descending.
	w.Descend()
	f.App.Redraw()
	f.TTY.TestBuffer(t, d21Buf)

	// Apply a filter within d2.
	w.MutateFiltering(func(bool) bool { return true })
	f.TTY.Inject(term.K('2'))
	f.TestTTY(t,
		"\n",
		" NAVIGATING  2", Styles,
		"************  ", term.DotHere, "\n",
		" d1   d21           content d21\n", Styles,
		"     ++++++++++++++",
		" d2   d22          \n", Styles,
		"####",
		" d3 ", Styles,
		"////",
	)
	w.MutateFiltering(func(bool) bool { return false })

	// Ascend, and test that the filter has been cleared again when ascending.
	w.Ascend()
	f.App.Redraw()
	f.TTY.TestBuffer(t, d2Buf)

	// Now move into d3, an empty directory.
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

func TestNewNavigation_FocusedWidgetNotCodeArea(t *testing.T) {
	testFocusedWidgetNotCodeArea(t, func(app cli.App) error {
		_, err := NewNavigation(app, NavigationSpec{})
		return err
	})
}

func setupNav(c testutil.Cleanuper) *Fixture {
	lscolors.SetTestLsColors(c)
	// Use a small TTY size to make the test buffer easier to build.
	return Setup(WithTTY(func(tty TTYCtrl) { tty.SetSize(6, 40) }))
}

func startNavigation(app cli.App, spec NavigationSpec) Navigation {
	w, _ := NewNavigation(app, spec)
	startMode(app, w, nil)
	return w
}

func getTestCursor() *testCursor {
	return &testCursor{root: testDir, pwd: []string{"d"}}
}
