package eval

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/elves/elvish/pkg/diag"
	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/glob"
	"github.com/elves/elvish/pkg/parse"
	"github.com/elves/elvish/pkg/util"
)

var outputCaptureBufferSize = 16

func (cp *compiler) compoundOp(n *parse.Compound) valuesOp {
	if len(n.Indexings) == 0 {
		return makeValuesOp(n, literalStr(""))
	}

	tilde := false
	indexings := n.Indexings

	if n.Indexings[0].Head.Type == parse.Tilde {
		// A lone ~.
		if len(n.Indexings) == 1 {
			return makeValuesOp(n, funcValuesOp(func(fm *Frame) ([]interface{}, error) {
				home, err := util.GetHome("")
				if err != nil {
					return nil, err
				}
				return []interface{}{home}, nil
			}))
		}
		tilde = true
		indexings = indexings[1:]
	}

	return makeValuesOp(n, compoundOp{tilde, cp.indexingOps(indexings)})
}

func (cp *compiler) compoundOps(ns []*parse.Compound) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.compoundOp(n)
	}
	return ops
}

type compoundOp struct {
	tilde  bool
	subops []valuesOp
}

func (op compoundOp) invoke(fm *Frame) ([]interface{}, error) {
	// Accumulator.
	vs, err := op.subops[0].exec(fm)
	if err != nil {
		return nil, err
	}

	for _, subop := range op.subops[1:] {
		us, err := subop.exec(fm)
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
			tilded, err := doTilde(v)
			if err != nil {
				return nil, err
			}
			newvs[i] = tilded
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
				results, err := doGlob(gp, fm.Interrupts())
				if err != nil {
					return nil, err
				}
				newvs = append(newvs, results...)
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

func doTilde(v interface{}) (interface{}, error) {
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
		dir, err := util.GetHome(uname)
		if err != nil {
			return nil, err
		}
		// We do not use path.Join, as it removes trailing slashes.
		//
		// TODO(xiaq): Make this correct on Windows.
		return dir + "/" + rest, nil
	case GlobPattern:
		if len(v.Segments) == 0 {
			return nil, ErrBadGlobPattern
		}
		switch seg := v.Segments[0].(type) {
		case glob.Literal:
			if len(v.Segments) == 1 {
				return nil, ErrBadGlobPattern
			}
			_, isSlash := v.Segments[1].(glob.Slash)
			if isSlash {
				// ~username or ~username/xxx. Replace the first segment with
				// the home directory of the specified user.
				dir, err := util.GetHome(seg.Data)
				if err != nil {
					return nil, err
				}
				v.Segments[0] = glob.Literal{Data: dir}
				return v, nil
			}
		case glob.Slash:
			dir, err := util.GetHome("")
			if err != nil {
				return nil, err
			}
			v.DirOverride = dir
			return v, nil
		}
		return nil, ErrCannotDetermineUsername
	default:
		return nil, fmt.Errorf("tilde doesn't work on value of type %s", vals.Kind(v))
	}
}

func (cp *compiler) arrayOp(n *parse.Array) valuesOp {
	return makeValuesOp(n, seqValuesOp{cp.compoundOps(n.Compounds)})
}

func (cp *compiler) arrayOps(ns []*parse.Array) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.arrayOp(n)
	}
	return ops
}

func (cp *compiler) indexingOp(n *parse.Indexing) valuesOp {
	if len(n.Indicies) == 0 {
		return cp.primaryOp(n.Head)
	}
	return makeValuesOp(n,
		&indexingOp{cp.primaryOp(n.Head), cp.arrayOps(n.Indicies)})
}

func (cp *compiler) indexingOps(ns []*parse.Indexing) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.indexingOp(n)
	}
	return ops
}

type indexingOp struct {
	headOp   valuesOp
	indexOps []valuesOp
}

