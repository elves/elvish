// Package eval handles evaluation of nodes and consists the runtime of the
// shell.
package eval

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unicode/utf8"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/store"
	"github.com/elves/elvish/stub"
	"github.com/elves/elvish/sys"
	"github.com/elves/elvish/util"
)

var Logger = util.GetLogger("[eval] ")

// FnPrefix is the prefix for the variable names of functions. Defining a
// function "foo" is equivalent to setting a variable named FnPrefix + "foo".
const FnPrefix = "&"

// Namespace is a map from name to variables.
type Namespace map[string]Variable

// Evaler is used to evaluate elvish sources. It maintains runtime context
// shared among all evalCtx instances.
type Evaler struct {
	Global  Namespace
	Modules map[string]Namespace
	Store   *store.Store
	Editor  Editor
	Stub    *stub.Stub
	intCh   chan struct{}
}

// EvalCtx maintains an Evaler along with its runtime context. After creation
// an EvalCtx is seldom modified, and new instances are created when needed.
type EvalCtx struct {
	*Evaler
	name, srcName, src string

	local, up   Namespace
	ports       []*Port
	positionals []Value
	verdict     bool

	begin, end int
	traceback  *util.Traceback
}

func (ec *EvalCtx) falsify() {
	ec.verdict = false
}

func (ec *EvalCtx) evaling(begin, end int) {
	ec.begin, ec.end = begin, end
}

// NewEvaler creates a new Evaler.
func NewEvaler(st *store.Store) *Evaler {
	return &Evaler{Namespace{}, map[string]Namespace{}, st, nil, nil, nil}
}

func (e *Evaler) searchPaths() []string {
	return builtinNamespace["paths"].(*EnvPathList).get()
}

const (
	outChanSize    = 32
	outChanLeader  = "▶ "
	falseIndicator = "✗"
	initIndent     = NoPretty
)

// NewTopEvalCtx creates a top-level evalCtx.
func NewTopEvalCtx(ev *Evaler, name, text string, ports []*Port) *EvalCtx {
	return &EvalCtx{
		ev, "top",
		name, text,
		ev.Global, Namespace{},
		ports, nil, true,
		0, len(text), nil,
	}
}

// fork returns a modified copy of ec. The ports are forked, and the name is
// changed to the given value. Other fields are copied shallowly.
func (ec *EvalCtx) fork(name string) *EvalCtx {
	newPorts := make([]*Port, len(ec.ports))
	for i, p := range ec.ports {
		newPorts[i] = p.Fork()
	}
	return &EvalCtx{
		ec.Evaler, name,
		ec.srcName, ec.src,
		ec.local, ec.up,
		newPorts, ec.positionals, true,
		ec.begin, ec.end, ec.traceback,
	}
}

// port returns ec.ports[i] or nil if i is out of range. This makes it possible
// to treat ec.ports as if it has an infinite tail of nil's.
func (ec *EvalCtx) port(i int) *Port {
	if i >= len(ec.ports) {
		return nil
	}
	return ec.ports[i]
}

// growPorts makes the size of ec.ports at least n, adding nil's if necessary.
func (ec *EvalCtx) growPorts(n int) {
	if len(ec.ports) >= n {
		return
	}
	ports := ec.ports
	ec.ports = make([]*Port, n)
	copy(ec.ports, ports)
}

func makeScope(s Namespace) scope {
	sc := scope{}
	for name := range s {
		sc[name] = true
	}
	return sc
}

// eval evaluates a chunk node n. The supplied name and text are used in
// diagnostic messages.
func (ev *Evaler) eval(op Op, ports []*Port, name, text string) (bool, error) {
	ec := NewTopEvalCtx(ev, name, text, ports)
	err := ec.PEval(op)
	return ec.verdict, err
}

func (ec *EvalCtx) Interrupts() <-chan struct{} {
	return ec.intCh
}

