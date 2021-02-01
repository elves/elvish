// Package eval handles evaluation of parsed Elvish code and provides runtime
// facilities.
package eval

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"

	"github.com/xiaq/persistent/vector"
	"src.elv.sh/pkg/daemon"
	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/env"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/logutil"
	"src.elv.sh/pkg/parse"
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
	mu sync.RWMutex

	global, builtin *Ns

	deprecations deprecationRegistry

	// State of the module system.
	//
	// Library directory.
	libDir string
	// Internal modules are indexed by use specs. External modules are indexed by
	// absolute paths.
	modules map[string]*Ns

	// Various states and configs exposed to Elvish code.
	//
	// The prefix to prepend to value outputs when writing them to terminal,
	// exposed as $value-out-prefix.
	valuePrefix string
	// Whether to notify the success of background jobs, exposed as
	// $notify-bg-job-sucess.
	notifyBgJobSuccess bool
	// The current number of background jobs, exposed as $num-bg-jobs.
	numBgJobs int
	// Command-line arguments, exposed as $args.
	args vals.List
	// Chdir hooks, exposed indirectly as $before-chdir and $after-chdir.
	beforeChdir, afterChdir []func(string)

	// Dependencies.
	//
	// TODO: Remove these dependency by providing more general extension points.
	daemonClient daemon.Client
	editor       Editor
}

// Editor is the interface that the line editor has to satisfy. It is needed so
// that this package does not depend on the edit package.
type Editor interface {
	Notify(string, ...interface{})
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
	beforeChdirElvish, afterChdirElvish := vector.Empty, vector.Empty

	ev := &Evaler{
		global:  new(Ns),
		builtin: builtin,

		deprecations: newDeprecationRegistry(),

		modules: map[string]*Ns{"builtin": builtin},

		valuePrefix:        defaultValuePrefix,
		notifyBgJobSuccess: defaultNotifyBgJobSuccess,
		numBgJobs:          0,
		args:               vals.EmptyList,
	}

	ev.beforeChdir = []func(string){
		adaptChdirHook("before-chdir", ev, &beforeChdirElvish)}
	ev.afterChdir = []func(string){
		adaptChdirHook("after-chdir", ev, &afterChdirElvish)}

	moreBuiltins := NsBuilder{}.
		Add("pwd", NewPwdVar(ev)).
		Add("before-chdir", vars.FromPtr(&beforeChdirElvish)).
		Add("after-chdir", vars.FromPtr(&afterChdirElvish)).
		Add("value-out-indicator", vars.FromPtrWithMutex(
			&ev.valuePrefix, &ev.mu)).
		Add("notify-bg-job-success", vars.FromPtrWithMutex(
			&ev.notifyBgJobSuccess, &ev.mu)).
		Add("num-bg-jobs", vars.FromGet(func() interface{} {
			return strconv.Itoa(ev.getNumBgJobs())
		})).
		Add("args", vars.FromGet(func() interface{} {
			return ev.getArgs()
		})).
		Ns()
	builtin.slots = append(builtin.slots, moreBuiltins.slots...)
	builtin.names = append(builtin.names, moreBuiltins.names...)
	builtin.deleted = append(builtin.deleted, make([]bool, len(moreBuiltins.names))...)

	return ev
}

