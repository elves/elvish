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

type argsGetter struct {
	cp *compiler
	fn *parse.Form
	ok bool
	n  int
}

func getArgs(cp *compiler, fn *parse.Form) *argsGetter {
	return &argsGetter{cp, fn, true, 0}
}

func (ag *argsGetter) errorpf(r diag.Ranger, format string, args ...any) {
	if ag.ok {
		ag.cp.errorpf(r, format, args...)
		ag.ok = false
	}
}

func (ag *argsGetter) get(i int, what string) *argAsserter {
	if ag.n < i+1 {
		ag.n = i + 1
	}
	if i >= len(ag.fn.Args) {
		ag.errorpf(diag.PointRanging(ag.fn.To), "need %s", what)
		return &argAsserter{ag, what, nil}
	}
	return &argAsserter{ag, what, ag.fn.Args[i]}
}

func (ag *argsGetter) has(i int) bool { return i < len(ag.fn.Args) }

func (ag *argsGetter) hasKeyword(i int, kw string) bool {
	if i < len(ag.fn.Args) {
		s, ok := cmpd.StringLiteral(ag.fn.Args[i])
		return ok && s == kw
	}
	return false
}

func (ag *argsGetter) optionalKeywordBody(i int, kw string) *parse.Primary {
	if ag.has(i+1) && ag.hasKeyword(i, kw) {
		return ag.get(i+1, kw+" body").thunk()
	}
	return nil
}

func (ag *argsGetter) finish() bool {
	if ag.n < len(ag.fn.Args) {
		ag.errorpf(
			diag.Ranging{From: ag.fn.Args[ag.n].Range().From, To: ag.fn.To},
			"superfluous arguments")
	}
	return ag.ok
}

type argAsserter struct {
	ag   *argsGetter
	what string
	node *parse.Compound
}

func (aa *argAsserter) any() *parse.Compound {
	return aa.node
}

func (aa *argAsserter) stringLiteral() string {
	if aa.node == nil {
		return ""
	}
	s, err := cmpd.StringLiteralOrError(aa.node, aa.what)
	if err != nil {
		aa.ag.errorpf(aa.node, "%v", err)
		return ""
	}
	return s
}

func (aa *argAsserter) lambda() *parse.Primary {
	if aa.node == nil {
		return nil
	}
	lambda, ok := cmpd.Lambda(aa.node)
	if !ok {
		aa.ag.errorpf(aa.node,
			"%s must be lambda, found %s", aa.what, cmpd.Shape(aa.node))
		return nil
	}
	return lambda
}

func (aa *argAsserter) thunk() *parse.Primary {
	lambda := aa.lambda()
	if lambda == nil {
		return nil
	}
	if len(lambda.Elements) > 0 {
		aa.ag.errorpf(lambda, "%s must not have arguments", aa.what)
		return nil
	}
	if len(lambda.MapPairs) > 0 {
		aa.ag.errorpf(lambda, "%s must not have options", aa.what)
		return nil
	}
	return lambda
}
