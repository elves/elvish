// Package eval handles evaluation of parsed Elvish code and provides runtime
// facilities.
package eval

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/elves/elvish/daemon"
	"github.com/elves/elvish/eval/bundled"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/sys"
	"github.com/elves/elvish/util"
	"github.com/xiaq/persistent/vector"
)

var logger = util.GetLogger("[eval] ")

const (
	// FnSuffix is the suffix for the variable names of functions. Defining a
	// function "foo" is equivalent to setting a variable named "foo~", and vice
	// versa.
	FnSuffix = "~"
	// NsSuffix is the suffix for the variable names of namespaces. Defining a
	// namespace foo is equivalent to setting a variable named "foo:", and vice
	// versa.
	NsSuffix = ":"
)

const (
	defaultValuePrefix        = "▶ "
	defaultNotifyBgJobSuccess = true
	initIndent                = vals.NoPretty
)

// Evaler is used to evaluate elvish sources. It maintains runtime context
// shared among all evalCtx instances.
type Evaler struct {
	evalerScopes

	state state

	// Chdir hooks.
	beforeChdir []func(string)
	afterChdir  []func(string)

	// Used to receive SIGINT.
	intCh chan struct{}

	// State of the module system.
	libDir  string
	bundled map[string]string
	modules map[string]Ns

	// Dependencies.
	//
	// TODO: Remove these dependency by providing more general extension points.
	DaemonClient *daemon.Client
	Editor       Editor
}

type evalerScopes struct {
	Global  Ns
	Builtin Ns
}

// NewEvaler creates a new Evaler.
func NewEvaler() *Evaler {
	builtin := builtinNs.Clone()

	ev := &Evaler{
		state: state{
			valuePrefix:        defaultValuePrefix,
			notifyBgJobSuccess: defaultNotifyBgJobSuccess,
			numBgJobs:          0,
		},
		evalerScopes: evalerScopes{
			Global:  make(Ns),
			Builtin: builtin,
		},
		modules: map[string]Ns{
			"builtin": builtin,
		},
		bundled: bundled.Get(),
		Editor:  nil,
		intCh:   nil,
	}

	beforeChdirElvish, afterChdirElvish := vector.Empty, vector.Empty
	ev.beforeChdir = append(ev.beforeChdir,
		adaptChdirHook("before-chdir", ev, &beforeChdirElvish))
	ev.afterChdir = append(ev.afterChdir,
		adaptChdirHook("after-chdir", ev, &afterChdirElvish))
	builtin["before-chdir"] = vars.FromPtr(&beforeChdirElvish)
	builtin["after-chdir"] = vars.FromPtr(&afterChdirElvish)

	builtin["value-out-indicator"] = vars.FromPtrWithMutex(
		&ev.state.valuePrefix, &ev.state.mutex)
	builtin["notify-bg-job-success"] = vars.FromPtrWithMutex(
		&ev.state.notifyBgJobSuccess, &ev.state.mutex)
	builtin["num-bg-jobs"] = vars.FromGet(func() interface{} {
		return strconv.Itoa(ev.state.getNumBgJobs())
	})
	builtin["pwd"] = PwdVariable{ev}

	return ev
}

