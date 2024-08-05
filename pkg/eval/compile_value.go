package eval

import (
	"errors"
	"fmt"
	"strings"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/fsutil"
	"src.elv.sh/pkg/glob"
	"src.elv.sh/pkg/parse"
)

// An operation that produces values.
type valuesOp interface {
	diag.Ranger
	exec(*Frame) ([]any, Exception)
}

var outputCaptureBufferSize = 16

// Can be mutated for testing.
var getHome = fsutil.GetHome

func (cp *compiler) compoundOp(n *parse.Compound) valuesOp {
	if len(n.Indexings) == 0 {
		return literalValues(n, "")
	}

	tilde := false
	indexings := n.Indexings

	if n.Indexings[0].Head.Type == parse.Tilde {
		// A lone ~.
		if len(n.Indexings) == 1 {
			return loneTildeOp{n.Range()}
		}
		tilde = true
		indexings = indexings[1:]
	}

	return compoundOp{n.Range(), tilde, cp.indexingOps(indexings)}
}

type loneTildeOp struct{ diag.Ranging }

func (op loneTildeOp) exec(fm *Frame) ([]any, Exception) {
	home, err := getHome("")
	if err != nil {
		return nil, fm.errorp(op, err)
	}
	return []any{home}, nil
}

func (cp *compiler) compoundOps(ns []*parse.Compound) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.compoundOp(n)
	}
	return ops
}

type compoundOp struct {
	diag.Ranging
	tilde  bool
	subops []valuesOp
}

