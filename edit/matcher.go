package edit

import (
	"strings"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/util"
)

var _ = registerVariable("-use-subseq-matcher", func() eval.Variable {
	return eval.NewPtrVariableWithValidator(eval.Bool(false), eval.ShouldBeBool)
})

func (ed *Editor) useSubseqMatcher() bool {
	return bool(ed.variables["-use-subseq-matcher"].Get().(eval.Bool).Bool())
}

func (ed *Editor) matcher() func(string, string) bool {
	if ed.useSubseqMatcher() {
		return util.HasSubseq
	} else {
		return strings.HasPrefix
	}
}
