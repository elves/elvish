package navigation

import (
	"errors"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/el/layout"
	"github.com/elves/elvish/cli/el/listbox"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/util"
)

var styles = map[rune]string{
	'-': "inverse",
	'+': "blue",
	'#': "inverse blue",
	'x': "red",

	't': "magenta",
	'T': "magenta inverse",
}

var testDir = util.Dir{
	"a": "",
	"d": util.Dir{
		"d1": "content\td1\nline 2",
		"d2": util.Dir{
			"d21":     "content d21",
			"d22":     "content d22",
			"d23.exe": "",
		},
		"d3":  util.Dir{},
		".dh": "hidden",
	},
	"f": "",
}

func TestNavigation_FakeFS(t *testing.T) {
	cursor := getTestCursor()
	testNavigation(t, cursor)
}

func TestNavigation_RealFS(t *testing.T) {
	cleanupFs := util.InTestDirWithSetup(testDir)
	err := os.Chdir("d")
	if err != nil {
		panic(err)
	}
	defer cleanupFs()
	fmt.Println("ext", path.Ext("d2/d23.exe"))
	testNavigation(t, nil)
}

func TestErrorInAscend(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	c := getTestCursor()
	c.ascendErr = errors.New("cannot ascend")
	Start(app, Config{Cursor: c})

	ttyCtrl.Inject(term.K(ui.Left))
	ttyCtrl.TestNotesBuffer(t, makeNotesBuf(styled.Plain("cannot ascend")))
}

func TestErrorInDescend(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	c := getTestCursor()
	c.descendErr = errors.New("cannot descend")
	Start(app, Config{Cursor: c})

	ttyCtrl.Inject(term.K(ui.Down))
	ttyCtrl.Inject(term.K(ui.Right))
	ttyCtrl.TestNotesBuffer(t, makeNotesBuf(styled.Plain("cannot descend")))
}

func TestErrorInCurrent(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	c := getTestCursor()
	c.currentErr = errors.New("ERR")
	Start(app, Config{Cursor: c})

	buf := makeBuf(styled.MarkLines(
		" a   ERR            ", styles,
		"     xxx",
		" d  ", styles,
		"++++",
		" f  ",
	))
	ttyCtrl.TestBuffer(t, buf)

	// Test that Right does nothing.
	ttyCtrl.Inject(term.K(ui.Right))
	ttyCtrl.TestBuffer(t, buf)
}

func TestErrorInParent(t *testing.T) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	c := getTestCursor()
	c.parentErr = errors.New("ERR")
	Start(app, Config{Cursor: c})

	buf := makeBuf(styled.MarkLines(
		"ERR   d1            content    d1", styles,
		"xxx  --------------",
		"      d2            line 2", styles,
		"     ++++++++++++++",
		"      d3           ", styles,
		"     ++++++++++++++",
	))
	ttyCtrl.TestBuffer(t, buf)
}

func TestGetSelectedName(t *testing.T) {
	app, _, cleanup := setup()
	defer cleanup()

	wantName := ""
	if name := SelectedName(app); name != wantName {
		t.Errorf("Got name %q, want %q", name, wantName)
	}

	Start(app, Config{Cursor: getTestCursor()})

	wantName = "d1"
	if name := SelectedName(app); name != wantName {
		t.Errorf("Got name %q, want %q", name, wantName)
	}
}

