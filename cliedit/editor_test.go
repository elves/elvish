package cliedit

import (
	"reflect"
	"testing"
)

func TestAddsHistoryAfterAccepting(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	feedInput(f.TTYCtrl, "echo x\n")
	f.Wait()

	cmds, err := f.Store.Cmds(0, 100)
	if err != nil {
		panic(err)
	}
	wantCmds := []string{"echo x"}
	if !reflect.DeepEqual(cmds, wantCmds) {
		t.Errorf("got cmds %v, want %v", cmds, wantCmds)
	}
}