func (op compoundOp) exec(fm *Frame) ([]any, Exception) {
	// Accumulator.
	vs, exc := op.subops[0].exec(fm)
	if exc != nil {
		return nil, exc
	}

	for _, subop := range op.subops[1:] {
		us, exc := subop.exec(fm)
		if exc != nil {
			return nil, exc
		}
		var err error
		vs, err = outerProduct(vs, us, vals.Concat)
		if err != nil {
			return nil, fm.errorp(op, err)
		}
	}
	if op.tilde {
		newvs := make([]any, len(vs))
		for i, v := range vs {
			tilded, err := doTilde(v)
			if err != nil {
				return nil, fm.errorp(op, err)
			}
			newvs[i] = tilded
		}
		vs = newvs
	}
	hasGlob := false
	for _, v := range vs {
		if _, ok := v.(globPattern); ok {
			hasGlob = true
			break
		}
	}
	if hasGlob {
		newvs := make([]any, 0, len(vs))
		for _, v := range vs {
			if gp, ok := v.(globPattern); ok {
				results, err := doGlob(fm.Context(), gp)
				if err != nil {
					return nil, fm.errorp(op, err)
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

func outerProduct(vs []any, us []any, f func(any, any) (any, error)) ([]any, error) {
	ws := make([]any, len(vs)*len(us))
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
	ErrBadglobPattern          = errors.New("bad globPattern; elvish bug")
	ErrCannotDetermineUsername = errors.New("cannot determine user name from glob pattern")
)

func doTilde(v any) (any, error) {
	switch v := v.(type) {
	case string:
		s := v
		// TODO: Make this correct on Windows.
		i := strings.Index(s, "/")
		var uname, rest string
		if i == -1 {
			uname = s
		} else {
			uname = s[:i]
			rest = s[i:]
		}
		dir, err := getHome(uname)
		if err != nil {
			return nil, err
		}
		return dir + rest, nil
	case globPattern:
		if len(v.Segments) == 0 {
			return nil, ErrBadglobPattern
		}
		switch seg := v.Segments[0].(type) {
		case glob.Literal:
			if len(v.Segments) == 1 {
				return nil, ErrBadglobPattern
			}
			_, isSlash := v.Segments[1].(glob.Slash)
			if isSlash {
				// ~username/xxx. Replace the first segment with the home
				// directory of the specified user.
				dir, err := getHome(seg.Data)
				if err != nil {
					return nil, err
				}
				v.Segments[0] = glob.Literal{Data: dir}
				return v, nil
			}
		case glob.Slash:
			dir, err := getHome("")
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
	return seqValuesOp{n.Range(), cp.compoundOps(n.Compounds)}
}

func (cp *compiler) arrayOps(ns []*parse.Array) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.arrayOp(n)
	}
	return ops
}

func (cp *compiler) indexingOp(n *parse.Indexing) valuesOp {
	if len(n.Indices) == 0 {
		return cp.primaryOp(n.Head)
	}
	return &indexingOp{n.Range(), cp.primaryOp(n.Head), cp.arrayOps(n.Indices)}
}

func (cp *compiler) indexingOps(ns []*parse.Indexing) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.indexingOp(n)
	}
	return ops
}

type indexingOp struct {
	diag.Ranging
	headOp   valuesOp
	indexOps []valuesOp
}

func (op *indexingOp) exec(fm *Frame) ([]any, Exception) {
	vs, exc := op.headOp.exec(fm)
	if exc != nil {
		return nil, exc
	}
	for _, indexOp := range op.indexOps {
		indices, exc := indexOp.exec(fm)
		if exc != nil {
			return nil, exc
		}
		newvs := make([]any, 0, len(vs)*len(indices))
		for _, v := range vs {
			for _, index := range indices {
				result, err := vals.Index(v, index)
				if err != nil {
					return nil, fm.errorp(op, err)
				}
				newvs = append(newvs, result)
			}
		}
		vs = newvs
	}
	return vs, nil
}

func (cp *compiler) primaryOp(n *parse.Primary) valuesOp {
	switch n.Type {
	case parse.Bareword, parse.SingleQuoted, parse.DoubleQuoted:
		return literalValues(n, n.Value)
	case parse.Variable:
		sigil, qname := SplitSigil(n.Value)
		ref := resolveVarRef(cp, qname, n)
		if ref == nil {
			cp.autofixUnresolvedVar(qname)
			// The variable name might be a prefix of a valid variable name.
			// Ideally, we'd want to match the variable name to all possible
			// names to check if that's actually the case, but it's a bit
			// expensive and let's call this good enough for now.
			cp.errorpfPartial(n, "variable $%s not found", parse.Quote(qname))
		}
		return &variableOp{n.Range(), sigil != "", qname, ref}
	case parse.Wildcard:
		seg, err := wildcardToSegment(parse.SourceText(n))
		if err != nil {
			cp.errorpf(n, "%s", err)
		}
		vs := []any{
			globPattern{Pattern: glob.Pattern{Segments: []glob.Segment{seg}, DirOverride: ""},
				Flags: 0, Buts: nil, TypeCb: nil}}
		return literalValues(n, vs...)
	case parse.Tilde:
		cp.errorpf(n, "compiler bug: Tilde not handled in .compound")
		return literalValues(n, "~")
	case parse.ExceptionCapture:
		return exceptionCaptureOp{n.Range(), cp.chunkOp(n.Chunk)}
	case parse.OutputCapture:
		return outputCaptureOp{n.Range(), cp.chunkOp(n.Chunk)}
	case parse.List:
		return listOp{n.Range(), cp.compoundOps(n.Elements)}
	case parse.Lambda:
		return cp.lambda(n)
	case parse.Map:
		return mapOp{n.Range(), cp.mapPairs(n.MapPairs)}
	case parse.Braced:
		return seqValuesOp{n.Range(), cp.compoundOps(n.Braced)}
	default:
		cp.errorpf(n, "bad PrimaryType; parser bug")
		return literalValues(n, parse.SourceText(n))
	}
}

func (cp *compiler) primaryOps(ns []*parse.Primary) []valuesOp {
	ops := make([]valuesOp, len(ns))
	for i, n := range ns {
		ops[i] = cp.primaryOp(n)
	}
	return ops
}

type variableOp struct {
	diag.Ranging
	explode bool
	qname   string
	ref     *varRef
}

func (op variableOp) exec(fm *Frame) ([]any, Exception) {
	variable := deref(fm, op.ref)
	if variable == nil {
		return nil, fm.errorpf(op, "variable $%s not found", parse.Quote(op.qname))
	}
	value := variable.Get()
	if op.explode {
		vs, err := vals.Collect(value)
		return vs, fm.errorp(op, err)
	}
	return []any{value}, nil
}

type listOp struct {
	diag.Ranging
	subops []valuesOp
}

func (op listOp) exec(fm *Frame) ([]any, Exception) {
	list := vals.EmptyList
	for _, subop := range op.subops {
		moreValues, exc := subop.exec(fm)
		if exc != nil {
			return nil, exc
		}
		for _, moreValue := range moreValues {
			list = list.Conj(moreValue)
		}
	}
	return []any{list}, nil
}

type exceptionCaptureOp struct {
	diag.Ranging
	subop effectOp
}

func (op exceptionCaptureOp) exec(fm *Frame) ([]any, Exception) {
	exc := op.subop.exec(fm)
	if exc == nil {
		return []any{OK}, nil
	}
	return []any{exc}, nil
}

type outputCaptureOp struct {
	diag.Ranging
	subop effectOp
}

func (op outputCaptureOp) exec(fm *Frame) ([]any, Exception) {
	outPort, collect, err := ValueCapturePort()
	if err != nil {
		return nil, fm.errorp(op, err)
	}
	exc := op.subop.exec(fm.forkWithOutput(outPort))
	return collect(), exc
}

func (cp *compiler) lambda(n *parse.Primary) valuesOp {
	// Parse signature.
	var (
		argNames      []string
		restArg       int = -1
		optNames      []string
		optDefaultOps []valuesOp
	)
	if len(n.Elements) > 0 {
		// Argument list.
		argNames = make([]string, len(n.Elements))
		seenName := make(map[string]bool)
		for i, arg := range n.Elements {
			ref := stringLiteralOrError(cp, arg, "argument name")
			sigil, qname := SplitSigil(ref)
			name, rest := SplitQName(qname)
			if rest != "" {
				cp.errorpf(arg, "argument name must be unqualified")
			}
			if name == "" {
				cp.errorpfPartial(arg, "argument name must not be empty")
			}
			if sigil == "@" {
				if restArg != -1 {
					cp.errorpf(arg, "only one argument may have @ prefix")
				}
				restArg = i
			}
			if name != "_" {
				if seenName[name] {
					cp.errorpfPartial(arg, "duplicate argument name '%s'", name)
				} else {
					seenName[name] = true
				}
			}
			argNames[i] = name
		}
	}
	if len(n.MapPairs) > 0 {
		optNames = make([]string, len(n.MapPairs))
		optDefaultOps = make([]valuesOp, len(n.MapPairs))
		for i, opt := range n.MapPairs {
			qname := stringLiteralOrError(cp, opt.Key, "option name")
			name, rest := SplitQName(qname)
			if rest != "" {
				cp.errorpf(opt.Key, "option name must be unqualified")
			}
			if name == "" {
				cp.errorpfPartial(opt.Key, "option name must not be empty")
			}
			optNames[i] = name
			if opt.Value == nil {
				cp.errorpfPartial(opt.Key, "option must have default value")
			} else {
				optDefaultOps[i] = cp.compoundOp(opt.Value)
			}
		}
	}

	local, capture := cp.pushScope()
	for _, argName := range argNames {
		local.add(argName)
	}
	for _, optName := range optNames {
		local.add(optName)
	}
	scopeSizeInit := len(local.infos)
	chunkOp := cp.chunkOp(n.Chunk)
	newLocal := local.infos[scopeSizeInit:]
	cp.popScope()

	return &lambdaOp{n.Range(), argNames, restArg, optNames, optDefaultOps, newLocal, capture, chunkOp, cp.src}
}

type lambdaOp struct {
	diag.Ranging
	argNames      []string
	restArg       int
	optNames      []string
	optDefaultOps []valuesOp
	newLocal      []staticVarInfo
	capture       *staticUpNs
	subop         effectOp
	srcMeta       parse.Source
}

func (op *lambdaOp) exec(fm *Frame) ([]any, Exception) {
	capture := &Ns{
		make([]vars.Var, len(op.capture.infos)),
		make([]staticVarInfo, len(op.capture.infos))}
	for i, info := range op.capture.infos {
		if info.local {
			capture.slots[i] = fm.local.slots[info.index]
			capture.infos[i] = fm.local.infos[info.index]
		} else {
			capture.slots[i] = fm.up.slots[info.index]
			capture.infos[i] = fm.up.infos[info.index]
		}
	}
	optDefaults := make([]any, len(op.optDefaultOps))
	for i, op := range op.optDefaultOps {
		defaultValue, err := evalForValue(fm, op, "option default value")
		if err != nil {
			return nil, err
		}
		optDefaults[i] = defaultValue
	}
	return []any{&Closure{op.argNames, op.restArg, op.optNames, optDefaults, op.srcMeta, op.Range(), op.subop, op.newLocal, capture}}, nil
}

type mapOp struct {
	diag.Ranging
	pairsOp *mapPairsOp
}

func (op mapOp) exec(fm *Frame) ([]any, Exception) {
	m := vals.EmptyMap
	exc := op.pairsOp.exec(fm, func(k, v any) Exception {
		m = m.Assoc(k, v)
		return nil
	})
	if exc != nil {
		return nil, exc
	}
	return []any{m}, nil
}

func (cp *compiler) mapPairs(pairs []*parse.MapPair) *mapPairsOp {
	npairs := len(pairs)
	keysOps := make([]valuesOp, npairs)
	valuesOps := make([]valuesOp, npairs)
	begins, ends := make([]int, npairs), make([]int, npairs)
	for i, pair := range pairs {
		keysOps[i] = cp.compoundOp(pair.Key)
		if pair.Value == nil {
			p := pair.Range().To
			valuesOps[i] = literalValues(diag.PointRanging(p), true)
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

func (op *mapPairsOp) exec(fm *Frame, f func(k, v any) Exception) Exception {
	for i := range op.keysOps {
		keys, exc := op.keysOps[i].exec(fm)
		if exc != nil {
			return exc
		}
		values, exc := op.valuesOps[i].exec(fm)
		if exc != nil {
			return exc
		}
		if len(keys) != len(values) {
			return fm.errorpf(diag.Ranging{From: op.begins[i], To: op.ends[i]},
				"%d keys but %d values", len(keys), len(values))
		}
		for j, key := range keys {
			err := f(key, values[j])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type literalValuesOp struct {
	diag.Ranging
	values []any
}

func (op literalValuesOp) exec(*Frame) ([]any, Exception) {
	return op.values, nil
}

func literalValues(r diag.Ranger, vs ...any) valuesOp {
	return literalValuesOp{r.Range(), vs}
}

type seqValuesOp struct {
	diag.Ranging
	subops []valuesOp
}

func (op seqValuesOp) exec(fm *Frame) ([]any, Exception) {
	var values []any
	for _, subop := range op.subops {
		moreValues, exc := subop.exec(fm)
		if exc != nil {
			return nil, exc
		}
		values = append(values, moreValues...)
	}
	return values, nil
}

type nopValuesOp struct{ diag.Ranging }

func (nopValuesOp) exec(fm *Frame) ([]any, Exception) { return nil, nil }

func evalForValue(fm *Frame, op valuesOp, what string) (any, Exception) {
	values, exc := op.exec(fm)
	if exc != nil {
		return nil, exc
	}
	if len(values) != 1 {
		return nil, fm.errorp(op, errs.ArityMismatch{What: what,
			ValidLow: 1, ValidHigh: 1, Actual: len(values)})
	}
	return values[0], nil
}
