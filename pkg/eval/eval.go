// Package eval handles evaluation of parsed Elvish code and provides runtime
// facilities.
package eval

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/elves/elvish/pkg/daemon"
	"github.com/elves/elvish/pkg/eval/bundled"
	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/eval/vars"
	"github.com/elves/elvish/pkg/parse"
	"github.com/elves/elvish/pkg/util"
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

	// State of the module system.
	libDir  string
	bundled map[string]string
	// Internal modules are indexed by use specs. External modules are indexed by
	// absolute paths.
	modules map[string]Ns

	// Dependencies.
	//
	// TODO: Remove these dependency by providing more general extension points.
	DaemonClient daemon.Client
	Editor       Editor
}

// EvalCfg keeps configuration for the (*Evaler).Eval method.
type EvalCfg struct {
	// Ports used in evaluation. Must contain at least 3 elements; the Eval
	// method panics otherwise.
	Ports []*Port
	// Callback to get a channel of interrupt signals and a function to call
	// when the channel is no longer needed.
	Interrupt func() (<-chan struct{}, func())
	// Whether the Eval method should try to put the Elvish in the foreground
	// after evaluating the Op.
	PutInFg bool
}

type evalerScopes struct {
	Global  Ns
	Builtin Ns
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
		ports, cleanup := portsFromFiles(
			[3]*os.File{os.Stdin, os.Stdout, os.Stderr}, ev.state.getValuePrefix())
		defer cleanup()
		for it := (*pfns).Iterator(); it.HasElem(); it.Next() {
			fn, ok := it.Elem().(Callable)
			if !ok {
				fmt.Fprintln(os.Stderr, name, "hook must be callable")
				continue
			}
			fm := NewTopFrame(ev,
				NewInternalGoSource("["+name+" hook]"), ports[:])
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
	ev.Builtin["args"] = vars.NewReadOnly(v)
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

// Eval evaluates an Op with the specified configuration.
func (ev *Evaler) Eval(op Op, cfg EvalCfg) error {
	if cfg.Ports == nil {
		cfg.Ports = DevNullPorts[:]
	}
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
func (ev *Evaler) ParseAndCompile(src *Source) (Op, error) {
	n, err := parse.AsChunk(src.Name, src.Code)
	if err != nil {
		return Op{}, err
	}
	return ev.Compile(n, src)
}

// Compile compiles Elvish code in the global scope. If the error is not nil, it
// can be passed to GetCompilationError to retrieve more details.
func (ev *Evaler) Compile(n *parse.Chunk, src *Source) (Op, error) {
	return ev.CompileWithGlobal(n, src, ev.Global)
}

// CompileWithGlobal compiles Elvish code in an alternative global scope. If the
// error is not nil, it can be passed to GetCompilationError to retrieve more
// details.
//
// TODO(xiaq): To use the Op created, the caller must create a Frame and mutate
// its local scope manually. Consider restructuring the API to make that
// unnecessary.
func (ev *Evaler) CompileWithGlobal(n *parse.Chunk, src *Source, g Ns) (Op, error) {
	return compile(ev.Builtin.static(), g.static(), n, src)
}
