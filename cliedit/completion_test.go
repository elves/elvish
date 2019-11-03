package cliedit

import (
	"testing"

	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/util"
)

func TestCompletion(t *testing.T) {
	_, cleanupDir := eval.InTempHome()
	util.ApplyDir(util.Dir{"a": "", "b": ""})
	defer cleanupDir()
	ed, ttyCtrl, ev, _, cleanup := setupStarted()
	defer cleanup()

	ed.app.CodeArea().MutateState(func(s *codearea.State) {
		s.CodeBuffer.InsertAtDot("echo ")
	})
	evalf(ev, "edit:completion:start")
	wantBuf := ui.NewBufferBuilder(40).
		WriteStyled(styled.MarkLines(
			"~> echo a ", styles,
			"        --",
			"COMPLETING argument ", styles,
			"mmmmmmmmmmmmmmmmmmm ")).
		SetDotToCursor().
		Newline().
		WriteStyled(styled.MarkLines(
			"a  b", styles,
			"#   ",
		)).
		Buffer()
	ttyCtrl.TestBuffer(t, wantBuf)
}
