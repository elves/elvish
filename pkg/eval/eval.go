// Package eval handles evaluation of parsed Elvish code and provides runtime
// facilities.
package eval

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"

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
)

// Evaler provides methods for evaluating code, and maintains state that is
// persisted between evaluation of different pieces of code. An Evaler is safe
// to use concurrently.
type Evaler struct {
	// The following fields must only be set before the Evaler is used to
	// evaluate any code; mutating them afterwards may cause race conditions.

	// Command-line arguments, exposed as $args.
	Args vals.List
	// Hooks to run before exit or exec.
	PreExitHooks []func()
	// Chdir hooks, exposed indirectly as $before-chdir and $after-chdir.
	BeforeChdir, AfterChdir []func(string)
	// Directories to search libraries.
	LibDirs []string
	// Source code of internal bundled modules indexed by use specs.
	BundledModules map[string]string
	// Callback to notify the success or failure of background jobs. Must not be
	// mutated once the Evaler is used to evaluate any code.
	BgJobNotify func(string)
	// Path to the rc file, and path to the rc file actually evaluated. These
	// are not used by the Evaler itself right now; they are here so that they
	// can be exposed to the runtime: module.
	RcPath, EffectiveRcPath string

	mu sync.RWMutex
	// Mutations to fields below must be guarded by mutex.
	//
	// Note that this is *not* a GIL; most state mutations when executing Elvish
	// code is localized and do not need to hold this mutex.
	//
	// TODO: Actually guard all mutations by this mutex.

	global, builtin *Ns

	deprecations deprecationRegistry

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
}

// NewEvaler creates a new Evaler.
func NewEvaler() *Evaler {
	builtin := builtinNs.Ns()

	newListVar := func(l vals.List) vars.PtrVar { return vars.FromPtr(&l) }
	beforeExitHookElvish := newListVar(vals.EmptyList)
	beforeChdirElvish := newListVar(vals.EmptyList)
	afterChdirElvish := newListVar(vals.EmptyList)

	ev := &Evaler{
		global:  new(Ns),
		builtin: builtin,

		deprecations: newDeprecationRegistry(),

		modules:        make(map[string]*Ns),
		BundledModules: make(map[string]string),

		valuePrefix:        defaultValuePrefix,
		notifyBgJobSuccess: defaultNotifyBgJobSuccess,
		numBgJobs:          0,
		Args:               vals.EmptyList,
	}

	ev.PreExitHooks = []func(){func() {
		CallHook(ev, nil, "before-exit", beforeExitHookElvish.Get().(vals.List))
	}}
	ev.BeforeChdir = []func(string){func(path string) {
		CallHook(ev, nil, "before-chdir", beforeChdirElvish.Get().(vals.List), path)
	}}
	ev.AfterChdir = []func(string){func(path string) {
		CallHook(ev, nil, "after-chdir", afterChdirElvish.Get().(vals.List), path)
	}}

	ev.ExtendBuiltin(BuildNs().
		AddVar("pwd", NewPwdVar(ev)).
		AddVar("before-exit", beforeExitHookElvish).
		AddVar("before-chdir", beforeChdirElvish).
		AddVar("after-chdir", afterChdirElvish).
		AddVar("value-out-indicator",
			vars.FromPtrWithMutex(&ev.valuePrefix, &ev.mu)).
		AddVar("notify-bg-job-success",
			vars.FromPtrWithMutex(&ev.notifyBgJobSuccess, &ev.mu)).
		AddVar("num-bg-jobs",
			vars.FromGet(func() any { return strconv.Itoa(ev.getNumBgJobs()) })).
		AddVar("args", vars.FromGet(func() any { return ev.Args })))

	// Install the "builtin" module after extension is complete.
	ev.modules["builtin"] = ev.builtin

	return ev
}

// PreExit runs all pre-exit hooks.
func (ev *Evaler) PreExit() {
	for _, hook := range ev.PreExitHooks {
		hook()
	}
}

// Access methods.

// Global returns the global Ns.
func (ev *Evaler) Global() *Ns {
	ev.mu.RLock()
	defer ev.mu.RUnlock()
	return ev.global
}