func (op *indexingOp) invoke(fm *Frame) ([]interface{}, error) {
	vs, err := op.headOp.exec(fm)
	if err != nil {
		return nil, err
	}
	for _, indexOp := range op.indexOps {
		indicies, err := indexOp.exec(fm)
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

func (cp *compiler) primaryOp(n *parse.Primary) valuesOp {
	var body valuesOpBody
	switch n.Type {
	case parse.Bareword, parse.SingleQuoted, parse.DoubleQuoted:
		body = literalStr(n.Value)
	case parse.Variable:
		sigil, qname := SplitVariableRef(n.Value)
		if !cp.registerVariableGet(qname) {
			cp.errorpf(n, "variable $%s not found", qname)
		}
		body = &variableOp{sigil != "", qname}
	case parse.Wildcard:
		seg, err := wildcardToSegment(n.SourceText())
		if err != nil {
			cp.errorpf(n, "%s", err)
		}
		vs := []interface{}{
			GlobPattern{glob.Pattern{[]glob.Segment{seg}, ""}, 0, nil}}
		body = literalValues(vs...)
	case parse.Tilde:
		cp.errorpf(n, "compiler bug: Tilde not handled in .compound")
		body = literalStr("~")
	case parse.ExceptionCapture:
		body = exceptionCaptureOp{cp.chunkOp(n.Chunk)}
	case parse.OutputCapture:
		body = outputCaptureOp{cp.chunkOp(n.Chunk)}
	case parse.List:
		body = cp.list(n)
	case parse.Lambda:
		body = cp.lambda(n)
	case parse.Map:
		body = cp.map_(n)
	case parse.Braced:
		body = cp.braced(n)
	default:
		cp.errorpf(n, "bad PrimaryType; parser bug")
		body = literalStr(n.SourceText())
	}
	return makeValuesOp(n, body)
}

func (cp *compiler) primaryOps(ns []*parse.Primary) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.primaryOp(n)
	}
	return ops
}

type variableOp struct {
	explode bool
	qname   string
}

func (op variableOp) invoke(fm *Frame) ([]interface{}, error) {
	variable := fm.ResolveVar(op.qname)
	if variable == nil {
		return nil, fmt.Errorf("variable $%s not found", op.qname)
	}
	value := variable.Get()
	if op.explode {
		return vals.Collect(value)
	}
	return []interface{}{value}, nil
}

func (cp *compiler) list(n *parse.Primary) valuesOpBody {
	return listOp{cp.compoundOps(n.Elements)}
}

type listOp struct{ subops []valuesOp }

func (op listOp) invoke(fm *Frame) ([]interface{}, error) {
	list := vals.EmptyList
	for _, subop := range op.subops {
		moreValues, err := subop.exec(fm)
		if err != nil {
			return nil, err
		}
		for _, moreValue := range moreValues {
			list = list.Cons(moreValue)
		}
	}
	return []interface{}{list}, nil
}

type exceptionCaptureOp struct{ subop effectOp }

func (op exceptionCaptureOp) invoke(fm *Frame) ([]interface{}, error) {
	err := op.subop.exec(fm)
	if err == nil {
		return []interface{}{OK}, nil
	}
	return []interface{}{err.(*Exception)}, nil
}

type outputCaptureOp struct{ subop effectOp }

func (op outputCaptureOp) invoke(fm *Frame) ([]interface{}, error) {
	return captureOutput(fm, op.subop.exec)
}

func captureOutput(fm *Frame, f func(*Frame) error) ([]interface{}, error) {
	vs := []interface{}{}
	var m sync.Mutex
	err := pipeOutput(fm, f,
		func(ch <-chan interface{}) {
			for v := range ch {
				m.Lock()
				vs = append(vs, v)
				m.Unlock()
			}
		},
		func(r *os.File) {
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
		})
	return vs, err
}

func pipeOutput(fm *Frame, f func(*Frame) error, valuesCb func(<-chan interface{}), bytesCb func(*os.File)) error {
	newFm := fm.fork("[output capture]")

	ch := make(chan interface{}, outputCaptureBufferSize)
	r, w, err := os.Pipe()
	if err != nil {
		return err
	}
	newFm.ports[1] = &Port{
		Chan: ch, CloseChan: true,
		File: w, CloseFile: true,
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		valuesCb(ch)
	}()
	go func() {
		defer wg.Done()
		defer r.Close()
		bytesCb(r)
	}()

	err = f(newFm)
	newFm.Close()
	wg.Wait()
	return err
}

