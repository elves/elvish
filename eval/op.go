package eval

import (
	"fmt"
	"os"

	"github.com/elves/elvish/parse"
)

const (
	pipelineChanBufferSize = 32
)

// Definition of Op and friends and combinators.

// Op operates on an evalCtx.
type Op func(*evalCtx)

// valuesOp operates on an evalCtx and results in some values.
type valuesOp struct {
	tr typeRun
	f  func(*evalCtx) []Value
}

// portOp operates on an evalCtx and results in a port.
type portOp func(*evalCtx) *port

// stateUpdatesOp operates on an evalCtx and results in a receiving channel
// of StateUpdate's.
type stateUpdatesOp func(*evalCtx) <-chan *stateUpdate

func combineChunk(ops []valuesOp) valuesOp {
	f := func(ec *evalCtx) []Value {
		for _, op := range ops {
			s := op.f(ec)
			if HasFailure(s) {
				return s
			}
		}
		return []Value{success}
	}
	return valuesOp{newHomoTypeRun(&exitusType{}, 1, true), f}
}

func combineClosure(argNames []string, op valuesOp, up map[string]Type) valuesOp {
	f := func(ec *evalCtx) []Value {
		evCaptured := make(map[string]Variable, len(up))
		for name := range up {
			evCaptured[name] = ec.ResolveVar("", name)
		}
		return []Value{newClosure(argNames, op, evCaptured)}
	}
	return valuesOp{newFixedTypeRun(callableType{}), f}
}

var noExitus = newFailure("no exitus")

