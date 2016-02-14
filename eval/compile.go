package eval

//go:generate ./boilerplate.py

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/elves/elvish/errutil"
	"github.com/elves/elvish/glob"
	"github.com/elves/elvish/osutil"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/sys"
)

const (
	// InterruptDeadline is the amount of time elvish waits for foreground
	// tasks to finish after receiving a SIGINT. If a task didn't actually exit
	// in time, its exit status takes the special "still running" value.
	InterruptDeadline = 50 * time.Millisecond
)

var (
	ErrStillRunning = errors.New("still running")
)

// Whether elvish should attempt to put itself in foreground after each pipeline
// execution.
var PutInForeground = true

var outputCaptureBufferSize = 16

type scope map[string]bool

type (
	// Op is a compiled operation.
	Op func(*EvalCtx)
	// ValuesOp is a compiled Value-generating operation.
	ValuesOp func(*EvalCtx) []Value
	// VariableOp is a compiled Variable-generating operation.
	VariableOp func(*EvalCtx) Variable
)

// compiler maintains the set of states needed when compiling a single source
// file.
type compiler struct {
	// Used in error messages.
	name, source string
	// Lexical scopes.
	scopes []scope
	// Variables captured from outer scopes.
	capture scope
	// Stored error.
	error error
}

func (cp *compiler) thisScope() scope {
	return cp.scopes[len(cp.scopes)-1]
}

func (cp *compiler) errorf(p int, format string, args ...interface{}) {
	throw(errutil.NewContextualError(cp.name, "syntax error", cp.source, p, format, args...))
}

func compile(name, source string, sc scope, n *parse.Chunk) (op Op, err error) {
	cp := &compiler{name, source, []scope{sc}, scope{}, nil}
	defer errutil.Catch(&err)
	return cp.chunk(n), nil
}

func (cp *compiler) chunk(n *parse.Chunk) Op {
	ops := cp.pipelines(n.Pipelines)

	return func(ec *EvalCtx) {
		for _, op := range ops {
			op(ec)
		}
	}
}

const pipelineChanBufferSize = 32

func (cp *compiler) pipeline(n *parse.Pipeline) Op {
	ops := cp.forms(n.Forms)
	p := n.Begin()

	return func(ec *EvalCtx) {
		var nextIn *Port

		errorChans := make([]chan Error, len(ops))

		// For each form, create a dedicated evalCtx and run asynchronously
		for i, op := range ops {
			newEc := ec.fork(fmt.Sprintf("form op %v", op))
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
				newEc.ports[1] = &Port{
					File: writer, Chan: ch, CloseFile: true, CloseChan: true}
				nextIn = &Port{
					File: reader, Chan: ch, CloseFile: true, CloseChan: false}
			}
			thisOp := op
			errorChans[i] = make(chan Error)
			thisErrorChan := errorChans[i]
			go func() {
				err := newEc.PEval(thisOp)
				// Logger.Printf("closing ports of %s", newEc.context)
				ClosePorts(newEc.ports)
				thisErrorChan <- Error{err}
			}()
		}

		intCh := make(chan os.Signal)
		signal.Notify(intCh, syscall.SIGINT)
		interrupted := make(chan struct{})
		cancel := make(chan struct{}, 1)
		go func() {
			// When SIGINT is received, sleep for InterruptDeadline before the
			// closing interrupted channel.
			select {
			case <-intCh:
			case <-cancel:
				return
			}
			select {
			case <-time.After(InterruptDeadline):
			case <-cancel:
				return
			}
			close(interrupted)
		}()

		// Wait for all forms to finish and collect error returns, unless an
		// interrupt was received and the form didn't quit within
		// InterruptDeadline.
		errors := make([]Error, len(ops))
		for i, errorChan := range errorChans {
			select {
			case errors[i] = <-errorChan:
			case <-interrupted:
				errors[i] = Error{ErrStillRunning}
			}
		}

		// Make sure the SIGINT listener exits.
		close(cancel)
		signal.Stop(intCh)

		// Make sure I am in foreground.
		if PutInForeground && sys.IsATTY(0) {
			err := sys.Tcsetpgrp(0, syscall.Getpgrp())
			if err != nil {
				throw(err)
			}
		}

		if !allok(errors) {
			if len(errors) == 1 {
				throw(errors[0].inner)
			} else {
				throw(multiError{errors})
			}
		}
	}
}

