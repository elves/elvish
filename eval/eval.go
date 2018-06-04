// Package eval handles evaluation of parsed Elvish code and provides runtime
// facilities.
package eval

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/elves/elvish/daemon"
	"github.com/elves/elvish/eval/bundled"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/sys"
	"github.com/elves/elvish/util"
	"github.com/xiaq/persistent/hashmap"
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
	defaultNumBgJobs          = 0
	initIndent                = vals.NoPretty
)

// Evaler is used to evaluate elvish sources. It maintains runtime context
// shared among all evalCtx instances.
type Evaler struct {
	evalerScopes
	valuePrefix        string
	notifyBgJobSuccess bool
	numBgJobs          int
	beforeChdir        []func(string)
	afterChdir         []func(string)
	DaemonClient       *daemon.Client
	modules            map[string]Ns
	// bundled modules
	bundled map[string]string
	Editor  Editor
	libDir  string
	intCh   chan struct{}
}

type evalerScopes struct {
	Global  Ns
	Builtin Ns
}

// NewEvaler creates a new Evaler.
func NewEvaler() *Evaler {
	builtin := builtinNs.Clone()

	ev := &Evaler{
		valuePrefix:        defaultValuePrefix,
		notifyBgJobSuccess: defaultNotifyBgJobSuccess,
		numBgJobs:          defaultNumBgJobs,
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

	builtin["value-out-indicator"] = vars.FromPtr(&ev.valuePrefix)
	builtin["notify-bg-job-success"] = vars.FromPtr(&ev.notifyBgJobSuccess)
	builtin["num-bg-jobs"] = vars.FromPtr(&ev.numBgJobs)
	builtin["pwd"] = PwdVariable{ev}

	return ev
}

func adaptChdirHook(name string, ev *Evaler, pfns *vector.Vector) func(string) {
	return func(path string) {
		stdPorts := newStdPorts(os.Stdin, os.Stdout, os.Stderr, ev.valuePrefix)
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

// SetArgs sets the $args builtin variable.
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
func (ev *Evaler) EvalWithStdPorts(op Op, src *Source) error {
	stdPorts := newStdPorts(os.Stdin, os.Stdout, os.Stderr, ev.valuePrefix)
	defer stdPorts.close()
	return ev.Eval(op, stdPorts.ports[:], src)
}

// Eval sets up the Evaler with the given ports and evaluates an Op.
// The supplied name and text are used in diagnostic messages.
func (ev *Evaler) Eval(op Op, ports []*Port, src *Source) error {
	// Ignore TTOU.
	//
	// When a subprocess in its own process group puts itself in the foreground,
	// Elvish will be put in the background. When the code finishes execution,
	// Elvish will attempt to move itself back to the foreground by calling
	// tcsetpgrp. However, whenever a background process calls tcsetpgrp (or
	// otherwise attempts to modify the terminal configuration), TTOU will be
	// sent, whose default handler is to stop the process. Or, if the process
	// lives in an orphaned process group (which is often the case for Elvish),
	// the call will outright fail. Therefore, for Elvish to be able to move
	// itself back to the foreground later, we need to ignore TTOU now.
	ignoreTTOU()
	defer unignoreTTOU()

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

	err := ev.eval(op, ports, src)

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
func (ev *Evaler) eval(op Op, ports []*Port, src *Source) error {
	ec := NewTopFrame(ev, src, ports)
	return ec.Eval(op)
}

// Compile compiles Elvish code in the global scope. If the error is not nil, it
// always has type CompilationError.
func (ev *Evaler) Compile(n *parse.Chunk, src *Source) (Op, error) {
	return compile(ev.Builtin.static(), ev.Global.static(), n, src)
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
	return ev.EvalWithStdPorts(op, src)
}

// SourceRC evaluates a rc.elv file. It evaluates the file in the global
// namespace. If the file defines a $-exports- map variable, the variable is
// removed and its content is poured into the global namespace, not overriding
// existing variables.
func (ev *Evaler) SourceRC(src *Source) error {
	n, err := parse.Parse(src.name, src.code)
	if err != nil {
		return err
	}
	op, err := ev.Compile(n, src)
	if err != nil {
		return err
	}
	errEval := ev.EvalWithStdPorts(op, src)
	var errExports error
	if ev.Global.HasName("-exports-") {
		exports := ev.Global.PopName("-exports-").Get()
		switch exports := exports.(type) {
		case hashmap.Map:
			for it := exports.Iterator(); it.HasElem(); it.Next() {
				k, v := it.Elem()
				if name, ok := k.(string); ok {
					if !ev.Global.HasName(name) {
						ev.Global.Add(name, vars.NewAnyWithInit(v))
					}
				} else {
					errKey := fmt.Errorf("non-string key in $-exports-: %s",
						vals.Repr(k, vals.NoPretty))
					errExports = util.Errors(errExports, errKey)
				}
			}
		default:
			errExports = fmt.Errorf(
				"$-exports- should be a map, got %s", vals.Kind(exports))
		}
	}
	return util.Errors(errEval, errExports)
}
