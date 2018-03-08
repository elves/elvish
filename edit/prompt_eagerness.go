package edit

import (
	"os"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
)

func init() {
	atEditorInit(func(ed *editor, ns eval.Ns) {
		ed.promptEagerness = 5
		ns["-prompt-eagerness"] = vars.NewFromPtr(&ed.promptEagerness)
	})
}

func shouldUpdatePrompt(ed *editor) bool {
	if ed.promptEagerness >= 10 {
		return true
	}
	if ed.promptEagerness >= 5 {
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