func combinePipeline(ops []stateUpdatesOp, p parse.Pos) valuesOp {
	f := func(ec *evalCtx) []Value {
		var nextIn *port
		updates := make([]<-chan *stateUpdate, len(ops))
		// For each form, create a dedicated evalCtx and run
		for i, op := range ops {
			newEc := ec.copy(fmt.Sprintf("form op %v", op))
			if i > 0 {
				newEc.ports[0] = nextIn
			}
			if i < len(ops)-1 {
				// Each internal port pair consists of a (byte) pipe pair and a
				// channel.
				// os.Pipe sets O_CLOEXEC, which is what we want.
				reader, writer, e := os.Pipe()
				if e != nil {
					ec.errorf(p, "failed to create pipe: %s", e)
				}
				ch := make(chan Value, pipelineChanBufferSize)
				newEc.ports[1] = &port{
					f: writer, ch: ch, closeF: true, closeCh: true}
				nextIn = &port{
					f: reader, ch: ch, closeF: true, closeCh: false}
			}
			updates[i] = op(newEc)
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
	return valuesOp{newHomoTypeRun(&exitusType{}, len(ops), false), f}
}

func combineSpecialForm(op exitusOp, ports []portOp, p parse.Pos) stateUpdatesOp {
	// ec here is always a subevaler created in combinePipeline, so it can
	// be safely modified.
	return func(ec *evalCtx) <-chan *stateUpdate {
		ec.applyPortOps(ports)
		return ec.execSpecial(op)
	}
}

func combineNonSpecialForm(cmdOp, argsOp valuesOp, ports []portOp, p parse.Pos) stateUpdatesOp {
	// ec here is always a subevaler created in combinePipeline, so it can
	// be safely modified.
	return func(ec *evalCtx) <-chan *stateUpdate {
		ec.applyPortOps(ports)

		cmd := cmdOp.f(ec)
		expect := "expect a single string or closure value"
		if len(cmd) != 1 {
			ec.errorf(p, expect)
		}
		switch cmd[0].(type) {
		case str, callable:
		default:
			ec.errorf(p, expect)
		}

		args := argsOp.f(ec)
		return ec.execNonSpecial(cmd[0], args)
	}
}

func combineSpaced(ops []valuesOp) valuesOp {
	tr := make(typeRun, 0, len(ops))
	for _, op := range ops {
		tr = append(tr, op.tr...)
	}

	f := func(ec *evalCtx) []Value {
		// Use number of compound expressions as an estimation of the number
		// of values
		vs := make([]Value, 0, len(ops))
		for _, op := range ops {
			us := op.f(ec)
			vs = append(vs, us...)
		}
		return vs
	}
	return valuesOp{tr, f}
}

func compound(lhs, rhs Value) Value {
	return str(toString(lhs) + toString(rhs))
}

func combineCompound(ops []valuesOp) valuesOp {
	// Non-proper compound: just return the sole subscript
	if len(ops) == 1 {
		return ops[0]
	}

	n := 1
	more := false
	for _, op := range ops {
		m, b := op.tr.count()
		n *= m
		more = more || b
	}

	f := func(ec *evalCtx) []Value {
		vs := []Value{str("")}
		for _, op := range ops {
			us := op.f(ec)
			if len(us) == 1 {
				u := us[0]
				for i := range vs {
					vs[i] = compound(vs[i], u)
				}
			} else {
				// Do a cartesian product
				newvs := make([]Value, len(vs)*len(us))
				for i, v := range vs {
					for j, u := range us {
						newvs[i*len(us)+j] = compound(v, u)
					}
				}
				vs = newvs
			}
		}
		return vs
	}
	return valuesOp{newHomoTypeRun(stringType{}, n, more), f}
}

func literalValue(v ...Value) valuesOp {
	tr := make(typeRun, len(v))
	for i := range tr {
		tr[i].t = v[i].Type()
	}
	f := func(e *evalCtx) []Value {
		return v
	}
	return valuesOp{tr, f}
}

func makeString(text string) valuesOp {
	return literalValue(str(text))
}

func makeVar(cc *compileCtx, qname string, p parse.Pos) valuesOp {
	ns, name := splitQualifiedName(qname)
	tr := newFixedTypeRun(cc.mustResolveVar(ns, name, p))
	f := func(ec *evalCtx) []Value {
		variable := ec.ResolveVar(ns, name)
		if variable == nil {
			ec.errorf(p, "variable $%s not found; the compiler has a bug", name)
		}
		return []Value{variable.Get()}
	}
	return valuesOp{tr, f}
}

func combineSubscript(cc *compileCtx, left, right valuesOp, lp, rp parse.Pos) valuesOp {
	if !left.tr.mayCountTo(1) {
		cc.errorf(lp, "left operand of subscript must be a single value")
	}
	var t Type
	switch left.tr[0].t.(type) {
	case stringType:
		t = stringType{}
	case tableType, anyType:
		t = anyType{}
	default:
		cc.errorf(lp, "left operand of subscript must be of type string, env, table or any")
	}

	if !right.tr.mayCountTo(1) {
		cc.errorf(rp, "right operand of subscript must be a single value")
	}
	if _, ok := right.tr[0].t.(stringType); !ok {
		cc.errorf(rp, "right operand of subscript must be of type string")
	}

	f := func(ec *evalCtx) []Value {
		l := left.f(ec)
		if len(l) != 1 {
			ec.errorf(lp, "left operand of subscript must be a single value")
		}
		r := right.f(ec)
		if len(r) != 1 {
			ec.errorf(rp, "right operand of subscript must be a single value")
		}
		return []Value{evalSubscript(ec, l[0], r[0], lp, rp)}
	}
	return valuesOp{newFixedTypeRun(t), f}
}

func combineTable(list valuesOp, keys []valuesOp, values []valuesOp, p parse.Pos) valuesOp {
	f := func(ec *evalCtx) []Value {
		t := newTable()
		t.append(list.f(ec)...)
		for i, kop := range keys {
			vop := values[i]
			ks := kop.f(ec)
			vs := vop.f(ec)
			if len(ks) != len(vs) {
				ec.errorf(p, "Number of keys doesn't match number of values: %d vs. %d", len(ks), len(vs))
			}
			for j, k := range ks {
				t.Dict[toString(k)] = vs[j]
			}
		}
		return []Value{t}
	}
	return valuesOp{newFixedTypeRun(tableType{}), f}
}

func combineChanCapture(op valuesOp) valuesOp {
	tr := typeRun{typeStar{anyType{}, true}}
	f := func(ec *evalCtx) []Value {
		vs := []Value{}
		newEc := ec.copy(fmt.Sprintf("channel output capture %v", op))
		ch := make(chan Value)
		newEc.ports[1] = &port{ch: ch}
		go func() {
			for v := range ch {
				vs = append(vs, v)
			}
		}()
		op.f(newEc)
		newEc.closePorts()
		return vs
	}
	return valuesOp{tr, f}
}
