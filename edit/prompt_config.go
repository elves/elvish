package edit

import (
	"math"

	"github.com/elves/elvish/edit/prompt"
	"github.com/elves/elvish/eval"
)

func init() {
	atEditorInit(func(ed *Editor) {
		ed.Prompt = prompt.PromptInit()
		ed.variables["prompt"] = eval.NewVariableFromPtr(&ed.Prompt)
		ed.Rprompt = prompt.RpromptInit()
		ed.variables["rprompt"] = eval.NewVariableFromPtr(&ed.Rprompt)
		ed.RpromptPersistent = false
		ed.variables["rprompt-persistent"] = eval.NewVariableFromPtr(&ed.RpromptPersistent)
		ed.PromptsMaxWait = math.Inf(1)
		ed.variables["-prompts-max-wait"] = eval.NewVariableFromPtr(&ed.PromptsMaxWait)
	})
}
