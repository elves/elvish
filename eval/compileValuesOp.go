package eval

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/elves/elvish/glob"
	"github.com/elves/elvish/parse"
)

func (cp *compiler) compound(n *parse.Compound) ValuesOp {
	if len(n.Indexings) == 0 {
		return literalStr("")
	}

	tilde := false
	indexings := n.Indexings
	begins := indexingBegins(n.Indexings)[1:]

	if n.Indexings[0].Head.Type == parse.Tilde {
		// A lone ~.
		if len(n.Indexings) == 1 {
			return func(ec *EvalCtx) []Value {
				return []Value{String(mustGetHome(""))}
			}
		}
		tilde = true
		indexings = indexings[1:]
		begins = begins[1:]
	}

	ops := cp.indexings(indexings)

	return func(ec *EvalCtx) []Value {
		// Accumulator.
		vs := ops[0](ec)

		// Logger.Printf("concatenating %v with %d more", vs, len(ops)-1)

		for _, op := range ops[1:] {
			us := op(ec)
			vs = outerProduct(vs, us, cat)
			// Logger.Printf("with %v => %v", us, vs)
		}
		if tilde {
			newvs := make([]Value, len(vs))
			for i, v := range vs {
				newvs[i] = doTilde(v)
			}
			vs = newvs
		}
		hasGlob := false
		for _, v := range vs {
			if _, ok := v.(GlobPattern); ok {
				hasGlob = true
				break
			}
		}
		if hasGlob {
			newvs := make([]Value, 0, len(vs))
			for _, v := range vs {
				if gp, ok := v.(GlobPattern); ok {
					// Logger.Printf("globbing %v", gp)
					newvs = append(newvs, doGlob(gp)...)
				} else {
					newvs = append(newvs, v)
				}
			}
			vs = newvs
		}
		return vs
	}
}

func cat(lhs, rhs Value) Value {
	switch lhs := lhs.(type) {
	case String:
		switch rhs := rhs.(type) {
		case String:
			return lhs + rhs
		case GlobPattern:
			segs := stringToSegments(string(lhs))
			// We know rhs contains exactly one segment.
			segs = append(segs, rhs.Segments[0])
			return GlobPattern{segs}
		}
	case GlobPattern:
		// NOTE Modifies lhs in place.
		switch rhs := rhs.(type) {
		case String:
			lhs.append(stringToSegments(string(rhs))...)
			return lhs
		case GlobPattern:
			// We know rhs contains exactly one segment.
			lhs.append(rhs.Segments[0])
			return lhs
		}
	}
	throw(fmt.Errorf("unsupported concat: %s and %s", lhs.Kind(), rhs.Kind()))
	panic("unreachable")
}

func outerProduct(vs []Value, us []Value, f func(Value, Value) Value) []Value {
	ws := make([]Value, len(vs)*len(us))
	nu := len(us)
	for i, v := range vs {
		for j, u := range us {
			ws[i*nu+j] = f(v, u)
		}
	}
	return ws
}

func doTilde(v Value) Value {
	switch v := v.(type) {
	case String:
		s := string(v)
		i := strings.Index(s, "/")
		var uname, rest string
		if i == -1 {
			uname = s
		} else {
			uname = s[:i]
			rest = s[i+1:]
		}
		dir := mustGetHome(uname)
		return String(path.Join(dir, rest))
	case GlobPattern:
		if len(v.Segments) == 0 || v.Segments[0].Type != glob.Literal {
			throw(errors.New("cannot determine user name from glob pattern"))
		}
		s := v.Segments[0].Literal
		// Find / in the first segment to determine the username.
		i := strings.Index(s, "/")
		if i == -1 {
			throw(errors.New("cannot determine user name from glob pattern"))
		}
		uname := s[:i]
		dir := mustGetHome(uname)
		// Replace ~uname in first segment with the found path.
		v.Segments[0].Literal = dir + s[i:]
		return v
	default:
		throw(fmt.Errorf("tilde doesn't work on value of type %s", v.Kind()))
		panic("unreachable")
	}
}

func (cp *compiler) array(n *parse.Array) ValuesOp {
	return catValuesOps(cp.compounds(n.Compounds))
}

func catValuesOps(ops []ValuesOp) ValuesOp {
	return func(ec *EvalCtx) []Value {
		// Use number of compound expressions as an estimation of the number
		// of values
		vs := make([]Value, 0, len(ops))
		for _, op := range ops {
			us := op(ec)
			vs = append(vs, us...)
		}
		return vs
	}
}

func (cp *compiler) indexing(n *parse.Indexing) ValuesOp {
	if len(n.Indicies) == 0 {
		return cp.primary(n.Head)
	}

	headOp := cp.primary(n.Head)
	indexOps := cp.arrays(n.Indicies)
	// p := n.Begin()
	indexPoses := make([]int, len(n.Indicies))
	for i, index := range n.Indicies {
		indexPoses[i] = index.Begin()
	}

	return func(ec *EvalCtx) []Value {
		vs := headOp(ec)
		for _, indexOp := range indexOps {
			indicies := indexOp(ec)
			newvs := make([]Value, 0, len(vs)*len(indicies))
			for _, v := range vs {
				newvs = append(newvs, mustIndexer(v, ec).Index(indicies)...)
			}
			vs = newvs
		}
		return vs
	}
}

func literalValues(v ...Value) ValuesOp {
	return func(e *EvalCtx) []Value {
		return v
	}
}

func literalStr(text string) ValuesOp {
	return literalValues(String(text))
}