// Eval sets up the Evaler and evaluates a chunk. The supplied name and text are
// used in diagnostic messages.
func (ev *Evaler) Eval(op Op, name, text string) error {
	inCh := make(chan Value)
	close(inCh)

	outCh := make(chan Value, outChanSize)
	outDone := make(chan struct{})
	go func() {
		for v := range outCh {
			fmt.Printf("%s%s\n", outChanLeader, v.Repr(initIndent))
		}
		close(outDone)
	}()

	ports := []*Port{
		{File: os.Stdin, Chan: inCh},
		{File: os.Stdout, Chan: outCh},
		{File: os.Stderr, Chan: BlackholeChan},
	}

	signal.Ignore(syscall.SIGTTIN)
	signal.Ignore(syscall.SIGTTOU)
	stopSigGoroutine := make(chan struct{})
	sigGoRoutineDone := make(chan struct{})
	// XXX Should use fd of /dev/terminal instead of 0.
	if ev.Stub != nil && ev.Stub.Alive() && sys.IsATTY(0) {
		ev.Stub.SetTitle(summarize(text))
		dir, err := os.Getwd()
		if err != nil {
			dir = "/"
		}
		ev.Stub.Chdir(dir)
		err = sys.Tcsetpgrp(0, ev.Stub.Process().Pid)
		if err != nil {
			fmt.Println("failed to put stub in foreground:", err)
		}

		ev.intCh = make(chan struct{})
		go func() {
			closedIntCh := false
		loop:
			for {
				select {
				case sig := <-ev.Stub.Signals():
					switch sig {
					case syscall.SIGINT, syscall.SIGQUIT:
						if !closedIntCh {
							close(ev.intCh)
							closedIntCh = true
						}
					}
				case <-stopSigGoroutine:
					break loop
				}
			}
			ev.intCh = nil
			close(sigGoRoutineDone)
		}()
	} else {
		close(sigGoRoutineDone)
	}

	ret, err := ev.eval(op, ports, name, text)
	close(outCh)
	<-outDone
	close(stopSigGoroutine)
	<-sigGoRoutineDone

	if !ret {
		fmt.Println(falseIndicator)
	}

	// XXX Should use fd of /dev/tty instead of 0.
	if sys.IsATTY(0) {
		err := sys.Tcsetpgrp(0, syscall.Getpgrp())
		if err != nil {
			fmt.Println("failed to put myself in foreground:", err)
		}
	}

	return err
}

func summarize(text string) string {
	// TODO Make a proper summary.
	if len(text) < 32 {
		return text
	}
	var b bytes.Buffer
	for i, r := range text {
		if i+len(string(r)) >= 32 {
			break
		}
		b.WriteRune(r)
	}
	return b.String()
}

// Compile compiles elvish code in the global scope.
func (ev *Evaler) Compile(n *parse.Chunk, name, text string) (Op, error) {
	return compile(makeScope(ev.Global), n, name, text)
}

// PEval evaluates an op in a protected environment so that calls to errorf are
// wrapped in an Error.
func (ec *EvalCtx) PEval(op Op) (err error) {
	defer catch(&err, ec)
	op.Exec(ec)
	return nil
}

func (ec *EvalCtx) PCall(f Fn, args []Value, opts map[string]Value) (err error) {
	defer catch(&err, ec)
	f.Call(ec, args, opts)
	return nil
}

func (ec *EvalCtx) PCaptureOutput(f Fn, args []Value, opts map[string]Value) (vs []Value, err error) {
	defer catch(&err, ec)
	// XXX There is no source.
	return captureOutput(ec, Op{
		func(newec *EvalCtx) { f.Call(newec, args, opts) }, -1, -1}), nil
}

func catch(perr *error, ec *EvalCtx) {
	// NOTE: We have to duplicate instead of calling util.Catch here, since
	// recover can only catch a panic when called directly from a deferred
	// function.
	r := recover()
	if r == nil {
		return
	}
	if exc, ok := r.(util.Exception); ok {
		err := exc.Error
		if _, ok := err.(*util.TracebackError); !ok {
			if _, ok := err.(flow); !ok {
				err = ec.makeTracebackError(err)
			}
		}
		*perr = err
	} else if r != nil {
		panic(r)
	}
}

func (ec *EvalCtx) makeTracebackError(e error) *util.TracebackError {
	return &util.TracebackError{Cause: e, Traceback: ec.addTraceback()}
}

func (ec *EvalCtx) addTraceback() *util.Traceback {
	return &util.Traceback{
		Name: ec.srcName, Source: ec.src,
		Begin: ec.begin, End: ec.end, Next: ec.traceback,
	}
}

