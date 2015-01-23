package eval

import (
	"bytes"
	"fmt"
	"os"

	"github.com/elves/elvish/parse"
)

// Definition of Op and friends and combinators.

// Op operates on an Evaluator.
type Op func(*Evaluator)

// typeStar is either a single Type or a Kleene star of a Type.
type typeStar struct {
	t    Type
	star bool
}

func (ts typeStar) String() string {
	s := ts.t.String()
	if ts.star {
		s += "*"
	}
	return s
}

// typeRun is a run of typeStar's.
type typeRun []typeStar

func (tr typeRun) String() string {
	var buf bytes.Buffer
	for i, ts := range tr {
		if i > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(ts.String())
	}
	return buf.String()
}

// count returns the least number of Type's in a typeRun, and whether the
// actual number could be larger.
func (tr typeRun) count() (ntypes int, more bool) {
	for _, ts := range tr {
		if ts.star {
			more = true
		} else {
			ntypes++
		}
	}
	return
}

// mayCountTo returns whether a run of n Type's may match the typeRun.
func (tr typeRun) mayCountTo(n int) bool {
	ntypes, more := tr.count()
	if more {
		return n >= ntypes
	} else {
		return n == ntypes
	}
}

// newFixedTypeRun returns a typeRun where all typeStar's are simple
// non-starred types.
func newFixedTypeRun(ts ...Type) typeRun {
	tr := make([]typeStar, len(ts))
	for i, t := range ts {
		tr[i].t = t
	}
	return tr
}

// newFixedTypeRun returns a typeRun representing n types of t, and followed by
// a star of t if v is true.
func newHomoTypeRun(t Type, n int, v bool) typeRun {
	tr := make([]typeStar, n)
	for i := 0; i < n; i++ {
		tr[i].t = t
	}
	if v {
		tr = append(tr, typeStar{t, true})
	}
	return tr
}

// valuesOp operates on an Evaluator and results in some values.
type valuesOp struct {
	tr typeRun
	f  func(*Evaluator) []Value
}

// portOp operates on an Evaluator and results in a port.
type portOp func(*Evaluator) *port

// stateUpdatesOp operates on an Evaluator and results in a receiving channel
// of StateUpdate's.
type stateUpdatesOp func(*Evaluator) <-chan *StateUpdate

func combineChunk(ops []valuesOp) Op {
	return func(ev *Evaluator) {
		for _, op := range ops {
			s := op.f(ev)
			if ev.statusCb != nil {
				ev.statusCb(s)
			}
		}
	}
}

func combineClosure(argNames []string, ops []valuesOp, captured map[string]Type) valuesOp {
	op := combineChunk(ops)
	// BUG(xiaq): Closure arguments is (again) not supported
	f := func(ev *Evaluator) []Value {
		evCaptured := make(map[string]Variable, len(captured))
		for name := range captured {
			evCaptured[name] = ev.ResolveVar(name)
		}
		return []Value{NewClosure(argNames, op, evCaptured)}
	}
	return valuesOp{newFixedTypeRun(ClosureType{}), f}
}

var noExitus = newFailure("no exitus")

func combinePipeline(ops []stateUpdatesOp, p parse.Pos) valuesOp {
	f := func(ev *Evaluator) []Value {
		var nextIn *port
		updates := make([]<-chan *StateUpdate, len(ops))
		// For each form, create a dedicated Evaluator and run
		for i, op := range ops {
			newEv := ev.copy(fmt.Sprintf("form op %v", op))
			if i > 0 {
				newEv.ports[0] = nextIn
			}
			if i < len(ops)-1 {
				// Each internal port pair consists of a (byte) pipe pair and a
				// channel.
				// os.Pipe sets O_CLOEXEC, which is what we want.
				reader, writer, e := os.Pipe()
				if e != nil {
					ev.errorf(p, "failed to create pipe: %s", e)
				}
				// TODO Buffered channel?
				ch := make(chan Value)
				newEv.ports[1] = &port{
					f: writer, ch: ch, closeF: true, closeCh: true}
				nextIn = &port{
					f: reader, ch: ch, closeF: true, closeCh: false}
			}
			updates[i] = op(newEv)
		}
		// Collect exit values
		exits := make([]Value, len(ops))
		for i, update := range updates {
			ex := noExitus
			for up := range update {
				ex = up.Exitus
			}
			exits[i] = ex
		}
		return exits
	}
	return valuesOp{newHomoTypeRun(&StringType{}, len(ops), false), f}
}

func combineSpecialForm(op exitusOp, ports []portOp, p parse.Pos) stateUpdatesOp {
	// ev here is always a subevaluator created in combinePipeline, so it can
	// be safely modified.
	return func(ev *Evaluator) <-chan *StateUpdate {
		ev.applyPortOps(ports)
		return ev.execSpecial(op)
	}
}