func variable(qname string, p int) ValuesOp {
	splice, ns, name := parseVariable(qname)
	return func(ec *EvalCtx) []Value {
		variable := ec.ResolveVar(ns, name)
		if variable == nil {
			ec.errorf(p, "variable $%s not found", qname)
		}
		value := variable.Get()
		if splice {
			elemser, ok := value.(Elemser)
			if !ok {
				ec.errorf(p, "variable $%s (kind %s) cannot be spliced", qname, value.Kind())
			}
			return collectElems(elemser)
		}
		return []Value{value}
	}
}

func (cp *compiler) primary(n *parse.Primary) ValuesOp {
	switch n.Type {
	case parse.Bareword, parse.SingleQuoted, parse.DoubleQuoted:
		return literalStr(n.Value)
	case parse.Variable:
		qname := n.Value
		if !cp.registerVariableGet(qname) {
			cp.errorf(n.Begin(), "variable %s not found", n.Value)
		}
		return variable(qname, n.Begin())
	case parse.Wildcard:
		vs := []Value{GlobPattern{[]glob.Segment{
			wildcardToSegment(n.SourceText())}}}
		return func(ec *EvalCtx) []Value {
			return vs
		}
	case parse.Tilde:
		cp.errorf(n.Begin(), "compiler bug: Tilde not handled in .compound")
		return literalStr("~")
	case parse.ErrorCapture:
		return cp.errorCapture(n.Chunk)
	case parse.OutputCapture:
		return cp.outputCapture(n)
	case parse.List:
		op := cp.array(n.List)
		return func(ec *EvalCtx) []Value {
			return []Value{NewList(op(ec)...)}
		}
	case parse.Lambda:
		return cp.lambda(n)
	case parse.Map:
		return cp.map_(n)
	case parse.Braced:
		return cp.braced(n)
	default:
		cp.errorf(n.Begin(), "bad PrimaryType; parser bug")
		return literalStr(n.SourceText())
	}
}

func (cp *compiler) errorCapture(n *parse.Chunk) ValuesOp {
	op := cp.chunk(n)
	return func(ec *EvalCtx) []Value {
		return []Value{Error{ec.PEval(op)}}
	}
}

func (cp *compiler) outputCapture(n *parse.Primary) ValuesOp {
	op := cp.chunk(n.Chunk)
	// p := n.Chunk.Begin()
	return func(ec *EvalCtx) []Value {
		return captureOutput(ec, op)
	}
}

func captureOutput(ec *EvalCtx, op Op) []Value {
	vs := []Value{}
	newEc := ec.fork(fmt.Sprintf("channel output capture %v", op))

	pipeRead, pipeWrite, err := os.Pipe()
	if err != nil {
		throw(fmt.Errorf("failed to create pipe: %v", err))
	}
	bufferedPipeRead := bufio.NewReader(pipeRead)
	ch := make(chan Value, outputCaptureBufferSize)
	bytesCollected := make(chan bool)
	chCollected := make(chan bool)
	newEc.ports[1] = &Port{Chan: ch, File: pipeWrite, CloseFile: true}
	go func() {
		for v := range ch {
			vs = append(vs, v)
		}
		chCollected <- true
	}()
	go func() {
		for {
			line, err := bufferedPipeRead.ReadString('\n')
			if err == io.EOF {
				break
			} else if err != nil {
				// TODO report error
				log.Println(err)
				break
			}
			ch <- String(line[:len(line)-1])
		}
		bytesCollected <- true
	}()

	op(newEc)
	ClosePorts(newEc.ports)

	<-bytesCollected
	close(ch)
	<-chCollected

	return vs
}

func (cp *compiler) lambda(n *parse.Primary) ValuesOp {
	// Collect argument names
	var argNames []string
	var variadic bool
	if n.List != nil {
		// [argument list]{ chunk }
		argNames = make([]string, len(n.List.Compounds))
		for i, arg := range n.List.Compounds {
			name := mustString(cp, arg, "expect string")
			argNames[i] = name
		}
	} else {
		// { chunk }
		variadic = true
	}

	// XXX The fiddlings with cp.capture is likely wrong.
	thisScope := cp.pushScope()
	thisScope["args"] = true
	thisScope["kwargs"] = true
	for _, argName := range argNames {
		thisScope[argName] = true
	}
	op := cp.chunk(n.Chunk)
	capture := cp.capture
	cp.capture = scope{}
	cp.popScope()

	for name := range capture {
		cp.registerVariableGet(name)
	}

	return func(ec *EvalCtx) []Value {
		evCapture := make(map[string]Variable, len(capture))
		for name := range capture {
			evCapture[name] = ec.ResolveVar("", name)
		}
		return []Value{newClosure(argNames, op, evCapture, variadic)}
	}
}

func (cp *compiler) map_(n *parse.Primary) ValuesOp {
	npairs := len(n.MapPairs)
	keysOps := make([]ValuesOp, npairs)
	valuesOps := make([]ValuesOp, npairs)
	poses := make([]int, npairs)
	for i, pair := range n.MapPairs {
		keysOps[i] = cp.compound(pair.Key)
		if pair.Value == nil {
			valuesOps[i] = literalValues(Bool(true))
		} else {
			valuesOps[i] = cp.compound(n.MapPairs[i].Value)
		}
		poses[i] = n.MapPairs[i].Begin()
	}
	return func(ec *EvalCtx) []Value {
		m := make(map[Value]Value)
		for i := 0; i < npairs; i++ {
			keys := keysOps[i](ec)
			values := valuesOps[i](ec)
			if len(keys) != len(values) {
				ec.errorf(poses[i], "%d keys but %d values", len(keys), len(values))
			}
			for j, key := range keys {
				m[key] = values[j]
			}
		}
		return []Value{Map{&m}}
	}
}

func (cp *compiler) braced(n *parse.Primary) ValuesOp {
	ops := cp.compounds(n.Braced)
	// TODO: n.IsRange
	// isRange := n.IsRange
	return catValuesOps(ops)
}
