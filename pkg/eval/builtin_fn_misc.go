package eval

// Misc builtin functions.

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"path/filepath"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/parse"
)

// Builtins that have not been put into their own groups go here.

// TODO(xiaq): Document esleep.

func init() {
	addBuiltinFns(map[string]interface{}{
		"nop":        nop,
		"kind-of":    kindOf,
		"constantly": constantly,

		"resolve": resolve,

		"eval":    eval,
		"use-mod": useMod,
		"-source": source,

		// Time
		"esleep": sleep,
		"time":   timeCmd,

		"-ifaddrs": _ifaddrs,
	})

	// For rand and randint.
	rand.Seed(time.Now().UTC().UnixNano())
}

//elvdoc:fn nop
//
// ```elvish
// nop &any-opt= $value...
// ```
//
// Accepts arbitrary arguments and options and does exactly nothing.
//
// Examples:
//
// ```elvish-transcript
// ~> nop
// ~> nop a b c
// ~> nop &k=v
// ```
//
// Etymology: Various languages, in particular NOP in
// [assembly languages](https://en.wikipedia.org/wiki/NOP).

func nop(opts RawOptions, args ...interface{}) {
	// Do nothing
}

//elvdoc:fn kind-of
//
// ```elvish
// kind-of $value...
// ```
//
// Output the kinds of `$value`s. Example:
//
// ```elvish-transcript
// ~> kind-of lorem [] [&]
// ▶ string
// ▶ list
// ▶ map
// ```
//
// The terminology and definition of "kind" is subject to change.

func kindOf(fm *Frame, args ...interface{}) {
	out := fm.ports[1].Chan
	for _, a := range args {
		out <- vals.Kind(a)
	}
}

//elvdoc:fn constantly
//
// ```elvish
// constantly $value...
// ```
//
// Output a function that takes no arguments and outputs `$value`s when called.
// Examples:
//
// ```elvish-transcript
// ~> f=(constantly lorem ipsum)
// ~> $f
// ▶ lorem
// ▶ ipsum
// ```
//
// The above example is actually equivalent to simply `f = []{ put lorem ipsum }`;
// it is most useful when the argument is **not** a literal value, e.g.
//
// ```elvish-transcript
// ~> f = (constantly (uname))
// ~> $f
// ▶ Darwin
// ~> $f
// ▶ Darwin
// ```
//
// The above code only calls `uname` once, while if you do `f = []{ put (uname) }`,
// every time you invoke `$f`, `uname` will be called.
//
// Etymology: [Clojure](https://clojuredocs.org/clojure.core/constantly).

func constantly(args ...interface{}) Callable {
	// TODO(xiaq): Repr of this function is not right.
	return NewGoFn(
		"created by constantly",
		func(fm *Frame) {
			out := fm.ports[1].Chan
			for _, v := range args {
				out <- v
			}
		},
	)
}

//elvdoc:fn resolve
//
// ```elvish
// resolve $command
// ```
//
// Resolve `$command`. Command resolution is described in the
// [language reference](language.html). (TODO: actually describe it there.)
//
// Example:
//
// ```elvish-transcript
// ~> resolve echo
// ▶ <builtin echo>
// ~> fn f { }
// ~> resolve f
// ▶ <closure 0xc4201c24d0>
// ~> resolve cat
// ▶ <external cat>
// ```

func resolve(fm *Frame, head string) string {
	// Emulate static resolution of a command head. This needs to be kept in
	// sync with (*compiler).form.

	_, special := builtinSpecials[head]
	if special {
		return "special"
	}
	sigil, qname := SplitVariableRef(head)
	if sigil == "" && fm.ResolveVar(qname+FnSuffix) != nil {
		return "$" + qname + FnSuffix
	}
	return "(external " + parse.Quote(head) + ")"
}

//elvdoc:fn eval
//
// ```elvish
// eval $code &ns=$nil
// ```
//
// Evaluates `$code`, which should be a string. The evaluation happens in the
// namespace specified by the `&ns` option. If it is `$nil` (the default), a
// fresh empty namespace is created.
//
// If `$code` fails to parse or compile, the parse error or compilation error is
// raised as an exception.
//
// Examples:
//
// ```elvish-transcript
// ~> eval 'put x'
// ▶ x
// ~> ns = (ns [&x=initial])
// ~> eval 'put $x; x = altered; put $x' &ns=$ns
// ▶ initial
// ▶ altered
// ~> put $ns[x]
// ▶ altered
// ```
//
// NOTE: Unlike the `eval` found in many other dynamic languages, `eval` cannot
// affect the current namespace:
//
// ```elvish-transcript
// ~> eval 'x = value'
// ~> put $x
// compilation error: variable $x not found
// [tty 4], line 1: put $x
// ```

type evalOpts struct{ Ns Ns }

func (*evalOpts) SetDefaultOptions() {}

func eval(fm *Frame, opts evalOpts, code string) error {
	src := parse.Source{Name: fmt.Sprintf("[eval %d]", nextEvalCount()), Code: code}
	ns := opts.Ns
	if ns == nil {
		ns = make(Ns)
	}
	return evalInner(fm, src, ns, fm.traceback)
}

// Used to generate unique names for each source passed to eval.
var (
	evalCount      int
	evalCountMutex sync.Mutex
)