func combineNonSpecialForm(cmdOp, argsOp valuesOp, ports []portOp, p parse.Pos) stateUpdatesOp {
	// ev here is always a subevaluator created in combinePipeline, so it can
	// be safely modified.
	return func(ev *Evaluator) <-chan *StateUpdate {
		ev.applyPortOps(ports)

		cmd := cmdOp.f(ev)
		expect := "expect a single string or closure value"
		if len(cmd) != 1 {
			ev.errorf(p, expect)
		}
		switch cmd[0].(type) {
		case String, *Closure:
		default:
			ev.errorf(p, expect)
		}

		args := argsOp.f(ev)
		return ev.execNonSpecial(cmd[0], args)
	}
}

func combineSpaced(ops []valuesOp) valuesOp {
	tr := make(typeRun, 0, len(ops))
	for _, op := range ops {
		tr = append(tr, op.tr...)
	}

	f := func(ev *Evaluator) []Value {
		// Use number of compound expressions as an estimation of the number
		// of values
		vs := make([]Value, 0, len(ops))
		for _, op := range ops {
			us := op.f(ev)
			vs = append(vs, us...)
		}
		return vs
	}
	return valuesOp{tr, f}
}

func compound(ev *Evaluator, lhs, rhs Value) Value {
	return NewString(lhs.String() + rhs.String())
}

func combineCompound(ops []valuesOp) valuesOp {
	n := 1
	more := false
	for _, op := range ops {
		m, b := op.tr.count()
		n *= m
		more = more || b
	}

	f := func(ev *Evaluator) []Value {
		vs := ops[0].f(ev)
		for _, op := range ops[1:] {
			us := op.f(ev)
			if len(us) == 1 {
				u := us[0]
				for i := range vs {
					vs[i] = compound(ev, vs[i], u)
				}
			} else {
				// Do a cartesian product
				newvs := make([]Value, len(vs)*len(us))
				for i, v := range vs {
					for j, u := range us {
						newvs[i*len(us)+j] = compound(ev, v, u)
					}
				}
				vs = newvs
			}
		}
		return vs
	}
	return valuesOp{newHomoTypeRun(StringType{}, n, more), f}
}

func literalValue(v ...Value) valuesOp {
	tr := make(typeRun, len(v))
	for i := range tr {
		tr[i].t = v[i].Type()
	}
	f := func(e *Evaluator) []Value {
		return v
	}
	return valuesOp{tr, f}
}

func makeString(text string) valuesOp {
	return literalValue(NewString(text))
}

func makeVar(cp *Compiler, name string, p parse.Pos) valuesOp {
	tr := newFixedTypeRun(cp.mustResolveVar(name, p))
	f := func(ev *Evaluator) []Value {
		variable := ev.ResolveVar(name)
		if variable.valuePtr == nil {
			ev.errorf(p, "variable $%s not found; the compiler has a bug", name)
		}
		return []Value{*variable.valuePtr}
	}
	return valuesOp{tr, f}
}

func combineSubscript(cp *Compiler, left, right valuesOp, lp, rp parse.Pos) valuesOp {
	if !left.tr.mayCountTo(1) {
		// TODO Also check at runtime
		cp.errorf(lp, "left operand of subscript must be a single value")
	}
	var t Type
	switch left.tr[0].t.(type) {
	case EnvType, StringType:
		t = StringType{}
	case TableType, AnyType:
		t = AnyType{}
	default:
		cp.errorf(lp, "left operand of subscript must be of type string, env, table or any")
	}

	if !right.tr.mayCountTo(1) {
		// TODO Also check at runtime
		cp.errorf(rp, "right operand of subscript must be a single value")
	}
	if _, ok := right.tr[0].t.(StringType); !ok {
		cp.errorf(rp, "right operand of subscript must be of type string")
	}

	f := func(ev *Evaluator) []Value {
		l := left.f(ev)
		r := right.f(ev)
		return []Value{evalSubscript(ev, l[0], r[0], lp, rp)}
	}
	return valuesOp{newFixedTypeRun(t), f}
}

func combineTable(list valuesOp, keys []valuesOp, values []valuesOp, p parse.Pos) valuesOp {
	f := func(ev *Evaluator) []Value {
		t := NewTable()
		t.append(list.f(ev)...)
		for i, kop := range keys {
			vop := values[i]
			ks := kop.f(ev)
			vs := vop.f(ev)
			if len(ks) != len(vs) {
				ev.errorf(p, "Number of keys doesn't match number of values: %d vs. %d", len(ks), len(vs))
			}
			for j, k := range ks {
				t.Dict[k.String()] = vs[j]
			}
		}
		return []Value{t}
	}
	return valuesOp{newFixedTypeRun(TableType{}), f}
}

func combineChanCapture(op valuesOp) valuesOp {
	tr := typeRun{typeStar{AnyType{}, true}}
	f := func(ev *Evaluator) []Value {
		vs := []Value{}
		newEv := ev.copy(fmt.Sprintf("channel output capture %v", op))
		ch := make(chan Value)
		newEv.ports[1] = &port{ch: ch}
		go func() {
			for v := range ch {
				vs = append(vs, v)
			}
		}()
		op.f(newEv)
		newEv.closePorts()
		return vs
	}
	return valuesOp{tr, f}
}
