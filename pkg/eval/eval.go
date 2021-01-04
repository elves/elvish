// Package eval handles evaluation of parsed Elvish code and provides runtime
// facilities.
package eval

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"

	"github.com/elves/elvish/pkg/daemon"
	"github.com/elves/elvish/pkg/diag"
	"github.com/elves/elvish/pkg/eval/mods/bundled"
	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/eval/vars"
	"github.com/elves/elvish/pkg/logutil"
	"github.com/elves/elvish/pkg/parse"
	"github.com/xiaq/persistent/vector"
)

var logger = logutil.GetLogger("[eval] ")

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

// Evaler provides methods for evaluating code, and maintains state that is
// persisted between evaluation of different pieces of code. An Evaler is safe
// to use concurrently.
type Evaler struct {
	// All mutations to Evaler should be guarded by this mutex.
	//
	// Note that this is *not* a GIL; most state mutations when executing Elvish
	// code is localized and do not need to hold this mutex.
	//
	// TODO: Actually guard all mutations by this mutex.
	mu sync.Mutex

	evalerScopes

	state state

	// Chdir hooks.
	beforeChdir []func(string)
	afterChdir  []func(string)

	// State of the module system.
	libDir  string
	bundled map[string]string
	// Internal modules are indexed by use specs. External modules are indexed by
	// absolute paths.
	modules map[string]*Ns

	deprecations deprecationRegistry

	// Dependencies.
	//
	// TODO: Remove these dependency by providing more general extension points.
	DaemonClient daemon.Client
	Editor       Editor
}

type evalerScopes struct {
	Global  *Ns
	Builtin *Ns
}

//elvdoc:var after-chdir
//
// A list of functions to run after changing directory. These functions are always
// called with directory to change it, which might be a relative path. The
// following example also shows `$before-chdir`:
//
// ```elvish-transcript
// ~> before-chdir = [[dir]{ echo "Going to change to "$dir", pwd is "$pwd }]
// ~> after-chdir = [[dir]{ echo "Changed to "$dir", pwd is "$pwd }]
// ~> cd /usr
// Going to change to /usr, pwd is /Users/xiaq
// Changed to /usr, pwd is /usr
// /usr> cd local
// Going to change to local, pwd is /usr
// Changed to local, pwd is /usr/local
// /usr/local>
// ```
//
// @cf before-chdir

//elvdoc:var before-chdir
//
// A list of functions to run before changing directory. These functions are always
// called with the new working directory.
//
// @cf after-chdir

//elvdoc:var num-bg-jobs
//
// Number of background jobs.

//elvdoc:var notify-bg-job-success
//
// Whether to notify success of background jobs, defaulting to `$true`.
//
// Failures of background jobs are always notified.

//elvdoc:var value-out-indicator
//
// A string put before value outputs (such as those of of `put`). Defaults to
// `'▶ '`. Example:
//
// ```elvish-transcript
// ~> put lorem ipsum
// ▶ lorem
// ▶ ipsum
// ~> value-out-indicator = 'val> '
// ~> put lorem ipsum
// val> lorem
// val> ipsum
// ```
//
// Note that you almost always want some trailing whitespace for readability.