func (cp *compiler) lambda(n *parse.Primary) valuesOpBody {
	// Parse signature.
	var (
		argNames      []string
		restArgName   string
		optNames      []string
		optDefaultOps []valuesOp
	)
	if len(n.Elements) > 0 {
		// Argument list.
		argNames = make([]string, len(n.Elements))
		for i, arg := range n.Elements {
			ref := mustString(cp, arg, "argument name must be literal string")
			sigil, qname := SplitVariableRef(ref)
			explode := sigil != ""
			ns, name := SplitQNameNs(qname)
			if ns != "" {
				cp.errorpf(arg, "argument name must be unqualified")
			}
			if name == "" {
				cp.errorpf(arg, "argument name must not be empty")
			}
			if explode {
				if i != len(n.Elements)-1 {
					cp.errorpf(arg, "only the last argument may have @")
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
		optDefaultOps = make([]valuesOp, len(n.MapPairs))
		for i, opt := range n.MapPairs {
			qname := mustString(cp, opt.Key, "option name must be literal string")
			ns, name := SplitQNameNs(qname)
			if ns != "" {
				cp.errorpf(opt.Key, "option name must be unqualified")
			}
			if name == "" {
				cp.errorpf(opt.Key, "option name must not be empty")
			}
			optNames[i] = name
			if opt.Value == nil {
				cp.errorpf(opt.Key, "option must have default value")
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
		cp.registerVariableGet(name)
	}

	return &lambdaOp{argNames, restArgName, optNames, optDefaultOps, capture, subop, cp.srcMeta, n.Range().From, n.Range().To}
}

type lambdaOp struct {
	argNames      []string
	restArgName   string
	optNames      []string
	optDefaultOps []valuesOp
	capture       staticNs
	subop         effectOp
	srcMeta       *Source
	defBegin      int
	defEnd        int
}

func (op *lambdaOp) invoke(fm *Frame) ([]interface{}, error) {
	evCapture := make(Ns)
	for name := range op.capture {
		evCapture[name] = fm.ResolveVar(":" + name)
	}
	optDefaults := make([]interface{}, len(op.optDefaultOps))
	for i, op := range op.optDefaultOps {
		defaultValue, err := fm.ExecAndUnwrap("option default value", op).One().Any()
		if err != nil {
			return nil, err
		}
		optDefaults[i] = defaultValue
	}
	return []interface{}{&Closure{op.argNames, op.restArgName, op.optNames, optDefaults, op.subop, evCapture, op.srcMeta, op.defBegin, op.defEnd}}, nil
}

func (cp *compiler) map_(n *parse.Primary) valuesOpBody {
	return cp.mapPairs(n.MapPairs)
}

func (cp *compiler) mapPairs(pairs []*parse.MapPair) valuesOpBody {
	npairs := len(pairs)
	keysOps := make([]valuesOp, npairs)
	valuesOps := make([]valuesOp, npairs)
	begins, ends := make([]int, npairs), make([]int, npairs)
	for i, pair := range pairs {
		keysOps[i] = cp.compoundOp(pair.Key)
		if pair.Value == nil {
			p := pair.Range().To
			valuesOps[i] = valuesOp{literalValues(true), diag.Ranging{From: p, To: p}}
		} else {
			valuesOps[i] = cp.compoundOp(pairs[i].Value)
		}
		begins[i], ends[i] = pair.Range().From, pair.Range().To
	}
	return &mapPairsOp{keysOps, valuesOps, begins, ends}
}

type mapPairsOp struct {
	keysOps   []valuesOp
	valuesOps []valuesOp
	begins    []int
	ends      []int
}

func (op *mapPairsOp) invoke(fm *Frame) ([]interface{}, error) {
	m := vals.EmptyMap
	for i := range op.keysOps {
		keys, err := op.keysOps[i].exec(fm)
		if err != nil {
			return nil, err
		}
		values, err := op.valuesOps[i].exec(fm)
		if err != nil {
			return nil, err
		}
		if len(keys) != len(values) {
			return nil, fm.errorpf(diag.Ranging{From: op.begins[i], To: op.ends[i]},
				"%d keys but %d values", len(keys), len(values))
		}
		for j, key := range keys {
			m = m.Assoc(key, values[j])
		}
	}
	return []interface{}{m}, nil
}

func (cp *compiler) braced(n *parse.Primary) valuesOpBody {
	ops := cp.compoundOps(n.Braced)
	// TODO: n.IsRange
	// isRange := n.IsRange
	return seqValuesOp{ops}
}

type literalValuesOp struct{ values []interface{} }

func (op literalValuesOp) invoke(*Frame) ([]interface{}, error) {
	return op.values, nil
}

func literalValues(v ...interface{}) valuesOpBody {
	return literalValuesOp{v}
}

func literalStr(text string) valuesOpBody {
	return literalValues(text)
}

type seqValuesOp struct{ subops []valuesOp }

func (op seqValuesOp) invoke(fm *Frame) ([]interface{}, error) {
	var values []interface{}
	for _, subop := range op.subops {
		moreValues, err := subop.exec(fm)
		if err != nil {
			return nil, err
		}
		values = append(values, moreValues...)
	}
	return values, nil
}

type funcValuesOp func(*Frame) ([]interface{}, error)

func (op funcValuesOp) invoke(fm *Frame) ([]interface{}, error) { return op(fm) }
