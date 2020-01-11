package edit

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/pkg/store"
)

func TestEditor_AddsHistoryAfterAccepting(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	feedInput(f.TTYCtrl, "echo x\n")
	f.Wait()

	testCommands(t, f.Store, "echo x")
}

func TestEditor_DoesNotAddEmptyCommandToHistory(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	feedInput(f.TTYCtrl, "\n")
	f.Wait()

	testCommands(t, f.Store /* no commands */)
}

func TestEditor_DoesCommandWithLeadingSpaceToHistory(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	feedInput(f.TTYCtrl, " echo\n")
	f.Wait()

	testCommands(t, f.Store /* no commands */)
}

func testCommands(t *testing.T, store store.Store, wantCmds ...string) {
	t.Helper()
	cmds, err := store.Cmds(0, 1024)
	if err != nil {
		panic(err)
	}
	if !reflect.DeepEqual(cmds, wantCmds) {
		t.Errorf("got cmds %v, want %v", cmds, wantCmds)
	}
}
