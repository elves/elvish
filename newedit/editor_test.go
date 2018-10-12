package newedit

import (
	"os"
	"testing"

	"github.com/elves/elvish/eval"
)

var devNull *os.File

func init() {
	f, err := os.Open(os.DevNull)
	if err != nil {
		panic(err)
	}
	devNull = f
}

func TestNs(t *testing.T) {
	ev := eval.NewEvaler()
	ed := NewEditor(devNull, devNull, ev)
	ev.Global.AddNs("edit", ed.Ns())

	ev.EvalSourceInTTY(eval.NewScriptSource("[t]", "[t]", "edit:max-height = 20"))
	if ed.core.Config.Raw.MaxHeight != 20 {
		t.Errorf("Failed to set MaxHeight to 20 via binding")
	}
}

func TestInsertMode(t *testing.T) {
	// TODO
}

func TestDefaultBinding(t *testing.T) {
	// TODO
}
