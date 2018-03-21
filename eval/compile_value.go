package eval

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/glob"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/util"
)

var outputCaptureBufferSize = 16

// ValuesOp is an operation on an Frame that produce Value's.
type ValuesOp struct {
	Body       ValuesOpBody
	Begin, End int
}

// ValuesOpBody is the body of ValuesOp.
type ValuesOpBody interface {
	Invoke(*Frame) ([]interface{}, error)
}

// Exec executes a ValuesOp and produces Value's.
func (op ValuesOp) Exec(fm *Frame) ([]interface{}, error) {
	fm.begin, fm.end = op.Begin, op.End
	return op.Body.Invoke(fm)
}

func (cp *compiler) compound(n *parse.Compound) ValuesOpBody {
	if len(n.Indexings) == 0 {
		return literalStr("")
	}

	tilde := false
	indexings := n.Indexings

	if n.Indexings[0].Head.Type == parse.Tilde {
		// A lone ~.
		if len(n.Indexings) == 1 {
			return funcValuesOp(func(fm *Frame) ([]interface{}, error) {
				home, err := util.GetHome("")
				if err != nil {
					return nil, err
				}
				return []interface{}{home}, nil
			})
		}
		tilde = true
		indexings = indexings[1:]
	}

	return compoundOp{tilde, cp.indexingOps(indexings)}
}

type compoundOp struct {
	tilde  bool
	subops []ValuesOp
}

func (op compoundOp) Invoke(fm *Frame) ([]interface{}, error) {
	// Accumulator.
	vs, err := op.subops[0].Exec(fm)
	if err != nil {
		return nil, err
	}

	for _, subop := range op.subops[1:] {
		us, err := subop.Exec(fm)
		if err != nil {
			return nil, err
		}
		vs, err = outerProduct(vs, us, vals.Concat)
		if err != nil {
			return nil, err
		}
	}
	if op.tilde {
		newvs := make([]interface{}, len(vs))
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
		newvs := make([]interface{}, 0, len(vs))
		for _, v := range vs {
			if gp, ok := v.(GlobPattern); ok {
				// Logger.Printf("globbing %v", gp)
				newvs = append(newvs, doGlob(gp, fm.Interrupts())...)
			} else {
				newvs = append(newvs, v)
			}
		}
		vs = newvs
	}
	return vs, nil
}

func outerProduct(vs []interface{}, us []interface{}, f func(interface{}, interface{}) (interface{}, error)) ([]interface{}, error) {
	ws := make([]interface{}, len(vs)*len(us))
	nu := len(us)
	for i, v := range vs {
		for j, u := range us {
			var err error
			ws[i*nu+j], err = f(v, u)
			if err != nil {
				return nil, err
			}
		}
	}
	return ws, nil
}

// Errors thrown when globbing.
var (
	ErrBadGlobPattern          = errors.New("bad GlobPattern; elvish bug")
	ErrCannotDetermineUsername = errors.New("cannot determine user name from glob pattern")
)

func doTilde(v interface{}) interface{} {
	switch v := v.(type) {
	case string:
		s := v
		i := strings.Index(s, "/")
		var uname, rest string
		if i == -1 {
			uname = s
		} else {
			uname = s[:i]
			rest = s[i+1:]
		}
		dir := mustGetHome(uname)
		return path.Join(dir, rest)
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
		throw(fmt.Errorf("tilde doesn't work on value of type %s", vals.Kind(v)))
		panic("unreachable")
	}
}

func (cp *compiler) array(n *parse.Array) ValuesOpBody {
	return seqValuesOp{cp.compoundOps(n.Compounds)}
}

func (cp *compiler) indexing(n *parse.Indexing) ValuesOpBody {
	if len(n.Indicies) == 0 {
		return cp.primary(n.Head)
	}

	return &indexingOp{cp.primaryOp(n.Head), cp.arrayOps(n.Indicies)}
}

type indexingOp struct {
	headOp   ValuesOp
	indexOps []ValuesOp
}

func (op *indexingOp) Invoke(fm *Frame) ([]interface{}, error) {
	vs, err := op.headOp.Exec(fm)
	if err != nil {
		return nil, err
	}
	for _, indexOp := range op.indexOps {
		indicies, err := indexOp.Exec(fm)
		if err != nil {
			return nil, err
		}
		newvs := make([]interface{}, 0, len(vs)*len(indicies))
		for _, v := range vs {
			for _, index := range indicies {
				result, err := vals.Index(v, index)
				if err != nil {
					return nil, err
				}
				newvs = append(newvs, result)
			}
		}
		vs = newvs
	}
	return vs, nil
}