func adaptChdirHook(name string, ev *Evaler, pfns *vector.Vector) func(string) {
	return func(path string) {
		stdPorts := newStdPorts(
			os.Stdin, os.Stdout, os.Stderr, ev.state.getValuePrefix())
		defer stdPorts.close()
		for it := (*pfns).Iterator(); it.HasElem(); it.Next() {
			fn, ok := it.Elem().(Callable)
			if !ok {
				fmt.Fprintln(os.Stderr, name, "hook must be callable")
				continue
			}
			fm := NewTopFrame(ev,
				NewInternalSource("["+name+" hook]"), stdPorts.ports[:])
			err := fm.Call(fn, []interface{}{path}, NoOpts)
			if err != nil {
				// TODO: Stack trace
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}
}

// Close releases resources allocated when creating this Evaler. Currently this
// does nothing and always returns a nil error.
func (ev *Evaler) Close() error {
	return nil
}

// AddBeforeChdir adds a function to run before changing directory.
func (ev *Evaler) AddBeforeChdir(f func(string)) {
	ev.beforeChdir = append(ev.beforeChdir, f)
}

// AddAfterChdir adds a function to run after changing directory.
func (ev *Evaler) AddAfterChdir(f func(string)) {
	ev.afterChdir = append(ev.afterChdir, f)
}

// InstallDaemonClient installs a daemon client to the Evaler.
func (ev *Evaler) InstallDaemonClient(client *daemon.Client) {
	ev.DaemonClient = client
}

// InstallModule installs a module to the Evaler so that it can be used with
// "use $name" from script.
func (ev *Evaler) InstallModule(name string, mod Ns) {
	ev.modules[name] = mod
}

// InstallBundled installs a bundled module to the Evaler.
func (ev *Evaler) InstallBundled(name, src string) {
	ev.bundled[name] = src
}

// SetArgs replaces the $args builtin variable with a vector built from the
// argument.
func (ev *Evaler) SetArgs(args []string) {
	v := vector.Empty
	for _, arg := range args {
		v = v.Cons(arg)
	}
	ev.Builtin["args"] = vars.NewRo(v)
}

// SetLibDir sets the library directory, in which external modules are to be
// found.
func (ev *Evaler) SetLibDir(libDir string) {
	ev.libDir = libDir
}

func searchPaths() []string {
	return strings.Split(os.Getenv("PATH"), ":")
}

// growPorts makes the size of ec.ports at least n, adding nil's if necessary.
func (fm *Frame) growPorts(n int) {
	if len(fm.ports) >= n {
		return
	}
	ports := fm.ports
	fm.ports = make([]*Port, n)
	copy(fm.ports, ports)
}

// EvalWithStdPorts sets up the Evaler with standard ports and evaluates an Op.
// The supplied name and text are used in diagnostic messages.
//
// TODO(xiaq): This function can only be used to evaluae an effectOp, and cannot
// be used to call functions with stdPorts. Make the Evaler initialize a
// stdPorts on construction, instead of in this function, so that NewTopFrame
// does not require the caller to supply the ports.
func (ev *Evaler) EvalWithStdPorts(op Op) error {
	stdPorts := newStdPorts(
		os.Stdin, os.Stdout, os.Stderr, ev.state.getValuePrefix())
	defer stdPorts.close()
	return ev.Eval(op, stdPorts.ports[:])
}

// Eval sets up the Evaler with the given ports and evaluates an Op. The
// supplied name and text are used in diagnostic messages.
//
// TODO(xiaq): This method only differs from eval in that it sets up intCh for
// relaying interrupts and puts Elvish in the foreground afterwards. Factor out
// those logics.
func (ev *Evaler) Eval(op Op, ports []*Port) error {
	// Set up intCh.
	stopSigGoroutine := make(chan struct{})
	sigGoRoutineDone := make(chan struct{})
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

	err := ev.eval(op, ports)

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

	return err
}

// eval evaluates a chunk node n. The supplied name and text are used in
// diagnostic messages.
func (ev *Evaler) eval(op Op, ports []*Port) error {
	ec := NewTopFrame(ev, op.Src, ports)
	return ec.Eval(op.Inner)
}

// Compile compiles Elvish code in the global scope. If the error is not nil, it
// always has type CompilationError.
func (ev *Evaler) Compile(n *parse.Chunk, src *Source) (Op, error) {
	return ev.CompileWithGlobal(n, src, ev.Global)
}

// CompileWithGlobal compiles Elvish code in an alternative global scope. If the
// error is not nil, it always has type CompilationError.
//
// TODO(xiaq): To use the Op created, the caller must create a Frame and mutate
// its local scope manually. Consider restructuring the API to make that
// unnecessary.
func (ev *Evaler) CompileWithGlobal(n *parse.Chunk, src *Source, g Ns) (Op, error) {
	return compile(ev.Builtin.static(), g.static(), n, src)
}

// EvalSource evaluates a chunk of Elvish source.
func (ev *Evaler) EvalSource(src *Source) error {
	n, err := parse.Parse(src.name, src.code)
	if err != nil {
		return err
	}
	op, err := ev.Compile(n, src)
	if err != nil {
		return err
	}
	return ev.EvalWithStdPorts(op)
}
