package eval

// Misc builtin functions.

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/eval/vars"
	"github.com/elves/elvish/pkg/parse"
)

// Builtins that have not been put into their own groups go here.

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
		"sleep":  sleep,
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
	out := fm.OutputChan()
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
			out := fm.OutputChan()
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
// Output what `$command` resolves to in symbolic form. Command resolution is
// described in the [language reference](language.html#ordinary-command).
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
	special, fnRef := resolveCmdHeadInternally(fm, head, nil)
	switch {
	case special != nil:
		return "special"
	case fnRef != nil:
		return "$" + head + FnSuffix
	default:
		return "(external " + parse.Quote(head) + ")"
	}
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

type evalOpts struct{ Ns *Ns }

func (*evalOpts) SetDefaultOptions() {}

func eval(fm *Frame, opts evalOpts, code string) error {
	src := parse.Source{Name: fmt.Sprintf("[eval %d]", nextEvalCount()), Code: code}
	ns := opts.Ns
	if ns == nil {
		ns = new(Ns)
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

func useMod(fm *Frame, spec string) (*Ns, error) {
	return use(fm, spec, fm.traceback)
}

//elvdoc:fn -source
//
// ```elvish
// -source $filename
// ```
//
// Read the named file, and evaluate it in a temporary namespace built from the
// current local and up scope.
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
// Since the file is evaluated in a temporary namespace, any modifications to
// the namespace itself - creation of variables and deletion of variables - do
// not affect the code calling `-source`. For example:
//
// ```elvish-transcript
// ~> echo 'foo = lorem' > a.elv
// ~> -source a.elv
// ~> put $foo
// compilation error: 4-8 in [tty]: variable $foo not found
// compilation error: variable $foo not found
// [tty 3], line 1: put $foo
// ```
//
// However, the file may mutate variables that already exist, and such mutations
// are persisted:
//
// ```elvish-transcript
// ~> foo = lorem
// ~> echo 'foo = ipsum' > a.elv
// ~> -source a.elv
// ~> put $foo
// ▶ ipsum
// ```

func source(fm *Frame, fname string) error {
	code, err := readFileUTF8(fname)
	if err != nil {
		return err
	}
	src := parse.Source{Name: fname, Code: code, IsFile: true}
	tree, err := parse.ParseWithDeprecation(src, fm.ErrorFile())
	if err != nil {
		return err
	}
	// Amalgamate the up and local scope into a new scope to use as the global
	// scope to evaluate the code in.
	g := amalgamateNs(fm.local, fm.up)
	op, err := compile(fm.Builtin.static(), g.static(), tree, fm.ErrorFile())
	if err != nil {
		return err
	}
	newFm := fm.fork("[-source]")
	newFm.local = g
	newFm.srcMeta = src
	return op.Exec(newFm)
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

func amalgamateNs(local, up *Ns) *Ns {
	slots := append([]vars.Var(nil), local.slots...)
	names := append([]string(nil), local.names...)
	for i := range up.slots {
		if local.lookup(up.names[i]) == -1 {
			slots = append(slots, up.slots[i])
			names = append(names, up.names[i])
		}
	}
	return &Ns{slots, names}
}

// TimeAfter is used by the sleep command to obtain a channel that is delivered
// a value after the specified time.
//
// It is a variable to allow for unit tests to efficiently test the behavior of
// the `sleep` command, both by eliminating an actual sleep and verifying the
// duration was properly parsed.
var TimeAfter = func(fm *Frame, d time.Duration) <-chan time.Time {
	return time.After(d)
}

//elvdoc:fn sleep
//
// ```elvish
// sleep $duration
// ```
//
// Pauses for at least the specified duration. The actual pause duration depends
// on the system.
//
// This only affects the current Elvish context. It does not affect any other
// contexts that might be executing in parallel as a consequence of a command
// such as [`peach`](#peach).
//
// A duration can be a simple [number](../language.html#number) (with optional
// fractional value) without an explicit unit suffix, with an implicit unit of
// seconds.
//
// A duration can also be a string written as a sequence of decimal numbers,
// each with optional fraction, plus a unit suffix. For example, "300ms",
// "1.5h" or "1h45m7s". Valid time units are "ns", "us" (or "µs"), "ms", "s",
// "m", "h".
//
// Passing a negative duration causes an exception; this is different from the
// typical BSD or GNU `sleep` command that silently exits with a success status
// without pausing when given a negative duration.
//
// See the [Go documentation](https://golang.org/pkg/time/#ParseDuration) for
// more information about how durations are parsed.
//
// Examples:
//
// ```elvish-transcript
// ~> sleep 0.1    # sleeps 0.1 seconds
// ~> sleep 100ms  # sleeps 0.1 seconds
// ~> sleep 1.5m   # sleeps 1.5 minutes
// ~> sleep 1m30s  # sleeps 1.5 minutes
// ~> sleep -1
// Exception: sleep duration must be >= zero
// [tty 8], line 1: sleep -1
// ```

func sleep(fm *Frame, duration interface{}) error {
	var d time.Duration

	switch duration := duration.(type) {
	case float64:
		d = time.Duration(float64(time.Second) * duration)
	case string:
		f, err := strconv.ParseFloat(duration, 64)
		if err == nil { // it's a simple number assumed to have units == seconds
			d = time.Duration(float64(time.Second) * f)
		} else {
			d, err = time.ParseDuration(duration)
			if err != nil {
				return errors.New("invalid sleep duration")
			}
		}
	default:
		return errors.New("invalid sleep duration")
	}

	if d < 0 {
		return fmt.Errorf("sleep duration must be >= zero")
	}

	select {
	case <-fm.Interrupts():
		return ErrInterrupted
	case <-TimeAfter(fm, d):
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
		fmt.Fprintln(fm.OutputFile(), dt)
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
	out := fm.OutputChan()
	for _, addr := range addrs {
		out <- addr.String()
	}
	return nil
}