func (cp *compiler) form(n *parse.Form) Op {
	if len(n.Assignments) > 0 {
		if n.Head != nil {
			cp.errorf(n.Begin(), "temporary assignments not yet supported")
		}
		ops := cp.assignments(n.Assignments)
		return func(ec *EvalCtx) {
			for _, op := range ops {
				op(ec)
			}
		}
	}

	if n.Control != nil {
		return cp.control(n.Control)
	}

	headStr, ok := oneString(n.Head)
	if ok {
		compileForm, ok := builtinSpecials[headStr]
		if ok {
			// special form
			return compileForm(cp, n)
		}
		// Ignore the output. If a matching function exists it will be
		// captured and eventually the Evaler executes it. If not, nothing
		// happens here and the Evaler executes an external command.
		cp.registerVariableGet(FnPrefix + headStr)
		// XXX Dynamic head names should always refer to external commands
	}
	headOp := cp.compound(n.Head)
	argOps := cp.compounds(n.Args)
	// TODO: n.NamedArgs
	redirOps := cp.redirs(n.Redirs)
	// TODO: n.ErrorRedir

	p := n.Begin()
	// ec here is always a subevaler created in compiler.pipeline, so it can
	// be safely modified.
	return func(ec *EvalCtx) {
		// head
		headValues := headOp(ec)
		headMust := ec.must(headValues, "the head of command", p)
		headMust.mustLen(1)
		switch headValues[0].(type) {
		case String, Caller, Indexer:
		default:
			headMust.error("a string or callable", headValues[0].Kind())
		}

		// args
		var args []Value
		for _, argOp := range argOps {
			args = append(args, argOp(ec)...)
		}

		// redirs
		for _, redirOp := range redirOps {
			redirOp(ec)
		}

		ec.resolveCaller(headValues[0]).Call(ec, args)
	}
}

func (cp *compiler) control(n *parse.Control) Op {
	switch n.Kind {
	case parse.IfControl:
		condOps := cp.errorCaptures(n.Conditions)
		bodyOps := cp.chunks(n.Bodies)
		var elseOp Op
		if n.ElseBody != nil {
			elseOp = cp.chunk(n.ElseBody)
		}
		return func(ec *EvalCtx) {
			for i, condOp := range condOps {
				if condOp(ec)[0].(Error).inner == nil {
					bodyOps[i](ec)
					return
				}
			}
			if elseOp != nil {
				elseOp(ec)
			}
		}
	case parse.WhileControl:
		condOp := cp.errorCapture(n.Condition)
		bodyOp := cp.chunk(n.Body)
		return func(ec *EvalCtx) {
			for condOp(ec)[0].(Error).inner == nil {
				ex := ec.PEval(bodyOp)
				if ex == Continue {
					// do nothing
				} else if ex == Break {
					break
				} else {
					throw(ex)
				}
			}
		}
	case parse.ForControl:
		iteratorOp := cp.singleVariable(n.Iterator, "must be a single variable")
		valuesOp := cp.array(n.Array)
		bodyOp := cp.chunk(n.Body)
		return func(ec *EvalCtx) {
			iterator := iteratorOp(ec)
			values := valuesOp(ec)
			for _, v := range values {
				doSet(ec, []Variable{iterator}, []Value{v})
				ex := ec.PEval(bodyOp)
				if ex == Continue {
					// do nothing
				} else if ex == Break {
					break
				} else {
					throw(ex)
				}
			}
		}
	case parse.BeginControl:
		return cp.chunk(n.Body)
	default:
		cp.errorf(n.Begin(), "unknown ControlKind %s, compiler bug", n.Kind)
		panic("unreachable")
	}
}

func (cp *compiler) literal(n *parse.Primary, msg string) string {
	switch n.Type {
	case parse.Bareword, parse.SingleQuoted, parse.DoubleQuoted:
		return n.Value
	default:
		cp.errorf(n.Begin(), msg)
		return "" // not reached
	}
}

func (cp *compiler) assignment(n *parse.Assignment) Op {
	variableOps := cp.multiVariable(n.Dst)
	valuesOp := cp.compound(n.Src)

	return func(ec *EvalCtx) {
		variables := make([]Variable, len(variableOps))
		for i, variableOp := range variableOps {
			variables[i] = variableOp(ec)
		}
		doSet(ec, variables, valuesOp(ec))
	}
}

func (cp *compiler) multiVariable(n *parse.Indexing) []VariableOp {
	var variableOps []VariableOp
	if n.Head.Type == parse.Braced {
		// XXX ignore n.Indicies.
		compounds := n.Head.Braced
		indexings := make([]*parse.Indexing, len(compounds))
		for i, cn := range compounds {
			if len(cn.Indexings) != 1 {
				cp.errorf(cn.Begin(), "must be a variable spec")
			}
			indexings[i] = cn.Indexings[0]
		}
		variableOps = cp.singleVariables(indexings, "must be a variable spc")
	} else {
		variableOps = []VariableOp{cp.singleVariable(n, "must be a variable spec or a braced list of those")}
	}
	return variableOps
}

