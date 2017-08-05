package eval

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/elves/elvish/glob"
	"github.com/elves/elvish/parse"
)

var outputCaptureBufferSize = 16

// ValuesOp is an operation on an EvalCtx that produce Value's.
type ValuesOp struct {
	Func       ValuesOpFunc
	Begin, End int
}

// ValuesOpFunc is the body of ValuesOp.
type ValuesOpFunc func(*EvalCtx) []Value

// Exec executes a ValuesOp and produces Value's.
func (op ValuesOp) Exec(ec *EvalCtx) []Value {
	ec.begin, ec.end = op.Begin, op.End
	return op.Func(ec)
}

func (cp *compiler) compound(n *parse.Compound) ValuesOpFunc {
	if len(n.Indexings) == 0 {
		return literalStr("")
	}

	tilde := false
	indexings := n.Indexings

	if n.Indexings[0].Head.Type == parse.Tilde {
		// A lone ~.
		if len(n.Indexings) == 1 {
			return func(ec *EvalCtx) []Value {
				return []Value{String(mustGetHome(""))}
			}
		}
		tilde = true
		indexings = indexings[1:]
	}

	ops := cp.indexingOps(indexings)

	return func(ec *EvalCtx) []Value {
		// Accumulator.
		vs := ops[0].Exec(ec)

		// Logger.Printf("concatenating %v with %d more", vs, len(ops)-1)

		for _, op := range ops[1:] {
			us := op.Exec(ec)
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
					newvs = append(newvs, doGlob(gp, ec.Interrupts())...)
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
			return GlobPattern{glob.Pattern{segs, ""}, rhs.Flags, rhs.Buts}
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
			lhs.Flags |= rhs.Flags
			lhs.Buts = append(lhs.Buts, rhs.Buts...)
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

// Errors thrown when globbing.
var (
	ErrBadGlobPattern          = errors.New("bad GlobPattern; elvish bug")
	ErrCannotDetermineUsername = errors.New("cannot determine user name from glob pattern")
)

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
		if len(v.Segments) == 0 {
			throw(ErrBadGlobPattern)
		}
		switch seg := v.Segments[0].(type) {
		case glob.Literal:
			s := seg.Data
			// Find / in the first segment to determine the username.
			i := strings.Index(s, "/")
			if i == -1 {
				throw(ErrCannotDetermineUsername)
			}
			uname := s[:i]
			dir := mustGetHome(uname)
			// Replace ~uname in first segment with the found path.
			v.Segments[0] = glob.Literal{dir + s[i:]}
		case glob.Slash:
			v.DirOverride = mustGetHome("")
		default:
			throw(ErrCannotDetermineUsername)
		}
		return v
	default:
		throw(fmt.Errorf("tilde doesn't work on value of type %s", v.Kind()))
		panic("unreachable")
	}
}

func (cp *compiler) array(n *parse.Array) ValuesOpFunc {
	return catValuesOps(cp.compoundOps(n.Compounds))
}

func catValuesOps(ops []ValuesOp) ValuesOpFunc {
	return func(ec *EvalCtx) []Value {
		// Use number of compound expressions as an estimation of the number
		// of values
		vs := make([]Value, 0, len(ops))
		for _, op := range ops {
			us := op.Exec(ec)
			vs = append(vs, us...)
		}
		return vs
	}
}

func (cp *compiler) indexing(n *parse.Indexing) ValuesOpFunc {
	if len(n.Indicies) == 0 {
		return cp.primary(n.Head)
	}

	headOp := cp.primaryOp(n.Head)
	indexOps := cp.arrayOps(n.Indicies)

	return func(ec *EvalCtx) []Value {
		vs := headOp.Exec(ec)
		for _, indexOp := range indexOps {
			indicies := indexOp.Exec(ec)
			newvs := make([]Value, 0, len(vs)*len(indicies))
			for _, v := range vs {
				newvs = append(newvs, mustIndexer(v, ec).Index(indicies)...)
			}
			vs = newvs
		}
		return vs
	}
}

func literalValues(v ...Value) ValuesOpFunc {
	return func(e *EvalCtx) []Value {
		return v
	}
}

func literalStr(text string) ValuesOpFunc {
	return literalValues(String(text))
}

func variable(qname string) ValuesOpFunc {
	explode, ns, name := ParseAndFixVariable(qname)
	return func(ec *EvalCtx) []Value {
		variable := ec.ResolveVar(ns, name)
		if variable == nil {
			throwf("variable $%s not found", qname)
		}
		value := variable.Get()
		if explode {
			iterator, ok := value.(Iterable)
			if !ok {
				// Use qname[1:] to skip the leading "@"
				throwf("variable $%s (kind %s) cannot be exploded", qname[1:], value.Kind())
			}
			return collectFromIterable(iterator)
		}
		return []Value{value}
	}
}

func (cp *compiler) primary(n *parse.Primary) ValuesOpFunc {
	switch n.Type {
	case parse.Bareword, parse.SingleQuoted, parse.DoubleQuoted:
		return literalStr(n.Value)
	case parse.Variable:
		qname := n.Value
		if !cp.registerVariableGet(qname) {
			cp.errorf("variable $%s not found", n.Value)
		}
		return variable(qname)
	case parse.Wildcard:
		seg, err := wildcardToSegment(n.SourceText())
		if err != nil {
			cp.errorf("%s", err)
		}
		vs := []Value{
			GlobPattern{glob.Pattern{[]glob.Segment{seg}, ""}, 0, nil}}
		return func(ec *EvalCtx) []Value {
			return vs
		}
	case parse.Tilde:
		cp.errorf("compiler bug: Tilde not handled in .compound")
		return literalStr("~")
	case parse.ExceptionCapture:
		return cp.exceptionCapture(n.Chunk)
	case parse.OutputCapture:
		return cp.outputCapture(n)
	case parse.List:
		return cp.list(n)
	case parse.Lambda:
		return cp.lambda(n)
	case parse.Map:
		return cp.map_(n)
	case parse.Braced:
		return cp.braced(n)
	default:
		cp.errorf("bad PrimaryType; parser bug")
		return literalStr(n.SourceText())
	}
}

func (cp *compiler) list(n *parse.Primary) ValuesOpFunc {
	// TODO(xiaq): Use Vector.Cons to build the list, instead of building a
	// slice and converting to Vector.
	op := catValuesOps(cp.compoundOps(n.Elements))
	return func(ec *EvalCtx) []Value {
		return []Value{NewList(op(ec)...)}
	}
}

func (cp *compiler) exceptionCapture(n *parse.Chunk) ValuesOpFunc {
	op := cp.chunkOp(n)
	return func(ec *EvalCtx) []Value {
		err := ec.PEval(op)
		if err == nil {
			return []Value{OK}
		}
		return []Value{err.(*Exception)}
	}
}

func (cp *compiler) outputCapture(n *parse.Primary) ValuesOpFunc {
	op := cp.chunkOp(n.Chunk)
	return func(ec *EvalCtx) []Value {
		return captureOutput(ec, op)
	}
}

func captureOutput(ec *EvalCtx, op Op) []Value {
	vs, err := pcaptureOutput(ec, op)
	maybeThrow(err)
	return vs
}

func pcaptureOutput(ec *EvalCtx, op Op) ([]Value, error) {
	vs := []Value{}
	newEc := ec.fork(fmt.Sprintf("output capture %v", op))

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
			if line != "" {
				ch <- String(strings.TrimSuffix(line, "\n"))
			}
			if err != nil {
				if err != io.EOF {
					logger.Println("error on reading:", err)
				}
				break
			}
		}
		bytesCollected <- true
	}()

	err = newEc.PEval(op)
	ClosePorts(newEc.ports)

	<-bytesCollected
	pipeRead.Close()

	close(ch)
	<-chCollected

	return vs, err
}

