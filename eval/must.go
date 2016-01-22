package eval

import (
	"strconv"

	"github.com/elves/elvish/parse-ng"
)

type muster struct {
	ec   *evalCtx
	what string
	p    int
	vs   []Value
}

func (m *muster) error(want, got string) {
	m.ec.errorf(m.p, "%s must be %s; got %s", m.what, want, got)
}

func (ec *evalCtx) must(vs []Value, what string, pos int) *muster {
	return &muster{ec, what, pos, vs}
}

func (m *muster) mustLen(l int) {
	if len(m.vs) != l {
		m.ec.errorf(m.p, "%s must be exactly %d value; got %d", m.what, l, len(m.vs))
	}
}

func (m *muster) mustOne() Value {
	m.mustLen(1)
	return m.vs[0]
}

func (m *muster) zerothMustStr() str {
	v := m.vs[0]
	s, ok := v.(str)
	if !ok {
		m.ec.errorf(m.p, "%s must be a string; got %s (type %s)",
			m.what, v.Repr(), v.Type())
	}
	return s
}

func (m *muster) mustOneStr() str {
	m.mustLen(1)
	return m.zerothMustStr()
}

func (m *muster) zerothMustInt() int {
	s := m.zerothMustStr()
	i, err := strconv.Atoi(string(s))
	if err != nil {
		m.ec.errorf(m.p, "%s must be an integer; got %s", m.what, s)
	}
	return i
}

func (m *muster) mustOneInt() int {
	m.mustLen(1)
	return m.zerothMustInt()
}

func (m *muster) zerothMustNonNegativeInt() int {
	i := m.zerothMustInt()
	if i < 0 {
		m.ec.errorf(m.p, "%s must be non-negative; got %d", m.what, i)
	}
	return i
}

func (m *muster) mustOneNonNegativeInt() int {
	m.mustLen(1)
	return m.zerothMustNonNegativeInt()
}

func mustPrimary(cp *compiler, cn *parse.Compound, msg string) *parse.Primary {
	if len(cn.Indexeds) != 1 || len(cn.Indexeds[0].Indicies) > 0 {
		// cp.errorf(cn.Begin, msg)
	}
	return cn.Indexeds[0].Head
}

func mustVariableOrString(cp *compiler, cn *parse.Compound, msg string) (*parse.Primary, string) {
	pn := mustPrimary(cp, cn, msg)
	switch pn.Type {
	case parse.Variable:
		return pn, pn.Value[1:]
	case parse.Bareword, parse.SingleQuoted, parse.DoubleQuoted:
		return pn, pn.Value
	default:
		// cp.errorf(cn.Pos, msg)
		return nil, ""
	}
}

// mustVariable musts that a Compound contains exactly one Primary of type
// Variable.
func mustVariablePrimary(cp *compiler, cn *parse.Compound, msg string) (*parse.Primary, string) {
	pn, text := mustVariableOrString(cp, cn, msg)
	if pn.Type != parse.Variable {
		// cp.errorf(pn.Pos, msg)
	}
	return pn, text
}

// mustString musts that a Compound contains exactly one Primary of type
// Variable.
func mustString(cp *compiler, cn *parse.Compound, msg string) (*parse.Primary, string) {
	pn, text := mustVariableOrString(cp, cn, msg)
	if pn.Type == parse.Variable {
		// cp.errorf(pn.Pos, msg)
	}
	return pn, text
}

/*
// mustStartWithVariabl musts the first compound of the form is a
// VariablePrimary. This is merely for better error messages; No actual
// processing is done.
func mustStartWithVariable(cp *compiler, fn *parse.Form, form string) {
	if len(fn.Args.Nodes) == 0 {
		cp.errorf(fn.Pos, "expect variable after %s", form)
	}
	mustVariablePrimary(cp, fn.Args.Nodes[0], "expect variable")
}
*/
