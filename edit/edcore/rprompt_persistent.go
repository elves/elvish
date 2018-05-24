package edcore

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
)

func init() {
	atEditorInit(func(ed *editor, ns eval.Ns) {
		ed.RpromptPersistent = false
		ns["rprompt-persistent"] = vars.FromPtr(&ed.RpromptPersistent)
	})
}
