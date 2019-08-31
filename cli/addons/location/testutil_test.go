package location

import (
	"testing"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/edit/ui"
)

func setup() (*cli.App, cli.TTYCtrl, func()) {
	tty, ttyCtrl := cli.NewFakeTTY()
	// Use a smaller TTY size to make diffs easier to see.
	ttyCtrl.SetSize(20, 50)
	app := cli.NewApp(tty)
	codeCh, _ := app.ReadCodeAsync()
	return app, ttyCtrl, func() {
		app.CommitEOF()
		<-codeCh
	}
}

func testBuf(t *testing.T, ttyCtrl cli.TTYCtrl, wantBuf *ui.Buffer) {
	if !ttyCtrl.VerifyBuffer(wantBuf) {
		t.Errorf("Wanted buffer not shown")
		t.Logf("Want: %s", wantBuf.TTYString())
		t.Logf("Last buffer: %s", ttyCtrl.LastBuffer().TTYString())
	}
}

func testNotesBuf(t *testing.T, ttyCtrl cli.TTYCtrl, wantBuf *ui.Buffer) {
	if !ttyCtrl.VerifyNotesBuffer(wantBuf) {
		t.Errorf("Wanted notes buffer not shown")
		t.Logf("Want: %s", wantBuf.TTYString())
		t.Logf("Last buffer: %s", ttyCtrl.LastNotesBuffer().TTYString())
	}
}
