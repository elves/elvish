package edit

import (
	"os"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
)

func init() {
	atEditorInit(func(ed *editor, ns eval.Ns) {
		ed.promptsEagerness = 5
		ns["-prompts-eagerness"] = vars.NewFromPtr(&ed.promptsEagerness)
	})
}

func shouldUpdatePrompt(ed *editor) bool {
	if ed.promptsEagerness >= 10 {
		return true
	}
	if ed.promptsEagerness >= 5 {
		pwd, err := os.Getwd()
		if err != nil {
			pwd = "error"
		}
		oldPwd := ed.pwdOnLastPromptUpdate
		ed.pwdOnLastPromptUpdate = pwd
		return pwd != oldPwd
	}
	return false
}
