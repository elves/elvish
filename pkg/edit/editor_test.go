package edit

import (
	"reflect"
	"testing"

	"src.elv.sh/pkg/store/storedefs"
	"src.elv.sh/pkg/ui"
)

func TestEditor_AddsHistoryAfterAccepting(t *testing.T) {
	f := setup(t)

	feedInput(f.TTYCtrl, "echo x\n")
	f.Wait()

	testCommands(t, f.Store, storedefs.Cmd{Text: "echo x", Seq: 1})
}

func TestEditor_DoesNotAddEmptyCommandToHistory(t *testing.T) {
	f := setup(t)

	feedInput(f.TTYCtrl, "\n")
	f.Wait()

	testCommands(t, f.Store /* no commands */)
}

func TestEditor_Notify(t *testing.T) {
	f := setup(t)
	f.Editor.Notify(ui.T("note"))
	f.TestTTYNotes(t, "note")
}

func testCommands(t *testing.T, store storedefs.Store, wantCmds ...storedefs.Cmd) {
	t.Helper()
	cmds, err := store.CmdsWithSeq(0, 1024)
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(cmds, wantCmds) {
		t.Errorf("got cmds %v, want %v", cmds, wantCmds)
	}
}
