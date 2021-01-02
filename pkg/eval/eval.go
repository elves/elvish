// Package eval handles evaluation of parsed Elvish code and provides runtime
// facilities.
package eval

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/elves/elvish/pkg/daemon"
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

// Evaler is used to evaluate elvish sources. It maintains runtime context
// shared among all evalCtx instances.
type Evaler struct {
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

// EvalCfg keeps configuration for the (*Evaler).Eval method.
type EvalCfg struct {
	// Ports to use in evaluation. If nil, equivalent to specifying
	// DevNullPorts[:]. If not nil, must contain at least 3 elements.
	Ports []*Port
	// If true, the code will be parsed and compiled but not executed.
	NoExecute bool
	// Callback to get a channel of interrupt signals and a function to call
	// when the channel is no longer needed.
	Interrupt func() (<-chan struct{}, func())
	// Whether the Eval method should try to put the Elvish in the foreground
	// after the code is executed.
	PutInFg bool
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
		for it := (*pfns).Iterator(); it.HasElem(); it.Next() {
			fn, ok := it.Elem().(Callable)
			if !ok {
				fmt.Fprintln(os.Stderr, name, "hook must be callable")
				continue
			}
			fm := NewTopFrame(ev, parse.Source{Name: "[hook " + name + "]"}, ports[:])
			err := fn.Call(fm, []interface{}{path}, NoOpts)
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

// growPorts makes the size of ec.ports at least n, adding nil's if necessary.
func (fm *Frame) growPorts(n int) {
	if len(fm.ports) >= n {
		return
	}
	ports := fm.ports
	fm.ports = make([]*Port, n)
	copy(fm.ports, ports)
}

// Eval evaluates a piece of source code with the given configuration. The
// returned error may be a parse error, compilation error or exception.
func (ev *Evaler) Eval(src parse.Source, cfg EvalCfg) error {
	if cfg.Ports == nil {
		cfg.Ports = DevNullPorts[:]
	}
	op, err := ev.ParseAndCompile(src, cfg.Ports[2].File)
	if err != nil {
		return err
	}
	if cfg.NoExecute {
		return nil
	} else {
		return ev.execOp(op, cfg)
	}
}

// Executes an Op with the specified configuration.
func (ev *Evaler) execOp(op Op, cfg EvalCfg) error {
	fm := NewTopFrame(ev, op.Src, cfg.Ports)
	if cfg.Interrupt != nil {
		intCh, cleanup := cfg.Interrupt()
		defer cleanup()
		fm.intCh = intCh
	}
	if cfg.PutInFg {
		defer func() {
			err := putSelfInFg()
			if err != nil {
				fmt.Fprintln(cfg.Ports[2].File,
					"failed to put myself in foreground:", err)
			}
		}()
	}
	return op.Inner.exec(fm)
}

// ParseAndCompile parses and compiles a Source.
func (ev *Evaler) ParseAndCompile(src parse.Source, w io.Writer) (Op, error) {
	tree, err := parse.ParseWithDeprecation(src, w)
	if err != nil {
		return Op{}, err
	}
	return ev.Compile(tree, w)
}

// Compile compiles Elvish code in the global scope. If the error is not nil, it
// can be passed to GetCompilationError to retrieve more details.
func (ev *Evaler) Compile(tree parse.Tree, w io.Writer) (Op, error) {
	return ev.CompileWithGlobal(tree, ev.Global, w)
}

// CompileWithGlobal compiles Elvish code in an alternative global scope. If the
// error is not nil, it can be passed to GetCompilationError to retrieve more
// details.
//
// TODO(xiaq): To use the Op created, the caller must create a Frame and mutate
// its local scope manually. Consider restructuring the API to make that
// unnecessary.
func (ev *Evaler) CompileWithGlobal(tree parse.Tree, g *Ns, w io.Writer) (Op, error) {
	return compile(ev.Builtin.static(), g.static(), tree, w)
}
