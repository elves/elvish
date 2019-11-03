package navigation

import (
	"errors"
	"os"
	"testing"

	"github.com/elves/elvish/cli"
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
}

var testDir = util.Dir{
	"a": "",
	"d": util.Dir{
		"d1": "content\td1",
		"d2": util.Dir{
			"d21": "content d21",
			"d22": "content d22",
		},
		"d3": util.Dir{},
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
	testNavigation(t, nil)
}

func TestErrorInAscend(t *testing.T) {
	app, ttyCtrl, cleanup := setupApp()
	defer cleanup()

	c := getTestCursor()
	c.ascendErr = errors.New("cannot ascend")
	Start(app, Config{Cursor: c})

	ttyCtrl.Inject(term.K(ui.Left))
	ttyCtrl.TestNotesBuffer(t, makeNotesBuf(styled.Plain("cannot ascend")))
}

func TestErrorInDescend(t *testing.T) {
	app, ttyCtrl, cleanup := setupApp()
	defer cleanup()

	c := getTestCursor()
	c.descendErr = errors.New("cannot descend")
	Start(app, Config{Cursor: c})

	ttyCtrl.Inject(term.K(ui.Down))
	ttyCtrl.Inject(term.K(ui.Right))
	ttyCtrl.TestNotesBuffer(t, makeNotesBuf(styled.Plain("cannot descend")))
}

func TestErrorInCurrent(t *testing.T) {
	app, ttyCtrl, cleanup := setupApp()
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
	app, ttyCtrl, cleanup := setupApp()
	defer cleanup()

	c := getTestCursor()
	c.parentErr = errors.New("ERR")
	Start(app, Config{Cursor: c})

	buf := makeBuf(styled.MarkLines(
		"ERR   d1            content    d1", styles,
		"xxx  --------------",
		"      d2           ", styles,
		"     ++++++++++++++",
		"      d3           ", styles,
		"     ++++++++++++++",
	))
	ttyCtrl.TestBuffer(t, buf)

}

func TestErrorInDeepMode(t *testing.T) {
}

func testNavigation(t *testing.T, c Cursor) {
	app, ttyCtrl, cleanup := setupApp()
	defer cleanup()

	Start(app, Config{Cursor: c})

	// Test initial UI and file preview.
	// NOTE: Buffers are named after the file that is now being selected.
	d1Buf := makeBuf(styled.MarkLines(
		" a    d1            content    d1", styles,
		"     --------------",
		" d    d2           ", styles,
		"#### ++++++++++++++",
		" f    d3           ", styles,
		"     ++++++++++++++",
	))
	ttyCtrl.TestBuffer(t, d1Buf)

	// Test handling of selection change and directory preview.
	ttyCtrl.Inject(term.K(ui.Down))
	d2Buf := makeBuf(styled.MarkLines(
		" a    d1             d21                ", styles,
		"                    --------------------",
		" d    d2             d22                ", styles,
		"#### ##############",
		" f    d3           ", styles,
		"     ++++++++++++++",
	))
	ttyCtrl.TestBuffer(t, d2Buf)

	// Test handling of Right.
	ttyCtrl.Inject(term.K(ui.Right))
	d21Buf := makeBuf(styled.MarkLines(
		" d1   d21           content d21", styles,
		"     --------------",
		" d2   d22          ", styles,
		"####",
		" d3 ", styles,
		"++++",
	))
	ttyCtrl.TestBuffer(t, d21Buf)

	// Test handling of Left, and that the current column selects the directory
	// we just ascended from, thus reverting to wantBuf1.
	ttyCtrl.Inject(term.K(ui.Left))
	ttyCtrl.TestBuffer(t, d2Buf)

	// Test handling of Right on a regular file, i.e. do nothing. First move the
	// cursor to d1, which is a regular file.
	ttyCtrl.Inject(term.K(ui.Up))
	ttyCtrl.TestBuffer(t, d1Buf)
	// Now inject Right and verify that the buffer has not changed.
	ttyCtrl.Inject(term.K(ui.Right))
	ttyCtrl.TestBuffer(t, d1Buf)

	// Test handling of empty directories. First move into d3, an empty directory.
	ttyCtrl.Inject(term.K(ui.Down), term.K(ui.Down), term.K(ui.Right))
	d3NoneBuf := makeBuf(styled.MarkLines(
		" d1                 ",
		" d2 ", styles,
		"++++",
		" d3 ", styles,
		"####",
	))
	ttyCtrl.TestBuffer(t, d3NoneBuf)
	// Test that Up does nothing.
	ttyCtrl.Inject(term.K(ui.Up))
	ttyCtrl.TestBuffer(t, d3NoneBuf)
	// Test that Down does nothing.
	ttyCtrl.Inject(term.K(ui.Down))
	ttyCtrl.TestBuffer(t, d3NoneBuf)
	// Test that Right does nothing.
	ttyCtrl.Inject(term.K(ui.Right))
	ttyCtrl.TestBuffer(t, d3NoneBuf)
}

func makeBuf(navRegion styled.Text) *ui.Buffer {
	return ui.NewBufferBuilder(40).
		Newline().SetDotToCursor().
		WriteStyled(navRegion).Buffer()
}

func makeNotesBuf(content styled.Text) *ui.Buffer {
	return ui.NewBufferBuilder(40).WriteStyled(content).Buffer()
}

func setupApp() (cli.App, cli.TTYCtrl, func()) {
	tty, ttyCtrl := cli.NewFakeTTY()
	ttyCtrl.SetSize(24, 40)
	app := cli.NewApp(cli.AppSpec{TTY: tty})
	codeCh, _ := cli.ReadCodeAsync(app)
	return app, ttyCtrl, func() {
		app.CommitEOF()
		<-codeCh
	}
}

func getTestCursor() *testCursor {
	return &testCursor{root: testDir, pwd: []string{"d"}}
}