func nextEvalCount() int {
	evalCountMutex.Lock()
	defer evalCountMutex.Unlock()
	evalCount++
	return evalCount
}

//elvdoc:fn use-mod
//
// ```elvish
// use-mod $use-spec
// ```
//
// Imports a module, and outputs the namespace for the module.
//
// Most code should use the [use](language.html#importing-modules-with-use)
// special command instead.
//
// Examples:
//
// ```elvish-transcript
// ~> echo 'x = value' > a.elv
// ~> put (use-mod ./a)[x]
// ▶ value
// ```

func useMod(fm *Frame, spec string) (Ns, error) {
	return use(fm, spec, fm.traceback)
}

//elvdoc:fn -source
//
// ```elvish
// -source $filename
// ```
//
// Read the named file, and evaluate it in the current scope.
//
// This function is deprecated. Use [eval](#eval) instead.
//
// Examples:
//
// ```elvish-transcript
// ~> cat x.elv
// echo 'executing x.elv'
// foo = bar
// ~> -source x.elv
// executing x.elv
// ~> echo $foo
// bar
// ```
//
// Note that while in the example, you can reference `$foo` after sourcing `x.elv`,
// putting the `-source` command and reference to `$foo` in the **same code chunk**
// (e.g. by using <span class="key">Alt-Enter</span> to insert a literal Enter, or
// using `;`) is invalid:
//
// ```elvish-transcript
// ~> # A new Elvish session
// ~> cat x.elv
// echo 'executing x.elv'
// foo = bar
// ~> -source x.elv; echo $foo
// Compilation error: variable $foo not found
// [interactive], line 1:
// -source x.elv; echo $foo
// ```
//
// This is because the reading of the file is done in the evaluation phase, while
// the check for variables happens at the compilation phase (before evaluation). So
// the compiler has no evidence showing that `$foo` is actually valid, and will
// complain. (See [here](../learn/unique-semantics.html#execution-phases) for a
// more detailed description of execution phases.)
//
// To work around this, you can add a forward declaration for `$foo`:
//
// ```elvish-transcript
// ~> # Another new session
// ~> cat x.elv
// echo 'executing x.elv'
// foo = bar
// ~> foo = ''; -source x.elv; echo $foo
// executing x.elv
// bar
// ```

func source(fm *Frame, fname string) error {
	path, err := filepath.Abs(fname)
	if err != nil {
		return err
	}
	code, err := readFileUTF8(path)
	if err != nil {
		return err
	}
	src := parse.Source{Name: fname, Code: code, IsFile: true}
	tree, err := parse.ParseWithDeprecation(src, fm.ports[2].File)
	if err != nil {
		return err
	}
	scriptGlobal := fm.local.static()
	for name := range fm.up.static() {
		scriptGlobal.set(name)
	}
	op, err := compile(fm.Builtin.static(), scriptGlobal, tree, fm.ports[2].File)
	if err != nil {
		return err
	}
	return fm.Eval(op)
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

func sleep(fm *Frame, t float64) error {
	d := time.Duration(float64(time.Second) * t)
	select {
	case <-fm.Interrupts():
		return ErrInterrupted
	case <-time.After(d):
		return nil
	}
}

//elvdoc:fn time
//
// ```elvish
// time &on-end=$nil $callable
// ```
//
// Runs the callable, and call `$on-end` with the duration it took, as a
// number in seconds. If `$on-end` is `$nil` (the default), prints the
// duration in human-readable form.
//
// If `$callable` throws an exception, the exception is propagated after the
// on-end or default printing is done.
//
// If `$on-end` throws an exception, it is propagated, unless `$callable` has
// already thrown an exception.
//
// Example:
//
// ```elvish-transcript
// ~> time { sleep 1 }
// 1.006060647s
// ~> time { sleep 0.01 }
// 1.288977ms
// ~> t = ''
// ~> time &on-end=[x]{ t = $x } { sleep 1 }
// ~> put $t
// ▶ (float64 1.000925004)
// ~> time &on-end=[x]{ t = $x } { sleep 0.01 }
// ~> put $t
// ▶ (float64 0.011030208)
// ```

type timeOpt struct{ OnEnd Callable }

func (o *timeOpt) SetDefaultOptions() {}

func timeCmd(fm *Frame, opts timeOpt, f Callable) error {
	t0 := time.Now()
	err := f.Call(fm, NoArgs, NoOpts)
	t1 := time.Now()

	dt := t1.Sub(t0)
	if opts.OnEnd != nil {
		newFm := fm.fork("on-end callback of time")
		errCb := opts.OnEnd.Call(newFm, []interface{}{dt.Seconds()}, NoOpts)
		if err == nil {
			err = errCb
		}
	} else {
		fmt.Fprintln(fm.ports[1].File, dt)
	}

	return err
}

//elvdoc:fn -ifaddrs
//
// ```elvish
// -ifaddrs
// ```
//
// Output all IP addresses of the current host.
//
// This should be part of a networking module instead of the builtin module.

func _ifaddrs(fm *Frame) error {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return err
	}
	out := fm.ports[1].Chan
	for _, addr := range addrs {
		out <- addr.String()
	}
	return nil
}