func (cp *compiler) primary(n *parse.Primary) ValuesOpBody {
	switch n.Type {
	case parse.Bareword, parse.SingleQuoted, parse.DoubleQuoted:
		return literalStr(n.Value)
	case parse.Variable:
		explode, ns, name := ParseVariableRef(n.Value)
		if !cp.registerVariableGet(ns, name) {
			cp.errorf("variable $%s not found", n.Value)
		}
		return &variableOp{explode, ns, name}
	case parse.Wildcard:
		seg, err := wildcardToSegment(n.SourceText())
		if err != nil {
			cp.errorf("%s", err)
		}
		vs := []interface{}{
			GlobPattern{glob.Pattern{[]glob.Segment{seg}, ""}, 0, nil}}
		return literalValues(vs...)
	case parse.Tilde:
		cp.errorf("compiler bug: Tilde not handled in .compound")
		return literalStr("~")
	case parse.ExceptionCapture:
		return exceptionCaptureOp{cp.chunkOp(n.Chunk)}
	case parse.OutputCapture:
		return outputCaptureOp{cp.chunkOp(n.Chunk)}
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

type variableOp struct {
	explode bool
	ns      string
	name    string
}

func (op variableOp) Invoke(fm *Frame) ([]interface{}, error) {
	variable := fm.ResolveVar(op.ns, op.name)
	if variable == nil {
		return nil, fmt.Errorf("variable $%s:%s not found", op.ns, op.name)
	}
	value := variable.Get()
	if op.explode {
		return vals.Collect(value)
	}
	return []interface{}{value}, nil
}

func (cp *compiler) list(n *parse.Primary) ValuesOpBody {
	return listOp{cp.compoundOps(n.Elements)}
}

type listOp struct{ subops []ValuesOp }

func (op listOp) Invoke(fm *Frame) ([]interface{}, error) {
	list := vals.EmptyList
	for _, subop := range op.subops {
		moreValues, err := subop.Exec(fm)
		if err != nil {
			return nil, err
		}
		for _, moreValue := range moreValues {
			list = list.Cons(moreValue)
		}
	}
	return []interface{}{list}, nil
}

type exceptionCaptureOp struct{ subop Op }

func (op exceptionCaptureOp) Invoke(fm *Frame) ([]interface{}, error) {
	err := fm.Eval(op.subop)
	if err == nil {
		return []interface{}{OK}, nil
	}
	return []interface{}{err.(*Exception)}, nil
}

type outputCaptureOp struct{ subop Op }

func (op outputCaptureOp) Invoke(fm *Frame) ([]interface{}, error) {
	return pcaptureOutput(fm, op.subop)
}

func pcaptureOutput(fm *Frame, op Op) ([]interface{}, error) {
	vs := []interface{}{}
	var m sync.Mutex
	valueCb := func(ch <-chan interface{}) {
		for v := range ch {
			m.Lock()
			vs = append(vs, v)
			m.Unlock()
		}
	}
	bytesCb := func(r *os.File) {
		buffered := bufio.NewReader(r)
		for {
			line, err := buffered.ReadString('\n')
			if line != "" {
				v := strings.TrimSuffix(line, "\n")
				m.Lock()
				vs = append(vs, v)
				m.Unlock()
			}
			if err != nil {
				if err != io.EOF {
					logger.Println("error on reading:", err)
				}
				break
			}
		}
	}

	err := pcaptureOutputInner(fm, op, valueCb, bytesCb)
	return vs, err
}

func pcaptureOutputInner(fm *Frame, op Op, valuesCb func(<-chan interface{}), bytesCb func(*os.File)) error {

	newFm := fm.fork("[output capture]")

	ch := make(chan interface{}, outputCaptureBufferSize)
	pipeRead, pipeWrite, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("failed to create pipe: %v", err)
	}
	newFm.ports[1] = &Port{
		Chan: ch, CloseChan: true,
		File: pipeWrite, CloseFile: true,
	}

	bytesCollected := make(chan struct{})
	chCollected := make(chan struct{})

	go func() {
		valuesCb(ch)
		close(chCollected)
	}()
	go func() {
		bytesCb(pipeRead)
		pipeRead.Close()
		close(bytesCollected)
	}()

	err = newFm.Eval(op)

	newFm.Close()
	<-bytesCollected
	<-chCollected

	return err
}

func (cp *compiler) lambda(n *parse.Primary) ValuesOpBody {
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
			explode, ns, name := ParseVariableRef(qname)
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
			_, ns, name := ParseVariableRef(qname)
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
		thisScope.set(argName)
	}
	if restArgName != "" {
		thisScope.set(restArgName)
	}
	for _, optName := range optNames {
		thisScope.set(optName)
	}

	subop := cp.chunkOp(n.Chunk)

	// XXX The fiddlings with cp.capture is error-prone.
	capture := cp.capture
	cp.capture = make(staticNs)
	cp.popScope()

	for name := range capture {
		cp.registerVariableGetQname(name)
	}

	return &lambdaOp{argNames, restArgName, optNames, optDefaultOps, capture, subop, cp.srcMeta, n.Begin(), n.End()}
}

