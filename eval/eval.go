// Package eval handles evaluation of parsed Elvish code and provides runtime
// facilities.
package eval

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"unicode/utf8"

	"github.com/elves/elvish/daemon"
	"github.com/elves/elvish/eval/bundled"
	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/eval/vartypes"
	"github.com/elves/elvish/parse"
	"github.com/elves/elvish/sys"
	"github.com/elves/elvish/util"
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
	defaultValueOutIndicator = "▶ "
	initIndent               = types.NoPretty
)

// Evaler is used to evaluate elvish sources. It maintains runtime context
// shared among all evalCtx instances.
type Evaler struct {
	evalerScopes
	evalerPorts
	DaemonClient *daemon.Client
	modules      map[string]Ns
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
	builtin := makeBuiltinNs()

	ev := &Evaler{
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

	valueOutIndicator := defaultValueOutIndicator
	ev.evalerPorts = newEvalerPorts(os.Stdin, os.Stdout, os.Stderr, &valueOutIndicator)
	builtin["value-out-indicator"] = vartypes.NewString(&valueOutIndicator)

	return ev
}

// Close releases resources allocated when creating this Evaler.
func (ev *Evaler) Close() {
	ev.evalerPorts.close()
}

// InstallDaemonClient installs a daemon client to the Evaler.
func (ev *Evaler) InstallDaemonClient(client *daemon.Client) {
	ev.DaemonClient = client
	// XXX This is really brittle
	ev.Builtin["pwd"] = PwdVariable{client}
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

// SetLibDir sets the library directory, in which external modules are to be
// found.
func (ev *Evaler) SetLibDir(libDir string) {
	ev.libDir = libDir
}

func searchPaths() []string {
	return strings.Split(os.Getenv("PATH"), ":")
}

// growPorts makes the size of ec.ports at least n, adding nil's if necessary.
func (ec *Frame) growPorts(n int) {
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
	ec := NewTopFrame(ev, name, text, ports)
	return ec.PEval(op)
}

// Eval sets up the Evaler with standard ports and evaluates an Op. The supplied
// name and text are used in diagnostic messages.
func (ev *Evaler) Eval(op Op, name, text string) error {
	return ev.EvalWithPorts(ev.ports[:], op, name, text)
}

// EvalWithPorts sets up the Evaler with the given ports and evaluates an Op.
// The supplied name and text are used in diagnostic messages.
func (ev *Evaler) EvalWithPorts(ports []*Port, op Op, name, text string) error {
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

	return err
}

// Compile compiles elvish code in the global scope. If the error is not nil, it
// always has type CompilationError.
func (ev *Evaler) Compile(n *parse.Chunk, name, text string) (Op, error) {
	return compile(ev.Builtin.static(), ev.Global.static(), n, name, text)
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
