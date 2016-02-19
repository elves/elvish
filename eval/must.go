package eval

import (
	"fmt"
	"strconv"

	"github.com/elves/elvish/parse"
)

type muster struct {
	ec         *EvalCtx
	what       string
	begin, end int
	vs         []Value
}

func (m *muster) error(want, gotfmt string, gotargs ...interface{}) {
	m.ec.errorpf(m.begin, m.end, "%s must be %s; got "+gotfmt,
		append([]interface{}{m.what, want}, gotargs...))
}

func (ec *EvalCtx) must(vs []Value, what string, begin, end int) *muster {
	return &muster{ec, what, begin, end, vs}
}

func (m *muster) mustLen(l int) {
	if len(m.vs) != l {
		m.error(fmt.Sprintf("%d values", l), "%d", len(m.vs))
	}
}

func (m *muster) mustOne() Value {
	m.mustLen(1)
	return m.vs[0]
}

func (m *muster) zerothMustStr() String {
	v := m.vs[0]
	s, ok := v.(String)
	if !ok {
		m.error("a string", "%s (type %s)", v.Repr(NoPretty), v.Kind())
	}
	return s
}

func (m *muster) mustOneStr() String {
	m.mustLen(1)
	return m.zerothMustStr()
}

func (m *muster) zerothMustInt() int {
	s := m.zerothMustStr()
	i, err := strconv.Atoi(string(s))
	if err != nil {
		m.error("an integer", "%s", s)
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
		m.error("non-negative", "%d", i)
	}
	return i
}

func (m *muster) mustOneNonNegativeInt() int {
	m.mustLen(1)
	return m.zerothMustNonNegativeInt()
}

func onePrimary(cn *parse.Compound) *parse.Primary {
	if len(cn.Indexings) == 1 && len(cn.Indexings[0].Indicies) == 0 {
		return cn.Indexings[0].Head
	}
	return nil
}

func oneString(cn *parse.Compound) (string, bool) {
	pn := onePrimary(cn)
	if pn != nil {
		switch pn.Type {
		case parse.Bareword, parse.SingleQuoted, parse.DoubleQuoted:
			return pn.Value, true
		}
	}
	return "", false
}

func mustPrimary(cp *compiler, cn *parse.Compound, msg string) *parse.Primary {
	p := onePrimary(cn)
	if p == nil {
		cp.errorpf(cn.Begin(), cn.End(), msg)
	}
	return p
}

// mustString musts that a Compound contains exactly one Primary of type
// Variable.
func mustString(cp *compiler, cn *parse.Compound, msg string) string {
	s, ok := oneString(cn)
	if !ok {
		cp.errorpf(cn.Begin(), cn.End(), msg)
	}
	return s
}
