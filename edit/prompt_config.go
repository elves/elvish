package edit

import (
	"math"

	"github.com/elves/elvish/edit/prompt"
	"github.com/elves/elvish/eval"
)

func init() {
	atEditorInit(func(ed *editor, ns eval.Ns) {
		ed.Prompt = prompt.PromptInit()
		ns["prompt"] = eval.NewVariableFromPtr(&ed.Prompt)
		ed.Rprompt = prompt.RpromptInit()
		ns["rprompt"] = eval.NewVariableFromPtr(&ed.Rprompt)
		ed.RpromptPersistent = false
		ns["rprompt-persistent"] = eval.NewVariableFromPtr(&ed.RpromptPersistent)
		ed.PromptsMaxWait = math.Inf(1)
		ns["-prompts-max-wait"] = eval.NewVariableFromPtr(&ed.PromptsMaxWait)
	})
}