func (cp *compiler) singleVariable(n *parse.Indexing, msg string) VariableOp {
	// XXX will we be using this for purposes other than setting?
	varname := cp.literal(n.Head, msg)
	p := n.Begin()

	if len(n.Indicies) == 0 {
		cp.registerVariableSet(varname)

		return func(ec *EvalCtx) Variable {
			splice, ns, barename := parseVariable(varname)
			if splice {
				// XXX
				ec.errorf(p, "not yet supported")
			}
			variable := ec.ResolveVar(ns, barename)
			if variable == nil {
				// New variable.
				// XXX We depend on the fact that this variable will
				// immeidately be set.
				variable = NewPtrVariable(nil)
				ec.local[barename] = variable
			}
			return variable
		}
	}
	cp.registerVariableGet(varname)
	indexOps := cp.arrays(n.Indicies)
	indexBegins := make([]int, len(n.Indicies))
	indexEnds := make([]int, len(n.Indicies))
	for i, in := range n.Indicies {
		indexBegins[i] = in.Begin()
		indexEnds[i] = in.End()
	}

	return func(ec *EvalCtx) Variable {
		splice, ns, name := parseVariable(varname)
		if splice {
			// XXX
			ec.errorf(p, "not yet supported")
		}
		variable := ec.ResolveVar(ns, name)
		if variable == nil {
			ec.errorf(p, "variable $%s does not exisit, compiler bug", varname)
		}
		if len(indexOps) == 0 {
			// Just a variable, return directly.
			return variable
		}

		// Indexing. Do Index up to the last but one index.
		value := variable.Get()
		n := len(indexOps)
		for i, op := range indexOps[:n-1] {
			indexer, ok := value.(Indexer)
			if !ok {
				ec.errorf( /* from p to */ indexBegins[i], "cannot be indexed (value is %s, type %s)", value.Repr(), value.Kind())
			}

			indicies := op(ec)
			if len(indicies) != 1 {
				ec.errorf(indexBegins[i], "index must eval to a single Value (got %v)", indicies)
			}

			value = indexer.Index(indicies[0])
		}
		// Now this must be an IndexSetter.
		indexSetter, ok := value.(IndexSetter)
		if !ok {
			ec.errorf( /* from p to */ indexBegins[n-1], "cannot be indexed for setting (value is %s, type %s)", value.Repr(), value.Kind())
		}
		// XXX Duplicate code.
		indicies := indexOps[n-1](ec)
		if len(indicies) != 1 {
			ec.errorf(indexBegins[n-1], "index must eval to a single Value (got %v)", indicies)
		}
		return elemVariable{indexSetter, indicies[0]}
	}
}

func makeFlag(m parse.RedirMode) int {
	switch m {
	case parse.Read:
		return os.O_RDONLY
	case parse.Write:
		return os.O_WRONLY | os.O_CREATE
	case parse.ReadWrite:
		return os.O_RDWR | os.O_CREATE
	case parse.Append:
		return os.O_WRONLY | os.O_CREATE | os.O_APPEND
	default:
		// XXX should report parser bug
		panic("bad RedirMode; parser bug")
	}
}

const defaultFileRedirPerm = 0644

// redir compiles a Redir into a op.
func (cp *compiler) redir(n *parse.Redir) Op {
	var dstOp ValuesOp
	if n.Dest != nil {
		dstOp = cp.compound(n.Dest)
	}
	p := n.Begin()
	srcOp := cp.compound(n.Source)
	sourceIsFd := n.SourceIsFd
	pSrc := n.Source.Begin()
	mode := n.Mode
	flag := makeFlag(mode)

	return func(ec *EvalCtx) {
		var dst int
		if dstOp == nil {
			// use default dst fd
			switch mode {
			case parse.Read:
				dst = 0
			case parse.Write, parse.ReadWrite, parse.Append:
				dst = 1
			default:
				// XXX should report parser bug
				panic("bad RedirMode; parser bug")
			}
		} else {
			// dst must be a valid fd
			dst = ec.must(dstOp(ec), "FD", p).mustOneNonNegativeInt()
		}

		ec.growPorts(dst + 1)
		// Logger.Printf("closing old port %d of %s", dst, ec.context)
		ec.ports[dst].Close()

		srcMust := ec.must(srcOp(ec), "redirection source", pSrc)
		src := string(srcMust.mustOneStr())
		if sourceIsFd {
			if src == "-" {
				// close
				ec.ports[dst] = &Port{}
			} else {
				fd := srcMust.zerothMustNonNegativeInt()
				ec.ports[dst] = ec.ports[fd].Fork()
			}
		} else {
			f, err := os.OpenFile(src, flag, defaultFileRedirPerm)
			if err != nil {
				ec.errorf(p, "failed to open file %q: %s", src, err)
			}
			ec.ports[dst] = &Port{
				File: f, Chan: make(chan Value), CloseFile: true, CloseChan: true,
			}
		}
	}
}

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