// errorpf stops the ec.eval immediately by panicking with a diagnostic message.
// The panic is supposed to be caught by ec.eval.
func (ec *EvalCtx) errorpf(begin, end int, format string, args ...interface{}) {
	ec.begin, ec.end = begin, end
	throwf(format, args...)
}

// SourceText evaluates a chunk of elvish source.
func (ev *Evaler) SourceText(name, src string) error {
	n, err := parse.Parse(name, src)
	if err != nil {
		return err
	}
	op, err := ev.Compile(n, name, src)
	if err != nil {
		return err
	}
	return ev.Eval(op, name, src)
}

func readFileUTF8(fname string) (string, error) {
	bytes, err := ioutil.ReadFile(fname)
	if err != nil {
		return "", err
	}
	if !utf8.Valid(bytes) {
		return "", fmt.Errorf("%s: source is not valid UTF-8", fname)
	}
	return string(bytes), nil
}

// Source evaluates the content of a file.
func (ev *Evaler) Source(fname string) error {
	src, err := readFileUTF8(fname)
	if err != nil {
		return err
	}
	return ev.SourceText(fname, src)
}

// Builtin returns the builtin namespace.
func Builtin() Namespace {
	return builtinNamespace
}

// ErrStoreUnconnected is thrown by ResolveVar when a shared: variable needs to
// be resolved but the store is not connected.
var ErrStoreUnconnected = errors.New("store unconnected")

// ResolveVar resolves a variable. When the variable cannot be found, nil is
// returned.
func (ec *EvalCtx) ResolveVar(ns, name string) Variable {
	switch ns {
	case "local":
		return ec.getLocal(name)
	case "up":
		return ec.up[name]
	case "builtin":
		return builtinNamespace[name]
	case "":
		if v := ec.getLocal(name); v != nil {
			return v
		}
		if v, ok := ec.up[name]; ok {
			return v
		}
		return builtinNamespace[name]
	case "e", "E":
		if strings.HasPrefix(name, FnPrefix) {
			return NewRoVariable(ExternalCmd{name[len(FnPrefix):]})
		}
		return envVariable{name}
	case "shared":
		if ec.Store == nil {
			throw(ErrStoreUnconnected)
		}
		return sharedVariable{ec.Store, name}
	default:
		if ns, ok := ec.Modules[ns]; ok {
			return ns[name]
		} else {
			return nil
		}
	}
}

// getLocal finds the named local variable.
func (ec *EvalCtx) getLocal(name string) Variable {
	i, err := strconv.Atoi(name)
	if err == nil {
		// Logger.Println("positional variable", i)
		// Logger.Printf("EvalCtx=%p, args=%v", ec, ec.positionals)
		if i < 0 {
			i += len(ec.positionals)
		}
		if i < 0 || i >= len(ec.positionals) {
			// Logger.Print("out of range")
			return nil
		}
		// Logger.Print("found")
		return NewRoVariable(ec.positionals[i])
	}
	return ec.local[name]
}

var ErrMoreThanOneRest = errors.New("more than one @ lvalue")

// IterateInputs calls the passed function for each input element.
func (ec *EvalCtx) IterateInputs(f func(Value)) {
	var w sync.WaitGroup
	inputs := make(chan Value)

	w.Add(2)
	go func() {
		linesToChan(ec.ports[0].File, inputs)
		w.Done()
	}()
	go func() {
		for v := range ec.ports[0].Chan {
			inputs <- v
		}
		w.Done()
	}()
	go func() {
		w.Wait()
		close(inputs)
	}()

	for v := range inputs {
		f(v)
	}
}

func linesToChan(r io.Reader, ch chan<- Value) {
	filein := bufio.NewReader(r)
	for {
		line, err := filein.ReadString('\n')
		if line != "" {
			ch <- String(strings.TrimSuffix(line, "\n"))
		}
		if err != nil {
			if err != io.EOF {
				Logger.Println("error on reading:", err)
			}
			break
		}
	}
}

// OutputChan returns a channel onto which output can be written.
func (ec *EvalCtx) OutputChan() chan<- Value {
	return ec.ports[1].Chan
}
