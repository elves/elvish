package newedit

import (
	"testing"

	"github.com/elves/elvish/eval"
)

func TestNs(t *testing.T) {
	ev := eval.NewEvaler()
	ed := NewEditor(devNull, devNull, ev, testStore)
	ev.Global.AddNs("edit", ed.Ns())

	ev.EvalSourceInTTY(eval.NewScriptSource("[t]", "[t]", "edit:max-height = 20"))
	if ed.app.Config.Raw.MaxHeight != 20 {
		t.Errorf("Failed to set MaxHeight to 20 via binding")
	}
}

func TestAddCmdAfterReadline(t *testing.T) {
	// TODO
}

func TestInsertMode(t *testing.T) {
	// TODO
}

func TestDefaultBinding(t *testing.T) {
	// TODO
}