// NewEvaler creates a new Evaler.
func NewEvaler() *Evaler {
	builtin := builtinNs.Ns()

	ev := &Evaler{
		state: state{
			valuePrefix:        defaultValuePrefix,
			notifyBgJobSuccess: defaultNotifyBgJobSuccess,
			numBgJobs:          0,
		},
		evalerScopes: evalerScopes{
			Global:  new(Ns),
			Builtin: builtin,
		},
		modules: map[string]*Ns{
			"builtin": builtin,
		},
		bundled: bundled.Get(),

		deprecations: newDeprecationRegistry(),
	}

	beforeChdirElvish, afterChdirElvish := vector.Empty, vector.Empty
	ev.beforeChdir = append(ev.beforeChdir,
		adaptChdirHook("before-chdir", ev, &beforeChdirElvish))
	ev.afterChdir = append(ev.afterChdir,
		adaptChdirHook("after-chdir", ev, &afterChdirElvish))

	moreBuiltinsBuilder := make(NsBuilder)
	moreBuiltinsBuilder["before-chdir"] = vars.FromPtr(&beforeChdirElvish)
	moreBuiltinsBuilder["after-chdir"] = vars.FromPtr(&afterChdirElvish)

	moreBuiltinsBuilder["value-out-indicator"] = vars.FromPtrWithMutex(
		&ev.state.valuePrefix, &ev.state.mutex)
	moreBuiltinsBuilder["notify-bg-job-success"] = vars.FromPtrWithMutex(
		&ev.state.notifyBgJobSuccess, &ev.state.mutex)
	moreBuiltinsBuilder["num-bg-jobs"] = vars.FromGet(func() interface{} {
		return strconv.Itoa(ev.state.getNumBgJobs())
	})
	moreBuiltinsBuilder["pwd"] = NewPwdVar(ev)

	moreBuiltins := moreBuiltinsBuilder.Ns()
	builtin.slots = append(builtin.slots, moreBuiltins.slots...)
	builtin.names = append(builtin.names, moreBuiltins.names...)

	return ev
}

