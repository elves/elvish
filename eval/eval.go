// Package eval handles evaluation of nodes and consists the runtime of the
// shell.
package eval

//go:generate ./bundle-modules

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"unicode/utf8"

	"github.com/elves/elvish/daemon/api"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/program/daemon"
	"github.com/elves/elvish/sys"
	"github.com/elves/elvish/util"
)

var logger = util.GetLogger("[eval] ")

// FnPrefix is the prefix for the variable names of functions. Defining a
// function "foo" is equivalent to setting a variable named FnPrefix + "foo".
const FnPrefix = "&"

const (
	outChanSize              = 32
	defaultValueOutIndicator = "▶ "
	initIndent               = NoPretty
)

// Evaler is used to evaluate elvish sources. It maintains runtime context
// shared among all evalCtx instances.
type Evaler struct {
	Global  Scope
	Builtin Scope
	Modules map[string]Namespace
	Daemon  *api.Client
	ToSpawn *daemon.Daemon
	Editor  Editor
	DataDir string
	intCh   chan struct{}

	// Configurations.
	valueOutIndicator string
}

// EvalCtx maintains an Evaler along with its runtime context. After creation
// an EvalCtx is seldom modified, and new instances are created when needed.
type EvalCtx struct {
	*Evaler
	name    string
	srcName string
	src     string
	modPath string // Only nonempty when evaluating a module.

	local, up Scope
	ports     []*Port

	begin, end int
	traceback  *util.SourceContext

	background bool
}

// NewEvaler creates a new Evaler.
func NewEvaler(daemon *api.Client, toSpawn *daemon.Daemon,
	dataDir string, extraModules map[string]Namespace) *Evaler {

	builtin := Scope{makeBuiltinNamespace(daemon), map[string]Namespace{}}

	// TODO(xiaq): Create daemon namespace asynchronously.
	modules := map[string]Namespace{
		"daemon":  makeDaemonNamespace(daemon),
		"builtin": builtin.Names,
	}
	for name, mod := range extraModules {
		modules[name] = mod
	}

	ev := &Evaler{
		Global:  makeScope(),
		Builtin: builtin,
		Modules: modules,
		Daemon:  daemon,
		ToSpawn: toSpawn,
		Editor:  nil,
		DataDir: dataDir,
		intCh:   nil,

		valueOutIndicator: defaultValueOutIndicator,
	}
	builtin.Names["value-out-indicator"] = NewBackedVariable(&ev.valueOutIndicator)
	return ev
}

func (ev *Evaler) searchPaths() []string {
	return strings.Split(os.Getenv("PATH"), ":")
}

// NewTopEvalCtx creates a top-level evalCtx.
func NewTopEvalCtx(ev *Evaler, name, text string, ports []*Port) *EvalCtx {
	return &EvalCtx{
		ev, "top",
		name, text, "",
		ev.Global, makeScope(),
		ports,
		0, len(text), nil, false,
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
		ec.srcName, ec.src, ec.modPath,
		ec.local, ec.up,
		newPorts,
		ec.begin, ec.end, ec.traceback, ec.background,
	}
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

// eval evaluates a chunk node n. The supplied name and text are used in
// diagnostic messages.
func (ev *Evaler) eval(op Op, ports []*Port, name, text string) error {
	ec := NewTopEvalCtx(ev, name, text, ports)
	return ec.PEval(op)
}

// Eval sets up the Evaler with standard ports and evaluates an Op. The supplied
// name and text are used in diagnostic messages.
func (ev *Evaler) Eval(op Op, name, text string) error {
	inCh := make(chan Value)
	close(inCh)

	outCh := make(chan Value, outChanSize)
	outDone := make(chan struct{})
	go func() {
		for v := range outCh {
			fmt.Println(ev.valueOutIndicator + v.Repr(initIndent))
		}
		close(outDone)
	}()
	defer func() {
		close(outCh)
		<-outDone
	}()

	ports := []*Port{
		{File: os.Stdin, Chan: inCh},
		{File: os.Stdout, Chan: outCh},
		{File: os.Stderr, Chan: BlackholeChan},
	}

	return ev.EvalWithPorts(ports, op, name, text)
}

// EvalWithPorts sets up the Evaler with the given ports and evaluates an Op.
// The supplied name and text are used in diagnostic messages.
func (ev *Evaler) EvalWithPorts(ports []*Port, op Op, name, text string) error {
	// signal.Ignore(syscall.SIGTTIN)

	// Ignore TTOU.
	// When a subprocess in its own process group puts itself in the foreground,
	// the elvish will be in the background. In that case, elvish will move
	// itself back to the foreground by calling tcsetpgrp. However, whenever a
	// background process calls tcsetpgrp (or otherwise attempts to modify the
	// terminal configuration), TTOU will be sent, whose default handler is to
	// stop the process. When the process lives in an orphaned process group
	// (most likely for elvish), the call will outright fail. Therefore, for
	// elvish to be able to move itself back to the foreground, we need to
	// ignore TTOU.
	ignoreTTOU()
	stopSigGoroutine := make(chan struct{})
	sigGoRoutineDone := make(chan struct{})
	// Set up intCh.
	ev.intCh = make(chan struct{})
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGQUIT)
	go func() {
		closedIntCh := false
	loop:
		for {
			select {
			case <-sigCh:
				if !closedIntCh {
					close(ev.intCh)
					closedIntCh = true
				}
			case <-stopSigGoroutine:
				break loop
			}
		}
		ev.intCh = nil
		signal.Stop(sigCh)
		close(sigGoRoutineDone)
	}()

	err := ev.eval(op, ports, name, text)

	close(stopSigGoroutine)
	<-sigGoRoutineDone

	// Put myself in foreground, in case some command has put me in background.
	// XXX Should probably use fd of /dev/tty instead of 0.
	if sys.IsATTY(os.Stdin) {
		err := putSelfInFg()
		if err != nil {
			fmt.Println("failed to put myself in foreground:", err)
		}
	}

	// Un-ignore TTOU.
	unignoreTTOU()

	return err
}

