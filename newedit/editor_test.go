package newedit

import (
	"os"
	"testing"

	"github.com/elves/elvish/eval"
)

func TestNs(t *testing.T) {
	ev := eval.NewEvaler()
	ed := NewEditor(os.Stdin, os.Stdout, ev)
	ev.Global.AddNs("edit", ed.Ns())

	ev.EvalSource(eval.NewScriptSource("[t]", "[t]", "edit:max-height = 20"))
	if ed.core.Config.Raw.MaxHeight != 20 {
		t.Errorf("Failed to set MaxHeight to 20 via binding")
	}
}