func adaptChdirHook(name string, ev *Evaler, pfns *vector.Vector) func(string) {
	return func(path string) {
		ports, cleanup := portsFromFiles(
			[3]*os.File{os.Stdin, os.Stdout, os.Stderr}, ev.state.getValuePrefix())
		defer cleanup()
		callCfg := CallCfg{Args: []interface{}{path}, From: "[hook " + name + "]"}
		evalCfg := EvalCfg{Ports: ports[:]}
		for it := (*pfns).Iterator(); it.HasElem(); it.Next() {
			fn, ok := it.Elem().(Callable)
			if !ok {
				fmt.Fprintln(os.Stderr, name, "hook must be callable")
				continue
			}
			err := ev.Call(fn, callCfg, evalCfg)
			if err != nil {
				// TODO: Stack trace
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}
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
func (ev *Evaler) InstallDaemonClient(client daemon.Client) {
	ev.DaemonClient = client
}

// InstallModule installs a module to the Evaler so that it can be used with
// "use $name" from script.
func (ev *Evaler) InstallModule(name string, mod *Ns) {
	ev.modules[name] = mod
}

// SetArgs replaces the $args builtin variable with a vector built from the
// argument.
func (ev *Evaler) SetArgs(args []string) {
	v := vector.Empty
	for _, arg := range args {
		v = v.Cons(arg)
	}
	// TODO: Avoid creating the variable dynamically; instead, always create the
	// variable, and set its value here.
	ev.Builtin.slots = append(ev.Builtin.slots, vars.NewReadOnly(v))
	ev.Builtin.names = append(ev.Builtin.names, "args")
}

// SetLibDir sets the library directory, in which external modules are to be
// found.
func (ev *Evaler) SetLibDir(libDir string) {
	ev.libDir = libDir
}

func (ev *Evaler) registerDeprecation(d deprecation) bool {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	return ev.deprecations.register(d)
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

// EvalCfg keeps configuration for the (*Evaler).Eval method.
type EvalCfg struct {
	// Ports to use in evaluation. The first 3 elements, if not specified
	// (either being nil or Ports containing fewer than 3 elements),
	// will be filled with DevNullClosedChan, DevNullBlackholeChan and
	// DevNullBlackholeChan respectively.
	Ports []*Port
	// Callback to get a channel of interrupt signals and a function to call
	// when the channel is no longer needed.
	Interrupt func() (<-chan struct{}, func())
	// Whether the Eval method should try to put the Elvish in the foreground
	// after the code is executed.
	PutInFg bool
	// If not nil, used the given global namespace, instead of Evaler's own.
	Global *Ns
}

func (cfg *EvalCfg) fillDefaults(ev *Evaler) {
	if len(cfg.Ports) < 3 {
		cfg.Ports = append(cfg.Ports, make([]*Port, 3-len(cfg.Ports))...)
	}
	if cfg.Ports[0] == nil {
		cfg.Ports[0] = DevNullClosedChan
	}
	if cfg.Ports[1] == nil {
		cfg.Ports[1] = DevNullBlackholeChan
	}
	if cfg.Ports[2] == nil {
		cfg.Ports[2] = DevNullBlackholeChan
	}

	if cfg.Global == nil {
		cfg.Global = ev.Global
	}
}

// Eval evaluates a piece of source code with the given configuration. The
// returned error may be a parse error, compilation error or exception.
func (ev *Evaler) Eval(src parse.Source, cfg EvalCfg) error {
	cfg.fillDefaults(ev)
	errFile := cfg.Ports[2].File

	tree, err := parse.ParseWithDeprecation(src, errFile)
	if err != nil {
		return err
	}
	op, err := ev.compile(tree, cfg.Global, errFile)
	if err != nil {
		return err
	}
	fm, cleanup := ev.prepareFrame(src, cfg)
	defer cleanup()
	return op.exec(fm)
}

// CallCfg keeps configuration for the (*Evaler).Call method.
type CallCfg struct {
	// Arguments to pass to the the function.
	Args []interface{}
	// Options to pass to the function.
	Opts map[string]interface{}
	// The name of the internal source that is calling the function.
	From string
}

func (cfg *CallCfg) fillDefaults() {
	if cfg.Opts == nil {
		cfg.Opts = NoOpts
	}
	if cfg.From == "" {
		cfg.From = "[internal]"
	}
}

// Call calls a given function.
func (ev *Evaler) Call(f Callable, callCfg CallCfg, evalCfg EvalCfg) error {
	callCfg.fillDefaults()
	evalCfg.fillDefaults(ev)
	fm, cleanup := ev.prepareFrame(parse.Source{Name: callCfg.From}, evalCfg)
	defer cleanup()
	return f.Call(fm, callCfg.Args, callCfg.Opts)
}

func (ev *Evaler) prepareFrame(src parse.Source, cfg EvalCfg) (*Frame, func()) {
	var intCh <-chan struct{}
	var intChCleanup func()
	if cfg.Interrupt != nil {
		intCh, intChCleanup = cfg.Interrupt()
	}

	fm := &Frame{ev, src, cfg.Global, new(Ns), intCh, cfg.Ports, nil, false}
	return fm, func() {
		if intChCleanup != nil {
			intChCleanup()
		}
		if cfg.PutInFg {
			err := putSelfInFg()
			if err != nil {
				fmt.Fprintln(cfg.Ports[2].File,
					"failed to put myself in foreground:", err)
			}
		}
	}
}

// Check checks the given source code for any parse error and compilation error.
// It always tries to compile the code even if there is a parse error; both
// return values may be non-nil. If w is not nil, deprecation messages are
// written to it.
func (ev *Evaler) Check(src parse.Source, w io.Writer) (*parse.Error, *diag.Error) {
	tree, parseErr := parse.ParseWithDeprecation(src, w)
	return parse.GetError(parseErr), ev.CheckTree(tree, w)
}

// CheckTree checks the given parsed source tree for compilation errors. If w is
// not nil, deprecation messages are written to it.
func (ev *Evaler) CheckTree(tree parse.Tree, w io.Writer) *diag.Error {
	_, compileErr := ev.compile(tree, ev.Global, w)
	return GetCompilationError(compileErr)
}

// Compiles a parsed tree.
func (ev *Evaler) compile(tree parse.Tree, g *Ns, w io.Writer) (effectOp, error) {
	return compile(ev.Builtin.static(), g.static(), tree, w)
}
