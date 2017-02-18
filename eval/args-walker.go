package eval

import (
	"strings"

	"github.com/elves/elvish/parse"
)

type argsWalker struct {
	cp   *compiler
	form *parse.Form
	idx  int
}

func (cp *compiler) walkArgs(f *parse.Form) *argsWalker {
	return &argsWalker{cp, f, 0}
}

func (aw *argsWalker) more() bool {
	return aw.idx < len(aw.form.Args)
}

func (aw *argsWalker) next() *parse.Compound {
	if !aw.more() {
		aw.cp.errorpf(aw.form.End(), aw.form.End(), "need more arguments")
	}
	aw.idx++
	return aw.form.Args[aw.idx-1]
}

func (aw *argsWalker) nextMustBeOneOf(valids ...string) {
	n := aw.next()
	text := n.SourceText()
	for _, valid := range valids {
		if text == valid {
			return
		}
	}

	if len(valids) == 1 {
		aw.cp.errorpf(n.Begin(), n.End(), "should be %s", valids[0])
	}
	aw.cp.errorpf(n.Begin(), n.End(), "should be one of %s", strings.Join(valids, ","))
}

func (aw *argsWalker) nextLedBy(leader string) *parse.Compound {
	if !aw.more() || aw.form.Args[aw.idx].SourceText() != leader {
		return nil
	}
	aw.next()
	return aw.next()
}

func (aw *argsWalker) mustEnd() {
	if aw.more() {
		aw.cp.errorpf(aw.form.Args[aw.idx].Begin(), aw.form.End(), "too many arguments")
	}
}