func adaptChdirHook(name string, ev *Evaler, pfns *vector.Vector) func(string) {
	return func(path string) {
		ports, cleanup := PortsFromStdFiles(ev.ValuePrefix())
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

// Access methods.

// Global returns the global Ns.
func (ev *Evaler) Global() *Ns {
	ev.mu.RLock()
	defer ev.mu.RUnlock()
	return ev.global
}

// AddGlobal merges the given *Ns into the global namespace.
func (ev *Evaler) AddGlobal(ns *Ns) {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	ev.global = CombineNs(ev.global, ns)
}

// Builtin returns the builtin Ns.
func (ev *Evaler) Builtin() *Ns {
	ev.mu.RLock()
	defer ev.mu.RUnlock()
	return ev.builtin
}

// AddBuiltin merges the given *Ns into the builtin namespace.
func (ev *Evaler) AddBuiltin(ns *Ns) {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	ev.builtin = CombineNs(ev.builtin, ns)
}

func (ev *Evaler) registerDeprecation(d deprecation) bool {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	return ev.deprecations.register(d)
}

// Returns libdir.
func (ev *Evaler) getLibDir() string {
	ev.mu.RLock()
	defer ev.mu.RUnlock()
	return ev.libDir
}

// SetLibDir sets the library directory for finding external modules.
func (ev *Evaler) SetLibDir(libDir string) {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	ev.libDir = libDir
}

// AddModule add an internal module so that it can be used with "use $name" from
// script.
func (ev *Evaler) AddModule(name string, mod *Ns) {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	ev.modules[name] = mod
}

// ValuePrefix returns the prefix to prepend to value outputs when writing them
// to terminal.
func (ev *Evaler) ValuePrefix() string {
	ev.mu.RLock()
	defer ev.mu.RUnlock()
	return ev.valuePrefix
}

func (ev *Evaler) getNotifyBgJobSuccess() bool {
	ev.mu.RLock()
	defer ev.mu.RUnlock()
	return ev.notifyBgJobSuccess
}

func (ev *Evaler) getNumBgJobs() int {
	ev.mu.RLock()
	defer ev.mu.RUnlock()
	return ev.numBgJobs
}

func (ev *Evaler) addNumBgJobs(delta int) {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	ev.numBgJobs += delta
}

func (ev *Evaler) getArgs() vals.List {
	ev.mu.RLock()
	defer ev.mu.RUnlock()
	return ev.args
}

// SetArgs sets the value of the $args variable to a list of strings, built from
// the given slice.
func (ev *Evaler) SetArgs(args []string) {
	v := listOfStrings(args)
	ev.mu.Lock()
	defer ev.mu.Unlock()
	ev.args = v
}

// Returns copies of beforeChdir and afterChdir.
func (ev *Evaler) chdirHooks() ([]func(string), []func(string)) {
	ev.mu.RLock()
	defer ev.mu.RUnlock()
	return append(([]func(string))(nil), ev.beforeChdir...),
		append(([]func(string))(nil), ev.afterChdir...)
}

// AddBeforeChdir adds a function to run before changing directory.
func (ev *Evaler) AddBeforeChdir(f func(string)) {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	ev.beforeChdir = append(ev.beforeChdir, f)
}

// AddAfterChdir adds a function to run after changing directory.
func (ev *Evaler) AddAfterChdir(f func(string)) {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	ev.afterChdir = append(ev.afterChdir, f)
}

// SetDaemonClient sets the daemon client associated with the Evaler.
func (ev *Evaler) SetDaemonClient(client daemon.Client) {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	ev.daemonClient = client
}

// DaemonClient returns the daemon client associated with the Evaler.
func (ev *Evaler) DaemonClient() daemon.Client {
	ev.mu.RLock()
	defer ev.mu.RUnlock()
	return ev.daemonClient
}

// DaemonClient returns the editor associated with the Evaler.
func (ev *Evaler) Editor() Editor {
	ev.mu.RLock()
	defer ev.mu.RUnlock()
	return ev.editor
}

// Chdir changes the current directory. On success it also updates the PWD
// environment variable and records the new directory in the directory history.
// It runs the functions in beforeChdir immediately before changing the
// directory, and the functions in afterChdir immediately after (if chdir was
// successful). It returns nil as long as the directory changing part succeeds.
func (ev *Evaler) Chdir(path string) error {
	beforeChdir, afterChdir := ev.chdirHooks()

	for _, hook := range beforeChdir {
		hook(path)
	}

	err := os.Chdir(path)
	if err != nil {
		return err
	}

	for _, hook := range afterChdir {
		hook(path)
	}

	pwd, err := os.Getwd()
	if err != nil {
		logger.Println("getwd after cd:", err)
		return nil
	}
	os.Setenv(env.PWD, pwd)

	return nil
}

// EvalCfg keeps configuration for the (*Evaler).Eval method.
type EvalCfg struct {
	// Ports to use in evaluation. The first 3 elements, if not specified
	// (either being nil or Ports containing fewer than 3 elements),
	// will be filled with DummyInputPort, DummyOutputPort and
	// DummyOutputPort respectively.
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

func (cfg *EvalCfg) fillDefaults() {
	if len(cfg.Ports) < 3 {
		cfg.Ports = append(cfg.Ports, make([]*Port, 3-len(cfg.Ports))...)
	}
	if cfg.Ports[0] == nil {
		cfg.Ports[0] = DummyInputPort
	}
	if cfg.Ports[1] == nil {
		cfg.Ports[1] = DummyOutputPort
	}
	if cfg.Ports[2] == nil {
		cfg.Ports[2] = DummyOutputPort
	}
}

// Eval evaluates a piece of source code with the given configuration. The
// returned error may be a parse error, compilation error or exception.
func (ev *Evaler) Eval(src parse.Source, cfg EvalCfg) error {
	cfg.fillDefaults()
	errFile := cfg.Ports[2].File

	tree, err := parse.Parse(src, parse.Config{WarningWriter: errFile})
	if err != nil {
		return err
	}

	ev.mu.Lock()
	b := ev.builtin
	defaultGlobal := cfg.Global == nil
	if defaultGlobal {
		// If cfg.Global is nil, use the Evaler's default global, and also
		// mutate the default global.
		cfg.Global = ev.global
		// Continue to hold the mutex; it will be released when ev.global gets
		// mutated.
	} else {
		ev.mu.Unlock()
	}

	op, err := compile(b.static(), cfg.Global.static(), tree, errFile)
	if err != nil {
		if defaultGlobal {
			ev.mu.Unlock()
		}
		return err
	}

	fm, cleanup := ev.prepareFrame(src, cfg)
	defer cleanup()

	newLocal, exec := op.prepare(fm)
	if defaultGlobal {
		ev.global = newLocal
		ev.mu.Unlock()
	}

	return exec()
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
	evalCfg.fillDefaults()
	if evalCfg.Global == nil {
		evalCfg.Global = ev.Global()
	}
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
	tree, parseErr := parse.Parse(src, parse.Config{WarningWriter: w})
	return parse.GetError(parseErr), ev.CheckTree(tree, w)
}

// CheckTree checks the given parsed source tree for compilation errors. If w is
// not nil, deprecation messages are written to it.
func (ev *Evaler) CheckTree(tree parse.Tree, w io.Writer) *diag.Error {
	_, compileErr := ev.compile(tree, ev.Global(), w)
	return GetCompilationError(compileErr)
}

// Compiles a parsed tree.
func (ev *Evaler) compile(tree parse.Tree, g *Ns, w io.Writer) (nsOp, error) {
	return compile(ev.Builtin().static(), g.static(), tree, w)
}