type lambdaOp struct {
	argNames      []string
	restArgName   string
	optNames      []string
	optDefaultOps []ValuesOp
	capture       staticNs
	subop         Op
	srcMeta       *Source
	defBegin      int
	defEnd        int
}

func (op *lambdaOp) Invoke(fm *Frame) ([]interface{}, error) {
	evCapture := make(Ns)
	for name := range op.capture {
		evCapture[name] = fm.ResolveVar("", name)
	}
	optDefaults := make([]interface{}, len(op.optDefaultOps))
	for i, op := range op.optDefaultOps {
		defaultValue := fm.ExecAndUnwrap("option default value", op).One().Any()
		optDefaults[i] = defaultValue
	}
	// XXX(xiaq): Capture uses.
	return []interface{}{&Closure{op.argNames, op.restArgName, op.optNames, optDefaults, op.subop, evCapture, op.srcMeta, op.defBegin, op.defEnd}}, nil
}

func (cp *compiler) map_(n *parse.Primary) ValuesOpBody {
	return cp.mapPairs(n.MapPairs)
}

func (cp *compiler) mapPairs(pairs []*parse.MapPair) ValuesOpBody {
	npairs := len(pairs)
	keysOps := make([]ValuesOp, npairs)
	valuesOps := make([]ValuesOp, npairs)
	begins, ends := make([]int, npairs), make([]int, npairs)
	for i, pair := range pairs {
		keysOps[i] = cp.compoundOp(pair.Key)
		if pair.Value == nil {
			p := pair.End()
			valuesOps[i] = ValuesOp{literalValues(true), p, p}
		} else {
			valuesOps[i] = cp.compoundOp(pairs[i].Value)
		}
		begins[i], ends[i] = pair.Begin(), pair.End()
	}
	return &mapPairsOp{keysOps, valuesOps, begins, ends}
}

type mapPairsOp struct {
	keysOps   []ValuesOp
	valuesOps []ValuesOp
	begins    []int
	ends      []int
}

func (op *mapPairsOp) Invoke(fm *Frame) ([]interface{}, error) {
	m := vals.EmptyMap
	for i := range op.keysOps {
		keys, err := op.keysOps[i].Exec(fm)
		if err != nil {
			return nil, err
		}
		values, err := op.valuesOps[i].Exec(fm)
		if err != nil {
			return nil, err
		}
		if len(keys) != len(values) {
			fm.errorpf(op.begins[i], op.ends[i],
				"%d keys but %d values", len(keys), len(values))
		}
		for j, key := range keys {
			m = m.Assoc(key, values[j])
		}
	}
	return []interface{}{m}, nil
}

func (cp *compiler) braced(n *parse.Primary) ValuesOpBody {
	ops := cp.compoundOps(n.Braced)
	// TODO: n.IsRange
	// isRange := n.IsRange
	return seqValuesOp{ops}
}

type literalValuesOp struct{ values []interface{} }

func (op literalValuesOp) Invoke(*Frame) ([]interface{}, error) {
	return op.values, nil
}

func literalValues(v ...interface{}) ValuesOpBody {
	return literalValuesOp{v}
}

func literalStr(text string) ValuesOpBody {
	return literalValues(text)
}

type seqValuesOp struct{ subops []ValuesOp }

func (op seqValuesOp) Invoke(fm *Frame) ([]interface{}, error) {
	var values []interface{}
	for _, subop := range op.subops {
		moreValues, err := subop.Exec(fm)
		if err != nil {
			return nil, err
		}
		values = append(values, moreValues...)
	}
	return values, nil
}

type funcValuesOp func(*Frame) ([]interface{}, error)

func (op funcValuesOp) Invoke(fm *Frame) ([]interface{}, error) { return op(fm) }