// Compile compiles elvish code in the global scope. If the error is not nil, it
// always has type CompilationError.
func (ev *Evaler) Compile(n *parse.Chunk, name, text string) (Op, error) {
	return compile(ev.Builtin.static(), ev.Global.static(), n, name, text)
}

// PEval evaluates an op in a protected environment so that calls to errorf are
// wrapped in an Error.
func (ec *EvalCtx) PEval(op Op) (err error) {
	defer catch(&err, ec)
	op.Exec(ec)
	return nil
}

func (ec *EvalCtx) PCall(f Callable, args []Value, opts map[string]Value) (err error) {
	defer catch(&err, ec)
	f.Call(ec, args, opts)
	return nil
}

func (ec *EvalCtx) PCaptureOutput(f Callable, args []Value, opts map[string]Value) (vs []Value, err error) {
	// XXX There is no source.
	return pcaptureOutput(ec, Op{
		func(newec *EvalCtx) { f.Call(newec, args, opts) }, -1, -1})
}

func (ec *EvalCtx) PCaptureOutputInner(f Callable, args []Value, opts map[string]Value, valuesCb func(<-chan Value), bytesCb func(*os.File)) error {
	// XXX There is no source.
	return pcaptureOutputInner(ec, Op{
		func(newec *EvalCtx) { f.Call(newec, args, opts) }, -1, -1},
		valuesCb, bytesCb)
}

func catch(perr *error, ec *EvalCtx) {
	// NOTE: We have to duplicate instead of calling util.Catch here, since
	// recover can only catch a panic when called directly from a deferred
	// function.
	r := recover()
	if r == nil {
		return
	}
	if exc, ok := r.(util.Thrown); ok {
		err := exc.Wrapped
		if _, ok := err.(*Exception); !ok {
			err = ec.makeException(err)
		}
		*perr = err
	} else if r != nil {
		panic(r)
	}
}

// makeException turns an error into an Exception by adding traceback.
func (ec *EvalCtx) makeException(e error) *Exception {
	return &Exception{e, ec.addTraceback()}
}

func (ec *EvalCtx) addTraceback() *util.SourceContext {
	return &util.SourceContext{
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

// ErrStoreUnconnected is thrown by ResolveVar when a shared: variable needs to
// be resolved but the store is not connected.
var ErrStoreUnconnected = errors.New("store unconnected")

// ResolveVar resolves a variable. When the variable cannot be found, nil is
// returned.
func (ec *EvalCtx) ResolveVar(ns, name string) Variable {
	switch ns {
	case "local":
		return ec.local.Names[name]
	case "up":
		return ec.up.Names[name]
	case "builtin":
		return ec.Builtin.Names[name]
	case "":
		if v := ec.local.Names[name]; v != nil {
			return v
		}
		if v, ok := ec.up.Names[name]; ok {
			return v
		}
		return ec.Builtin.Names[name]
	case "e":
		if strings.HasPrefix(name, FnPrefix) {
			return NewRoVariable(ExternalCmd{name[len(FnPrefix):]})
		}
	case "E":
		return envVariable{name}
	case "shared":
		if ec.Daemon == nil {
			throw(ErrStoreUnconnected)
		}
		return sharedVariable{ec.Daemon, name}
	default:
		ns := ec.ResolveMod(ns)
		if ns != nil {
			return ns[name]
		}
	}
	return nil
}

func (ec *EvalCtx) ResolveMod(name string) Namespace {
	if ns, ok := ec.local.Uses[name]; ok {
		return ns
	}
	if ns, ok := ec.up.Uses[name]; ok {
		return ns
	}
	if ns, ok := ec.Builtin.Uses[name]; ok {
		return ns
	}
	return nil
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
				logger.Println("error on reading:", err)
			}
			break
		}
	}
}

// InputChan returns a channel from which input can be read.
func (ec *EvalCtx) InputChan() chan Value {
	return ec.ports[0].Chan
}

// InputFile returns a file from which input can be read.
func (ec *EvalCtx) InputFile() *os.File {
	return ec.ports[0].File
}

// OutputChan returns a channel onto which output can be written.
func (ec *EvalCtx) OutputChan() chan<- Value {
	return ec.ports[1].Chan
}

// OutputFile returns a file onto which output can be written.
func (ec *EvalCtx) OutputFile() *os.File {
	return ec.ports[1].File
}
