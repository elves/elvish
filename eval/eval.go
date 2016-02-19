// Package eval handles evaluation of nodes and consists the runtime of the
// shell.
package eval

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"syscall"
	"unicode/utf8"

	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/store"
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
}

// EvalCtx maintains an Evaler along with its runtime context. After creation
// an EvalCtx is not modified, and new instances are created when needed.
type EvalCtx struct {
	*Evaler
	name, text, context string

	local, up Namespace
	ports     []*Port
}

// NewEvaler creates a new Evaler.
func NewEvaler(st *store.Store) *Evaler {
	ev := &Evaler{nil, map[string]Namespace{}, st}

	// Construct initial global namespace
	pid := String(strconv.Itoa(syscall.Getpid()))
	ev.global = Namespace{
		"pid":   NewRoVariable(pid),
		"ok":    NewRoVariable(OK),
		"true":  NewRoVariable(Bool(true)),
		"false": NewRoVariable(Bool(false)),
		"paths": &EnvPathList{envName: "PATH"},
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
)

// NewTopEvalCtx creates a top-level evalCtx.
func NewTopEvalCtx(ev *Evaler, name, text string, ports []*Port) *EvalCtx {
	return &EvalCtx{
		ev,
		name, text, "top",
		ev.global, Namespace{},
		ports,
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
		newPorts,
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
	op, err := ev.Compile(name, text, n)
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
			fmt.Printf("%s%s\n", outChanLeader, v.Repr())
		}
		close(outDone)
	}()

	ports := []*Port{
		{File: os.Stdin, Chan: inCh},
		{File: os.Stdout, Chan: outCh},
		{File: os.Stderr},
	}

	err := ev.Eval("[interactive]", text, n, ports)
	close(outCh)
	<-outDone
	return err
}

// Compile compiles elvish code in the global scope.
func (ev *Evaler) Compile(name, text string, n *parse.Chunk) (Op, error) {
	return compile(name, text, makeScope(ev.global), n)
}

// PEval evaluates an op in a protected environment so that calls to errorf are
// wrapped in an Error.
func (ec *EvalCtx) PEval(op Op) (ex error) {
	defer util.Catch(&ex)
	op(ec)
	return nil
}

func (ec *EvalCtx) PCall(f Caller, args []Value) (ex error) {
	defer util.Catch(&ex)
	f.Call(ec, args)
	return nil
}

// errorf stops the ec.eval immediately by panicking with a diagnostic message.
// The panic is supposed to be caught by ec.eval.
func (ec *EvalCtx) errorf(p int, format string, args ...interface{}) {
	throw(util.NewContextualError(
		fmt.Sprintf("%s (%s)", ec.name, ec.context), "error",
		ec.text, p, format, args...))
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