func mustGetHome(uname string) string {
	dir, err := osutil.GetHome(uname)
	if err != nil {
		throw(err)
	}
	return dir
}

func doGlob(gp GlobPattern) []Value {
	names := glob.Pattern(gp).Glob()
	vs := make([]Value, len(names))
	for i, name := range names {
		vs[i] = String(name)
	}
	return vs
}

func catOps(ops []ValuesOp) ValuesOp {
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

func (cp *compiler) array(n *parse.Array) ValuesOp {
	return catOps(cp.compounds(n.Compounds))
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
			index := indexOp(ec)
			vs = outerProduct(vs, index, func(l, r Value) Value {
				return mustIndexer(l).Index(r)
			})
		}
		return vs
	}
}

func mustIndexer(v Value) Indexer {
	indexer, ok := v.(Indexer)
	if !ok {
		throw(fmt.Errorf("%s value cannot be indexed", v.Kind()))
	}
	return indexer
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
			list, ok := value.(List)
			if !ok {
				ec.errorf(p, "variable $%s is not a list", qname)
			}
			return *list.inner
		}
		return []Value{value}
	}
}

func (cp *compiler) registerVariableGet(qname string) bool {
	_, ns, name := parseVariable(qname)
	if ns != "" && ns != "local" && ns != "up" {
		// Variable in another mod, do nothing
		return true
	}
	// Find in local scope
	if ns == "" || ns == "local" {
		if cp.thisScope()[name] {
			return true
		} else if ns == "local" {
			return false
		}
	}
	// Find in upper scopes
	for i := len(cp.scopes) - 2; i >= 0; i-- {
		if cp.scopes[i][name] {
			// Existing name: record capture and return.
			cp.capture[name] = true
			return true
		}
	}
	return false
}

func (cp *compiler) registerVariableSet(qname string) bool {
	_, ns, name := parseVariable(qname)
	switch ns {
	case "local":
		cp.thisScope()[name] = true
		return true
	case "up":
		for i := len(cp.scopes) - 2; i >= 0; i-- {
			if cp.scopes[i][name] {
				// Existing name: record capture and return.
				cp.capture[name] = true
				return true
			}
		}
		return false
	case "":
		if cp.thisScope()[name] {
			// A name on current scope. Do nothing.
			return true
		}
		// Walk up the upper scopes
		for i := len(cp.scopes) - 2; i >= 0; i-- {
			if cp.scopes[i][name] {
				// Existing name. Do nothing
				cp.capture[name] = true
				return true
			}
		}
		// New name. Register on this scope!
		cp.thisScope()[name] = true
		return true
	default:
		// Variable in another mod, do nothing
		return true
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
	p := n.Chunk.Begin()
	return func(ec *EvalCtx) []Value {
		vs := []Value{}
		newEc := ec.fork(fmt.Sprintf("channel output capture %v", op))

		pipeRead, pipeWrite, err := os.Pipe()
		if err != nil {
			ec.errorf(p, "failed to create pipe: %v", err)
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
					log.Println()
					break
				}
				ch <- String(line[:len(line)-1])
			}
			bytesCollected <- true
		}()

		// XXX The exitus is discarded.
		op(newEc)
		ClosePorts(newEc.ports)

		<-bytesCollected
		close(ch)
		<-chCollected

		return vs
	}
}

func (cp *compiler) pushScope() scope {
	sc := scope{}
	cp.scopes = append(cp.scopes, sc)
	return sc
}

func (cp *compiler) popScope() {
	cp.scopes[len(cp.scopes)-1] = nil
	cp.scopes = cp.scopes[:len(cp.scopes)-1]
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
	nn := len(n.MapPairs)
	keysOps := make([]ValuesOp, nn)
	valuesOps := make([]ValuesOp, nn)
	poses := make([]int, nn)
	for i := 0; i < nn; i++ {
		keysOps[i] = cp.compound(n.MapPairs[i].Key)
		valuesOps[i] = cp.compound(n.MapPairs[i].Value)
		poses[i] = n.MapPairs[i].Begin()
	}
	return func(ec *EvalCtx) []Value {
		m := make(map[Value]Value)
		for i := 0; i < nn; i++ {
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
	return catOps(ops)
}

// parseVariable parses a variable name.
func parseVariable(qname string) (splice bool, ns string, name string) {
	if strings.HasPrefix(qname, "@") {
		splice = true
		qname = qname[1:]
		if qname == "" {
			qname = "args"
		}
	}

	i := strings.IndexRune(qname, ':')
	if i == -1 {
		return splice, "", qname
	}
	return splice, qname[:i], qname[i+1:]
}