func (cp *compiler) lambda(n *parse.Primary) ValuesOpFunc {
	// Parse signature.
	var (
		argNames      []string
		restArgName   string
		optNames      []string
		optDefaultOps []ValuesOp
	)
	if len(n.Elements) > 0 {
		// Argument list.
		argNames = make([]string, len(n.Elements))
		for i, arg := range n.Elements {
			qname := mustString(cp, arg, "argument name must be literal string")
			explode, ns, name := ParseAndFixVariable(qname)
			if ns != "" {
				cp.errorpf(arg.Begin(), arg.End(), "argument name must be unqualified")
			}
			if name == "" {
				cp.errorpf(arg.Begin(), arg.End(), "argument name must not be empty")
			}
			if explode {
				if i != len(n.Elements)-1 {
					cp.errorpf(arg.Begin(), arg.End(), "only the last argument may have @")
				}
				restArgName = name
				argNames = argNames[:i]
			} else {
				argNames[i] = name
			}
		}
	}
	if len(n.MapPairs) > 0 {
		optNames = make([]string, len(n.MapPairs))
		optDefaultOps = make([]ValuesOp, len(n.MapPairs))
		for i, opt := range n.MapPairs {
			qname := mustString(cp, opt.Key, "option name must be literal string")
			_, ns, name := ParseAndFixVariable(qname)
			if ns != "" {
				cp.errorpf(opt.Key.Begin(), opt.Key.End(), "option name must be unqualified")
			}
			if name == "" {
				cp.errorpf(opt.Key.Begin(), opt.Key.End(), "option name must not be empty")
			}
			optNames[i] = name
			if opt.Value == nil {
				cp.errorpf(opt.End(), opt.End(), "option must have default value")
			} else {
				optDefaultOps[i] = cp.compoundOp(opt.Value)
			}
		}
	}

	thisScope := cp.pushScope()
	for _, argName := range argNames {
		thisScope[argName] = true
	}
	if restArgName != "" {
		thisScope[restArgName] = true
	}
	for _, optName := range optNames {
		thisScope[optName] = true
	}

	thisScope["opts"] = true
	op := cp.chunkOp(n.Chunk)

	// XXX The fiddlings with cp.capture is error-prone.
	capture := cp.capture
	cp.capture = scope{}
	cp.popScope()

	for name := range capture {
		cp.registerVariableGet(name)
	}

	name, text := cp.name, cp.text

	return func(ec *EvalCtx) []Value {
		evCapture := make(map[string]Variable, len(capture))
		for name := range capture {
			evCapture[name] = ec.ResolveVar("", name)
		}
		optDefaults := make([]Value, len(optDefaultOps))
		for i, op := range optDefaultOps {
			values := op.Exec(ec)
			if len(values) != 1 {
				ec.errorpf(op.Begin, op.End, "option default value must evalute to a single value")
			}
			optDefaults[i] = values[0]
		}
		return []Value{&Closure{argNames, restArgName, optNames, optDefaults, op, evCapture, name, text}}
	}
}