func testNavigation(t *testing.T, c Cursor) {
	app, ttyCtrl, cleanup := setup()
	defer cleanup()

	Start(app, Config{Cursor: c})

	// Test initial UI and file preview.
	// NOTE: Buffers are named after the file that is now being selected.
	d1Buf := makeBuf(styled.MarkLines(
		" a    d1            content    d1", styles,
		"     --------------",
		" d    d2            line 2", styles,
		"#### ++++++++++++++",
		" f    d3           ", styles,
		"     ++++++++++++++",
	))
	ttyCtrl.TestBuffer(t, d1Buf)

	// Test scrolling of preview.
	ScrollPreview(app, 1)
	d1Buf2 := makeBuf(styled.MarkLines(
		" a    d1            line 2             │", styles,
		"     --------------                    t",
		" d    d2                               │", styles,
		"#### ++++++++++++++                    t",
		" f    d3                                ", styles,
		"     ++++++++++++++                    T",
		"                                        ", styles,
		"                                       T",
	))
	ttyCtrl.TestBuffer(t, d1Buf2)

	// Test handling of selection change and directory preview. Also test
	// LS_COLORS.
	Select(app, listbox.Next)
	d2Buf := makeBuf(styled.MarkLines(
		" a    d1             d21                ", styles,
		"                    --------------------",
		" d    d2             d22                ", styles,
		"#### ##############",
		" f    d3             d23.exe            ", styles,
		"     ++++++++++++++ xxxxxxxxxxxxxxxxxxxx",
	))
	ttyCtrl.TestBuffer(t, d2Buf)

	// Test handling of Descend.
	Descend(app)
	d21Buf := makeBuf(styled.MarkLines(
		" d1   d21           content d21", styles,
		"     --------------",
		" d2   d22          ", styles,
		"####",
		" d3   d23.exe      ", styles,
		"++++ xxxxxxxxxxxxxx",
	))
	ttyCtrl.TestBuffer(t, d21Buf)

	// Test handling of Ascend, and that the current column selects the
	// directory we just ascended from, thus reverting to wantBuf1.
	Ascend(app)
	ttyCtrl.TestBuffer(t, d2Buf)

	// Test handling of Descend on a regular file, i.e. do nothing. First move
	// the cursor to d1, which is a regular file.
	Select(app, listbox.Prev)
	ttyCtrl.TestBuffer(t, d1Buf)
	// Now descend, and verify that the buffer has not changed.
	Descend(app)
	ttyCtrl.TestBuffer(t, d1Buf)

	// Test showing hidden.
	MutateShowHidden(app, func(bool) bool { return true })
	ttyCtrl.TestBuffer(t, makeShowHiddenBuf(styled.MarkLines(
		" a    .dh           content    d1",
		" d    d1            line 2", styles,
		"#### --------------",
		" f    d2           ", styles,
		"     ++++++++++++++",
		"      d3           ", styles,
		"     ++++++++++++++",
	)))
	MutateShowHidden(app, func(bool) bool { return false })

	// Test filtering; current column shows d1, d2, d3 before filtering.
	MutateFiltering(app, func(bool) bool { return true })
	ttyCtrl.Inject(term.K('3'))
	ttyCtrl.TestBuffer(t, makeFilteringBuf("3",
		styled.MarkLines(
			" a    d3            ", styles,
			"     ##############",
			" d  ", styles,
			"####",
			" f  ", styles,
			"    ",
		)))
	MutateFiltering(app, func(bool) bool { return false })

	// Now move into d3, an empty directory. Test that the filter has been
	// cleared.
	Select(app, listbox.Next)
	Select(app, listbox.Next)
	Descend(app)
	d3NoneBuf := makeBuf(styled.MarkLines(
		" d1                 ",
		" d2 ", styles,
		"++++",
		" d3 ", styles,
		"####",
	))
	ttyCtrl.TestBuffer(t, d3NoneBuf)
	// Test that selecting the previous does nothing in an empty directory.
	Select(app, listbox.Prev)
	ttyCtrl.TestBuffer(t, d3NoneBuf)
	// Test that selecting the next does nothing in an empty directory.
	Select(app, listbox.Next)
	ttyCtrl.TestBuffer(t, d3NoneBuf)
	// Test that Descend does nothing in an empty directory.
	Descend(app)
	ttyCtrl.TestBuffer(t, d3NoneBuf)
}

func makeBuf(navRegion styled.Text) *ui.Buffer {
	return ui.NewBufferBuilder(40).SetDotHere().
		Newline().WriteStyled(layout.ModeLine(" NAVIGATING ", true)).
		Newline().WriteStyled(navRegion).Buffer()
}

func makeShowHiddenBuf(navRegion styled.Text) *ui.Buffer {
	return ui.NewBufferBuilder(40).SetDotHere().
		Newline().WriteStyled(layout.ModeLine(" NAVIGATING (show hidden) ", true)).
		Newline().WriteStyled(navRegion).Buffer()
}

func makeFilteringBuf(filter string, navRegion styled.Text) *ui.Buffer {
	return ui.NewBufferBuilder(40).
		Newline().WriteStyled(layout.ModeLine(" NAVIGATING ", true)).
		WritePlain(filter).SetDotHere().
		Newline().WriteStyled(navRegion).Buffer()
}

func makeNotesBuf(content styled.Text) *ui.Buffer {
	return ui.NewBufferBuilder(40).WriteStyled(content).Buffer()
}

func setup() (cli.App, cli.TTYCtrl, func()) {
	oldLsColors := os.Getenv("LS_COLORS")
	// Directories are blue, *.exe files are red.
	os.Setenv("LS_COLORS", "di=34:*.exe=31")
	tty, ttyCtrl := cli.NewFakeTTY()
	ttyCtrl.SetSize(6, 40)
	app := cli.NewApp(cli.AppSpec{TTY: tty})
	codeCh, _ := cli.ReadCodeAsync(app)
	return app, ttyCtrl, func() {
		app.CommitEOF()
		<-codeCh
		os.Setenv("LS_COLORS", oldLsColors)
	}
}

func getTestCursor() *testCursor {
	return &testCursor{root: testDir, pwd: []string{"d"}}
}