// ExtendGlobal extends the global namespace with the given namespace.
func (ev *Evaler) ExtendGlobal(ns Nser) {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	ev.global = CombineNs(ev.global, ns.Ns())
}

// DeleteFromGlobal deletes names from the global namespace.
func (ev *Evaler) DeleteFromGlobal(names map[string]struct{}) {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	g := ev.global.clone()
	for i := range g.infos {
		if _, ok := names[g.infos[i].name]; ok {
			g.infos[i].deleted = true
		}
	}
	ev.global = g
}

// Builtin returns the builtin Ns.
func (ev *Evaler) Builtin() *Ns {
	ev.mu.RLock()
	defer ev.mu.RUnlock()
	return ev.builtin
}

// ExtendBuiltin extends the builtin namespace with the given namespace.
func (ev *Evaler) ExtendBuiltin(ns Nser) {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	ev.builtin = CombineNs(ev.builtin, ns.Ns())
}

// ReplaceBuiltin replaces the builtin namespace. It should only be used in
// tests.
func (ev *Evaler) ReplaceBuiltin(ns *Ns) {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	ev.builtin = ns
}

func (ev *Evaler) registerDeprecation(d deprecation) bool {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	return ev.deprecations.register(d)
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

// Chdir changes the current directory, and updates $E:PWD on success
//
// It runs the functions in beforeChdir immediately before changing the
// directory, and the functions in afterChdir immediately after (if chdir was
// successful). It returns nil as long as the directory changing part succeeds.
func (ev *Evaler) Chdir(path string) error {
	for _, hook := range ev.BeforeChdir {
		hook(path)
	}

	err := os.Chdir(path)
	if err != nil {
		return err
	}

	for _, hook := range ev.AfterChdir {
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
	// Context that can be used to cancel the evaluation.
	Interrupts context.Context
	// Ports to use in evaluation. The first 3 elements, if not specified
	// (either being nil or Ports containing fewer than 3 elements),
	// will be filled with DummyInputPort, DummyOutputPort and
	// DummyOutputPort respectively.
	Ports []*Port
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

	op, _, err := compile(b.static(), cfg.Global.static(), nil, tree, errFile)
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
	// Arguments to pass to the function.
	Args []any
	// Options to pass to the function.
	Opts map[string]any
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
	intCtx := cfg.Interrupts
	if intCtx == nil {
		intCtx = context.Background()
	}

	ports := fillDefaultDummyPorts(cfg.Ports)

	fm := &Frame{ev, src, cfg.Global, new(Ns), nil, intCtx, ports, nil, false}
	return fm, func() {
		if cfg.PutInFg {
			err := putSelfInFg()
			if err != nil {
				fmt.Fprintln(ports[2].File,
					"failed to put myself in foreground:", err)
			}
		}
	}
}

func fillDefaultDummyPorts(ports []*Port) []*Port {
	growPorts(&ports, 3)
	if ports[0] == nil {
		ports[0] = DummyInputPort
	}
	if ports[1] == nil {
		ports[1] = DummyOutputPort
	}
	if ports[2] == nil {
		ports[2] = DummyOutputPort
	}
	return ports
}

// Check checks the given source code for any parse error, autofixes, and
// compilation error. It always tries to compile the code even if there is a
// parse error. If w is not nil, deprecation messages are written to it.
func (ev *Evaler) Check(src parse.Source, w io.Writer) (error, []string, error) {
	tree, parseErr := parse.Parse(src, parse.Config{WarningWriter: w})
	autofixes, compileErr := ev.CheckTree(tree, w)
	return parseErr, autofixes, compileErr
}

// CheckTree checks the given parsed source tree for autofixes and compilation
// errors. If w is not nil, deprecation messages are written to it.
func (ev *Evaler) CheckTree(tree parse.Tree, w io.Writer) ([]string, error) {
	ev.mu.RLock()
	b, g, m := ev.builtin, ev.global, ev.modules
	ev.mu.RUnlock()
	_, autofixes, compileErr := compile(b.static(), g.static(), mapKeys(m), tree, w)
	return autofixes, compileErr
}