func (cp *compiler) map_(n *parse.Primary) ValuesOpFunc {
	return cp.mapPairs(n.MapPairs)
}

func (cp *compiler) mapPairs(pairs []*parse.MapPair) ValuesOpFunc {
	npairs := len(pairs)
	keysOps := make([]ValuesOp, npairs)
	valuesOps := make([]ValuesOp, npairs)
	begins, ends := make([]int, npairs), make([]int, npairs)
	for i, pair := range pairs {
		keysOps[i] = cp.compoundOp(pair.Key)
		if pair.Value == nil {
			p := pair.End()
			valuesOps[i] = ValuesOp{literalValues(Bool(true)), p, p}
		} else {
			valuesOps[i] = cp.compoundOp(pairs[i].Value)
		}
		begins[i], ends[i] = pair.Begin(), pair.End()
	}
	return func(ec *EvalCtx) []Value {
		m := make(map[Value]Value)
		for i := 0; i < npairs; i++ {
			keys := keysOps[i].Exec(ec)
			values := valuesOps[i].Exec(ec)
			if len(keys) != len(values) {
				ec.errorpf(begins[i], ends[i],
					"%d keys but %d values", len(keys), len(values))
			}
			for j, key := range keys {
				m[key] = values[j]
			}
		}
		return []Value{Map{&m}}
	}
}

func (cp *compiler) braced(n *parse.Primary) ValuesOpFunc {
	ops := cp.compoundOps(n.Braced)
	// TODO: n.IsRange
	// isRange := n.IsRange
	return catValuesOps(ops)
}
