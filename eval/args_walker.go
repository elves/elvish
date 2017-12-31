package eval

import "github.com/elves/elvish/parse"

type errorpfer interface {
	errorpf(begin, end int, fmt string, args ...interface{})
}

// argsWalker is used by builtin special forms to implement argument parsing.
type argsWalker struct {
	cp   errorpfer
	form *parse.Form
	idx  int
}

func (cp *compiler) walkArgs(f *parse.Form) *argsWalker {
	return &argsWalker{cp, f, 0}
}

func (aw *argsWalker) more() bool {
	return aw.idx < len(aw.form.Args)
}

func (aw *argsWalker) peek() *parse.Compound {
	if !aw.more() {
		aw.cp.errorpf(aw.form.End(), aw.form.End(), "need more arguments")
	}
	return aw.form.Args[aw.idx]
}

func (aw *argsWalker) next() *parse.Compound {
	n := aw.peek()
	aw.idx++
	return n
}

// nextIs returns whether the next argument's source matches the given text. It
// also consumes the argument if it is.
func (aw *argsWalker) nextIs(text string) bool {
	if aw.more() && aw.form.Args[aw.idx].SourceText() == text {
		aw.idx++
		return true
	}
	return false
}

// nextMustLambda fetches the next argument, raising an error if it is not a
// lambda.
func (aw *argsWalker) nextMustLambda() *parse.Primary {
	n := aw.next()
	if len(n.Indexings) != 1 {
		aw.cp.errorpf(n.Begin(), n.End(), "must be lambda")
	}
	if len(n.Indexings[0].Indicies) != 0 {
		aw.cp.errorpf(n.Begin(), n.End(), "must be lambda")
	}
	pn := n.Indexings[0].Head
	if pn.Type != parse.Lambda {
		aw.cp.errorpf(n.Begin(), n.End(), "must be lambda")
	}
	return pn
}

func (aw *argsWalker) nextMustLambdaIfAfter(leader string) *parse.Primary {
	if aw.nextIs(leader) {
		return aw.nextMustLambda()
	}
	return nil
}

func (aw *argsWalker) mustEnd() {
	if aw.more() {
		aw.cp.errorpf(aw.form.Args[aw.idx].Begin(), aw.form.End(), "too many arguments")
	}
}
