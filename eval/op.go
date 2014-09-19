package eval

import (
	"fmt"
	"os"

	"github.com/xiaq/elvish/parse"
)

// Definition of Op and friends and combinators.

// Op operates on an Evaluator.
type Op func(*Evaluator)

// typeStar is either a single Type or a Kleene star of a Type.
type typeStar struct {
	t    Type
	star bool
}

// typeRun is a run of typeStar's.
type typeRun []typeStar

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

func combineClosure(ops []valuesOp, enclosed map[string]Type, bounds [2]StreamType) valuesOp {
	op := combineChunk(ops)
	// BUG(xiaq): Closure arguments is (again) not supported
	f := func(ev *Evaluator) []Value {
		enclosed := make(map[string]*Value, len(enclosed))
		for name := range enclosed {
			enclosed[name] = ev.scope[name]
		}
		return []Value{NewClosure(nil, op, enclosed, bounds)}
	}
	return valuesOp{newFixedTypeRun(&ClosureType{bounds}), f}
}

func combinePipeline(ops []stateUpdatesOp, bounds [2]StreamType, internals []StreamType, p parse.Pos) valuesOp {
	f := func(ev *Evaluator) []Value {
		// TODO(xiaq): Should catch when compiling
		if !ev.ports[0].compatible(bounds[0]) {
			ev.errorf(p, "pipeline input not satisfiable")
		}
		if !ev.ports[1].compatible(bounds[1]) {
			ev.errorf(p, "pipeline output not satisfiable")
		}
		var nextIn *port
		updates := make([]<-chan *StateUpdate, len(ops))
		// For each form, create a dedicated Evaluator and run
		for i, op := range ops {
			newEv := ev.copy(fmt.Sprintf("form op %v", op), false)
			if i > 0 {
				newEv.ports[0] = nextIn
			}
			if i < len(ops)-1 {
				switch internals[i] {
				case unusedStream:
					newEv.ports[1] = nil
					nextIn = nil
				case fdStream:
					// os.Pipe sets O_CLOEXEC, which is what we want.
					reader, writer, e := os.Pipe()
					if e != nil {
						ev.errorf(p, "failed to create pipe: %s", e)
					}
					newEv.ports[1] = &port{f: writer, shouldClose: true}
					nextIn = &port{f: reader, shouldClose: true}
				case chanStream:
					// TODO Buffered channel?
					ch := make(chan Value)
					// Only the writer closes the channel port
					newEv.ports[1] = &port{ch: ch, shouldClose: true}
					nextIn = &port{ch: ch}
				default:
					panic("bad StreamType value")
				}
			}
			updates[i] = op(newEv)
		}
		// Collect exit values
		exits := make([]Value, len(ops))
		for i, update := range updates {
			for up := range update {
				exits[i] = NewString(up.Msg)
			}
		}
		return exits
	}
	return valuesOp{newHomoTypeRun(&StringType{}, len(ops), false), f}
}

func combineForm(cmd valuesOp, tlist valuesOp, ports []portOp, a *formAnnotation, p parse.Pos) stateUpdatesOp {
	return func(ev *Evaluator) <-chan *StateUpdate {
		// XXX Currently it's guaranteed that cmd evaluates into a single
		// Value.
		cmd := cmd.f(ev)[0]
		cmdStr := cmd.String()
		fm := &form{
			name: cmdStr,
		}
		if tlist.f != nil {
			fm.args = tlist.f(ev)
		}

		switch a.commandType {
		case commandBuiltinFunction:
			fm.Command.Func = a.builtinFunc.fn
		case commandBuiltinSpecial:
			fm.Command.Special = a.specialOp
		case commandDefinedFunction:
			v, ok1 := ev.scope["fn-"+cmdStr]
			fn, ok2 := (*v).(*Closure)
			if !(ok1 && ok2) {
				panic("Compiler bug")
			}
			fm.Command.Closure = fn
		case commandClosure:
			fm.Command.Closure = cmd.(*Closure)
		case commandExternal:
			path, e := ev.search(cmdStr)
			if e != nil {
				ev.errorf(p, "%s", e)
			}
			fm.Command.Path = path
		default:
			panic("bad commandType value")
		}

		newEv := ev.copy(fmt.Sprintf("form redir %v", fm), true)
		newEv.growPorts(len(ports))

		for i, op := range ports {
			if op != nil {
				newEv.ports[i] = op(ev)
			}
		}

		return newEv.execForm(fm)
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

func caret(ev *Evaluator, lhs, rhs Value) Value {
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
					vs[i] = caret(ev, vs[i], u)
				}
			} else {
				// Do a cartesian product
				newvs := make([]Value, len(vs)*len(us))
				for i, v := range vs {
					for j, u := range us {
						newvs[i*len(us)+j] = caret(ev, v, u)
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
	tr := newFixedTypeRun(cp.resolveVar(name, p))
	f := func(ev *Evaluator) []Value {
		val, ok := ev.scope[name]
		if !ok {
			panic("Compiler bug")
		}
		return []Value{*val}
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
	case EnvType:
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
				t.Dict[k] = vs[j]
			}
		}
		return []Value{t}
	}
	return valuesOp{newFixedTypeRun(TableType{}), f}
}

func combineOutputCapture(op valuesOp, bounds [2]StreamType) valuesOp {
	// XXX Wrong type; tr should be variadic
	tr := newFixedTypeRun()
	f := func(ev *Evaluator) []Value {
		vs := []Value{}
		newEv := ev.copy(fmt.Sprintf("output capture %v", op), false)
		newEv.ports = make([]*port, len(ev.ports))
		copy(newEv.ports, ev.ports)
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
