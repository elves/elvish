package eval

import (
	"fmt"
	"os"

	"github.com/xiaq/elvish/parse"
)

// Definition of Op and friends and combinators.

// Op operates on an Evaluator.
type Op func(*Evaluator)

// valuesOp operates on an Evaluator and results in some values.
type valuesOp func(*Evaluator) []Value

// portOp operates on an Evaluator and results in a port.
type portOp func(*Evaluator) *port

// stateUpdatesOp operates on an Evaluator and results in a receiving channel
// of StateUpdate's.
type stateUpdatesOp func(*Evaluator) <-chan *StateUpdate

func combineChunk(ops []valuesOp) Op {
	return func(ev *Evaluator) {
		for _, op := range ops {
			s := op(ev)
			if ev.statusCb != nil {
				ev.statusCb(s)
			}
		}
	}
}

func combineClosure(ops []valuesOp, enclosed map[string]Type, bounds [2]StreamType) valuesOp {
	op := combineChunk(ops)
	// BUG(xiaq): Closure arguments is (again) not supported
	return func(ev *Evaluator) []Value {
		enclosed := make(map[string]*Value, len(enclosed))
		for name := range enclosed {
			enclosed[name] = ev.scope[name]
		}
		return []Value{NewClosure(nil, op, enclosed, bounds)}
	}
}

func combinePipeline(n parse.Node, ops []stateUpdatesOp, bounds [2]StreamType, internals []StreamType) valuesOp {
	return func(ev *Evaluator) []Value {
		// TODO(xiaq): Should catch when compiling
		if !ev.ports[0].compatible(bounds[0]) {
			ev.errorfNode(n, "pipeline input not satisfiable")
		}
		if !ev.ports[1].compatible(bounds[1]) {
			ev.errorfNode(n, "pipeline output not satisfiable")
		}
		var nextIn *port
		updates := make([]<-chan *StateUpdate, len(ops))
		// For each form, create a dedicated Evaluator and run
		for i, op := range ops {
			newEv := ev.copy(fmt.Sprintf("<form op %v>", op), false)
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
						ev.errorfNode(n, "failed to create pipe: %s", e)
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
}

func combineForm(n parse.Node, cmd valuesOp, tlist valuesOp, ports []portOp, a *formAnnotation) stateUpdatesOp {
	return func(ev *Evaluator) <-chan *StateUpdate {
		// XXX Currently it's guaranteed that cmd evaluates into a single
		// Value.
		cmd := cmd(ev)[0]
		cmdStr := cmd.String()
		fm := &form{
			name: cmdStr,
		}
		if tlist != nil {
			fm.args = tlist(ev)
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
				panic("Checker bug")
			}
			fm.Command.Closure = fn
		case commandClosure:
			fm.Command.Closure = cmd.(*Closure)
		case commandExternal:
			path, e := ev.search(cmdStr)
			if e != nil {
				ev.errorfNode(n, "%s", e)
			}
			fm.Command.Path = path
		default:
			panic("bad commandType value")
		}

		newEv := ev.copy(fmt.Sprintf("<form redir %v>", fm), true)
		newEv.growPorts(len(ports))

		for i, op := range ports {
			if op != nil {
				newEv.ports[i] = op(ev)
			}
		}

		return newEv.execForm(fm)
	}
}

func combineTermList(ops []valuesOp) valuesOp {
	return func(ev *Evaluator) []Value {
		// Use number of terms as an estimation of the number of values
		vs := make([]Value, 0, len(ops))
		for _, op := range ops {
			us := op(ev)
			vs = append(vs, us...)
		}
		return vs
	}
}

func combineTerm(ops []valuesOp) valuesOp {
	return func(ev *Evaluator) []Value {
		vs := ops[0](ev)
		for _, op := range ops[1:] {
			us := op(ev)
			if len(us) == 1 {
				u := us[0]
				for i := range vs {
					vs[i] = vs[i].Caret(ev, u)
				}
			} else {
				// Do a cartesian product
				newvs := make([]Value, len(vs)*len(us))
				for i, v := range vs {
					for j, u := range us {
						newvs[i*len(us)+j] = v.Caret(ev, u)
					}
				}
				vs = newvs
			}
		}
		return vs
	}
}

func literalValue(v ...Value) valuesOp {
	return func(e *Evaluator) []Value {
		return v
	}
}

func makeString(text string) valuesOp {
	return literalValue(NewString(text))
}

func makeVar(name string) valuesOp {
	return func(ev *Evaluator) []Value {
		val, ok := ev.scope[name]
		if !ok {
			panic("Checker bug")
		}
		return []Value{*val}
	}
}

func combineTable(n parse.Node, list valuesOp, keys []valuesOp, values []valuesOp) valuesOp {
	return func(ev *Evaluator) []Value {
		t := NewTable()
		t.append(list(ev)...)
		for i, kop := range keys {
			vop := values[i]
			ks := kop(ev)
			vs := vop(ev)
			if len(ks) != len(vs) {
				ev.errorfNode(n, "Number of keys doesn't match number of values: %d vs. %d", len(ks), len(vs))
			}
			for j, k := range ks {
				t.Dict[k] = vs[j]
			}
		}
		return []Value{t}
	}
}

func combineOutputCapture(op valuesOp, bounds [2]StreamType) valuesOp {
	return func(ev *Evaluator) []Value {
		vs := []Value{}
		newEv := ev.copy(fmt.Sprintf("<output capture %v>", op), true)
		newEv.ports = make([]*port, len(ev.ports))
		copy(newEv.ports, ev.ports)
		ch := make(chan Value)
		newEv.ports[1] = &port{ch: ch}
		go func() {
			for v := range ch {
				vs = append(vs, v)
			}
		}()
		op(newEv)
		return vs
	}
}
