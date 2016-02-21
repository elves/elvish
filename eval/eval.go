// Package eval handles evaluation of nodes and consists the runtime of the
// shell.
package eval

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
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
	global  Namespace
	modules map[string]Namespace
	store   *store.Store
	Stub    *stub.Stub
}

// EvalCtx maintains an Evaler along with its runtime context. After creation
// an EvalCtx is not modified, and new instances are created when needed.
type EvalCtx struct {
	*Evaler
	name, text, context string

	local, up Namespace
	ports     []*Port

	begin, end int
}

func (ec *EvalCtx) evaling(n parse.Node) {
	ec.begin, ec.end = n.Begin(), n.End()
}

// NewEvaler creates a new Evaler.
func NewEvaler(st *store.Store) *Evaler {
	ev := &Evaler{nil, map[string]Namespace{}, st, nil}

	// Construct initial global namespace
	pid := String(strconv.Itoa(syscall.Getpid()))
	ev.global = Namespace{
		"pid":   NewRoVariable(pid),
		"ok":    NewRoVariable(OK),
		"true":  NewRoVariable(Bool(true)),
		"false": NewRoVariable(Bool(false)),
		"paths": &EnvPathList{envName: "PATH"},
		"pwd":   PwdVariable{},
	}
	for _, b := range builtinFns {
		ev.global[FnPrefix+b.Name] = NewRoVariable(b)
	}

	return ev
}

func (e *Evaler) searchPaths() []string {
	return e.global["paths"].(*EnvPathList).get()
}

func (e *Evaler) AddModule(name string, ns Namespace) {
	e.modules[name] = ns
}

const (
	outChanSize   = 32
	outChanLeader = "▶ "
	initIndent    = 2
)

// NewTopEvalCtx creates a top-level evalCtx.
func NewTopEvalCtx(ev *Evaler, name, text string, ports []*Port) *EvalCtx {
	return &EvalCtx{
		ev,
		name, text, "top",
		ev.global, Namespace{},
		ports, 0, len(text),
	}
}

// fork returns a modified copy of ec. The ports are forked, and the context is
// changed to the given value. Other fields are copied shallowly.
func (ec *EvalCtx) fork(newContext string) *EvalCtx {
	newPorts := make([]*Port, len(ec.ports))
	for i, p := range ec.ports {
		newPorts[i] = p.Fork()
	}
	return &EvalCtx{
		ec.Evaler,
		ec.name, ec.text, newContext,
		ec.local, ec.up,
		newPorts, ec.begin, ec.end,
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

// Eval evaluates a chunk node n. The supplied name and text are used in
// diagnostic messages.
func (ev *Evaler) Eval(name, text string, n *parse.Chunk, ports []*Port) error {
	op, err := ev.Compile(n)
	if err != nil {
		return err
	}
	ec := NewTopEvalCtx(ev, name, text, ports)
	return ec.PEval(op)
}

func (ev *Evaler) EvalInteractive(text string, n *parse.Chunk) error {
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
		{File: os.Stderr},
	}

	signal.Ignore(syscall.SIGTTIN)
	signal.Ignore(syscall.SIGTTOU)
	// XXX Should use fd of /dev/terminal instead of 0.
	if ev.Stub != nil && sys.IsATTY(0) {
		ev.Stub.SetTitle(summarize(text))
		err := sys.Tcsetpgrp(0, ev.Stub.Process().Pid)
		if err != nil {
			fmt.Println("failed to put stub in foreground:", err)
		}
	}

	err := ev.Eval("[interactive]", text, n, ports)
	close(outCh)
	<-outDone

	// XXX Should use fd of /dev/terminal instead of 0.
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
func (ev *Evaler) Compile(n *parse.Chunk) (Op, error) {
	return compile(makeScope(ev.global), n)
}

// PEval evaluates an op in a protected environment so that calls to errorf are
// wrapped in an Error.
func (ec *EvalCtx) PEval(op Op) (err error) {
	// defer catch(&err, ec)
	defer util.Catch(&err)
	op.Exec(ec)
	return nil
}

func (ec *EvalCtx) PCall(f Caller, args []Value) (err error) {
	// defer catch(&err, ec)
	defer util.Catch(&err)
	f.Call(ec, args)
	return nil
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
		if _, ok := err.(*util.PosError); !ok {
			err = &util.PosError{ec.begin, ec.end, err}
		}
		*perr = err
	} else if r != nil {
		panic(r)
	}
}

// errorpf stops the ec.eval immediately by panicking with a diagnostic message.
// The panic is supposed to be caught by ec.eval.
func (ec *EvalCtx) errorpf(begin, end int, format string, args ...interface{}) {
	throw(&util.PosError{begin, end, fmt.Errorf(format, args...)})
}

func (ec *EvalCtx) errorf(format string, args ...interface{}) {
	ec.errorpf(ec.begin, ec.end, format, args...)
}

// SourceText evaluates a chunk of elvish source.
func (ev *Evaler) SourceText(src string) error {
	n, err := parse.Parse(src)
	if err != nil {
		return err
	}
	return ev.EvalInteractive(src, n)
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
	return ev.SourceText(src)
}

// Global returns the global namespace.
func (ev *Evaler) Global() map[string]Variable {
	return map[string]Variable(ev.global)
}

// ResolveVar resolves a variable. When the variable cannot be found, nil is
// returned.
func (ec *EvalCtx) ResolveVar(ns, name string) Variable {
	if ns == "env" {
		ev := envVariable{name}
		return ev
	}
	if mod, ok := ec.modules[ns]; ok {
		return mod[name]
	}

	may := func(n string) bool {
		return ns == "" || ns == n
	}
	if may("local") {
		if v, ok := ec.local[name]; ok {
			return v
		}
	}
	if may("up") {
		if v, ok := ec.up[name]; ok {
			return v
		}
	}
	return nil
}
