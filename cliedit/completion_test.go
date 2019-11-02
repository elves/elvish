package cliedit

import (
	"testing"

	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/util"
)

func TestCompletion(t *testing.T) {
	cleanupFs := util.SetupTestDir(util.Dir{"a": "", "b": ""}, "")
	defer cleanupFs()
	app, ttyCtrl, ns, ev := prepare()
	initCompletion(app, ev, ns)

	stop := run(app)
	defer stop()

	app.CodeArea().MutateState(func(s *codearea.State) {
		s.CodeBuffer.InsertAtDot("echo ")
	})
	evalf(ev, "edit:completion:start")
	wantBuf := ui.NewBufferBuilder(40).
		WriteStyled(styled.MarkLines(
			"echo a ", styles,
			"     --",
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
