package eval

import (
	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/parse/cmpd"
)

// Utilities for working with nodes.

func stringLiteralOrError(cp *compiler, n *parse.Compound, what string) string {
	s, err := cmpd.StringLiteralOrError(n, what)
	if err != nil {
		cp.errorpf(n, "%v", err)
	}
	return s
}

type errorpfer interface {
	errorpf(r diag.Ranger, fmt string, args ...interface{})
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
		aw.cp.errorpf(aw.form, "need more arguments")
	}
	return aw.form.Args[aw.idx]
}

func (aw *argsWalker) next() *parse.Compound {
	n := aw.peek()
	aw.idx++
	return n
}

func (aw *argsWalker) peekIs(text string) bool {
	return aw.more() && parse.SourceText(aw.form.Args[aw.idx]) == text
}

// nextIs returns whether the next argument's source matches the given text. It
// also consumes the argument if it is.
func (aw *argsWalker) nextIs(text string) bool {
	if aw.peekIs(text) {
		aw.idx++
		return true
	}
	return false
}

// nextMustLambda fetches the next argument, raising an error if it is not a
// lambda.
func (aw *argsWalker) nextMustLambda(what string) *parse.Primary {
	n := aw.next()
	pn, ok := cmpd.Lambda(n)
	if !ok {
		aw.cp.errorpf(n, "%s must be lambda, found %s", what, cmpd.Shape(n))
	}
	return pn
}

// nextMustThunk fetches the next argument, raising an error if it is not a
// thunk.
func (aw *argsWalker) nextMustThunk(what string) *parse.Primary {
	n := aw.nextMustLambda(what)
	if len(n.Elements) > 0 {
		aw.cp.errorpf(n, "%s must not have arguments", what)
	}
	if len(n.MapPairs) > 0 {
		aw.cp.errorpf(n, "%s must not have options", what)
	}
	return n
}

func (aw *argsWalker) nextMustThunkIfAfter(leader string) *parse.Primary {
	if aw.nextIs(leader) {
		return aw.nextMustLambda(leader + " body")
	}
	return nil
}

func (aw *argsWalker) mustEnd() {
	if aw.more() {
		aw.cp.errorpf(diag.Ranging{From: aw.form.Args[aw.idx].Range().From, To: aw.form.Range().To}, "too many arguments")
	}
}
