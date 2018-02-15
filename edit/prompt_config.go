package edit

import (
	"math"

	"github.com/elves/elvish/edit/prompt"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
)

func init() {
	atEditorInit(func(ed *editor, ns eval.Ns) {
		ed.Prompt = prompt.PromptInit()
		ns["prompt"] = vars.NewFromPtr(&ed.Prompt)
		ed.Rprompt = prompt.RpromptInit()
		ns["rprompt"] = vars.NewFromPtr(&ed.Rprompt)
		ed.RpromptPersistent = false
		ns["rprompt-persistent"] = vars.NewFromPtr(&ed.RpromptPersistent)
		ed.PromptsMaxWait = math.Inf(1)
		ns["-prompts-max-wait"] = vars.NewFromPtr(&ed.PromptsMaxWait)
	})
}
