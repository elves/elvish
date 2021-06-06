package edit

import (
	"reflect"
	"testing"

	"src.elv.sh/pkg/store"
)

func TestEditor_AddsHistoryAfterAccepting(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	feedInput(f.TTYCtrl, "echo x\n")
	f.Wait()

	testCommands(t, f.Store, store.Cmd{Text: "echo x", Seq: 1})
}

func TestEditor_DoesNotAddEmptyCommandToHistory(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	feedInput(f.TTYCtrl, "\n")
	f.Wait()

	testCommands(t, f.Store /* no commands */)
}

func testCommands(t *testing.T, store store.Store, wantCmds ...store.Cmd) {
	t.Helper()
	cmds, err := store.CmdsWithSeq(0, 1024)
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(cmds, wantCmds) {
		t.Errorf("got cmds %v, want %v", cmds, wantCmds)
	}
}
