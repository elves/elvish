package eval

// Misc builtin functions.

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"path/filepath"
	"time"
	"unicode/utf8"

	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/parse"
)

// Builtins that have not been put into their own groups go here.

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

//elvdoc:fn -source
//
// ```elvish
// -source $filename
// ```
//
// Read the named file, and evaluate it in the current scope.
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

// TODO(xiaq): Document esleep.

//elvdoc:fn -time
//
// ```elvish
// -time $callable
// ```
//
// Run the callable, and write the time used to run it. Example:
//
// ```elvish-transcript
// ~> -time { sleep 1 }
// 1.006060647s
// ```
//
// When the callable also produces outputs, they are a bit tricky to separate from
// the output of `-time`. The easiest workaround is to redirect the output into a
// temporary file:
//
// ```elvish-transcript
// ~> f = (mktemp)
// ~> -time { { echo output; sleep 1 } > $f }
// 1.005589823s
// ~> cat $f
// output
// ~> rm $f
// ```

//elvdoc:fn -ifaddrs
//
// ```elvish
// -ifaddrs
// ```
//
// Output all IP addresses of the current host.
//
// This should be part of a networking module instead of the builtin module.

func init() {
	addBuiltinFns(map[string]interface{}{
		"nop":        nop,
		"kind-of":    kindOf,
		"constantly": constantly,

		"resolve": resolve,

		"-source": source,

		// Time
		"esleep": sleep,
		"-time":  _time,

		"-ifaddrs": _ifaddrs,
	})

	// For rand and randint.
	rand.Seed(time.Now().UTC().UnixNano())
}

func nop(opts RawOptions, args ...interface{}) {
	// Do nothing
}

func kindOf(fm *Frame, args ...interface{}) {
	out := fm.ports[1].Chan
	for _, a := range args {
		out <- vals.Kind(a)
	}
}

func constantly(args ...interface{}) Callable {
	// XXX Repr of this fn is not right
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

func source(fm *Frame, fname string) error {
	path, err := filepath.Abs(fname)
	if err != nil {
		return err
	}
	code, err := readFileUTF8(path)
	if err != nil {
		return err
	}
	n, err := parse.AsChunk(fname, code)
	if err != nil {
		return err
	}
	scriptGlobal := fm.local.static()
	for name := range fm.up.static() {
		scriptGlobal.set(name)
	}
	op, err := compile(fm.Builtin.static(),
		scriptGlobal, n, NewScriptSource(path, code))
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

func _time(fm *Frame, f Callable) error {
	t0 := time.Now()
	err := f.Call(fm, NoArgs, NoOpts)
	t1 := time.Now()

	dt := t1.Sub(t0)
	fmt.Fprintln(fm.ports[1].File, dt)

	return err
}

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
