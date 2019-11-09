package cliedit

import (
	"testing"

	"github.com/elves/elvish/cli/el/layout"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/store/storedefs"
	"github.com/elves/elvish/styled"
)

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

// Smoke tests for individual addons.

func TestHistlistAddon(t *testing.T) {
	f := setupWithOpt(setupOpt{StoreOp: func(s storedefs.Store) {
		s.AddCmd("echo 1")
		s.AddCmd("echo 2")
	}})
	f.TTYCtrl.SetSize(24, 30) // Set width to 30
	defer f.Cleanup()

	f.TTYCtrl.Inject(term.K('R', ui.Ctrl))
	wantBuf := bbAddon("HISTLIST").
		WriteStyled(styled.MarkLines(
			"   0 echo 1",
			"   1 echo 2                   ", styles,
			"##############################",
		)).Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)
}

func TestLastCmdAddon(t *testing.T) {
	f := setupWithOpt(setupOpt{StoreOp: func(s storedefs.Store) {
		s.AddCmd("echo hello world")
	}})
	f.TTYCtrl.SetSize(24, 30) // Set width to 30
	defer f.Cleanup()

	f.TTYCtrl.Inject(term.K(',', ui.Alt))
	wantBuf := bbAddon("LASTCMD").
		WriteStyled(styled.MarkLines(
			"    echo hello world          ", styles,
			"##############################",
			"  0 echo",
			"  1 hello",
			"  2 world",
		)).Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)
}

func TestLocationAddon(t *testing.T) {
	f := setupWithOpt(setupOpt{StoreOp: func(s storedefs.Store) {
		s.AddDir("/usr/bin", 1)
		s.AddDir("/home/elf", 1)
	}})
	f.TTYCtrl.SetSize(24, 30) // Set width to 30
	defer f.Cleanup()

	f.TTYCtrl.Inject(term.K('L', ui.Ctrl))
	wantBuf := bbAddon("LOCATION").
		WriteStyled(styled.MarkLines(
			" 10 /home/elf                 ", styles,
			"##############################",
			" 10 /usr/bin",
		)).Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)
}

func bbAddon(name string) *ui.BufferBuilder {
	return ui.NewBufferBuilder(30).
		WritePlain("~> ").Newline().
		WriteStyled(layout.ModeLine(name, true)).SetDotToCursor().Newline()
}
