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

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/parse"
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

		"deprecate": deprecate,

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
// eval $code &ns=$nil &on-end=$nil
// ```
//
// Evaluates `$code`, which should be a string. The evaluation happens in a
// new, restricted namespace, whose initial set of variables can be specified by
// the `&ns` option. After evaluation completes, the new namespace is passed to
// the callback specified by `&on-end` if it is not nil.
//
// The namespace specified by `&ns` is never modified; it will not be affected
// by the creation or deletion of variables by `$code`. However, the values of
// the variables may be mutated by `$code`.
//
// If the `&ns` option is `$nil` (the default), a temporary namespace built by
// amalgamating the local and upvalue scopes of the caller is used.
//
// If `$code` fails to parse or compile, the parse error or compilation error is
// raised as an exception.
//
// Basic examples that do not modify the namespace or any variable:
//
// ```elvish-transcript
// ~> eval 'put x'
// ▶ x
// ~> x = foo
// ~> eval 'put $x'
// ▶ foo
// ~> ns = (ns [&x=bar])
// ~> eval &ns=$ns 'put $x'
// ▶ bar
// ```
//
// Examples that modify existing variables:
//
// ```elvish-transcript
// ~> y = foo
// ~> eval 'y = bar'
// ~> put $y
// ▶ bar
// ```
//
// Examples that creates new variables and uses the callback to access it:
//
// ```elvish-transcript
// ~> eval 'z = lorem'
// ~> put $z
// compilation error: variable $z not found
// [ttz 2], line 1: put $z
// ~> saved-ns = $nil
// ~> eval &on-end=[ns]{ saved-ns = $ns } 'z = lorem'
// ~> put $saved-ns[z]
// ▶ lorem
// ```

type evalOpts struct {
	Ns    *Ns
	OnEnd Callable
}

func (*evalOpts) SetDefaultOptions() {}

func eval(fm *Frame, opts evalOpts, code string) error {
	src := parse.Source{Name: fmt.Sprintf("[eval %d]", nextEvalCount()), Code: code}
	ns := opts.Ns
	if ns == nil {
		ns = CombineNs(fm.up, fm.local)
	}
	// The stacktrace already contains the line that calls "eval", so we pass
	// nil as the second argument.
	newNs, exc := fm.Eval(src, nil, ns)
	if opts.OnEnd != nil {
		newFm := fm.fork("on-end callback of eval")
		errCb := opts.OnEnd.Call(newFm, []interface{}{newNs}, NoOpts)
		if exc == nil {
			return errCb
		}
	}
	return exc
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
	return use(fm, spec, nil)
}

//elvdoc:fn -source
//
// ```elvish
// -source $filename
// ```
//
// Equivalent to `eval (slurp <$filename)`. Deprecated.

func source(fm *Frame, fname string) error {
	code, err := readFileUTF8(fname)
	if err != nil {
		return err
	}
	src := parse.Source{Name: fname, Code: code, IsFile: true}
	// Amalgamate the up and local scope into a new scope to use as the global
	// scope to evaluate the code in.
	ns := CombineNs(fm.up, fm.local)
	_, exc := fm.Eval(src, nil, ns)
	return exc
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

//elvdoc:fn deprecate
//
// ```elvish
// deprecate $msg
// ```
//
// Shows the given deprecation message to stderr. If called from a function
// or module, also shows the call site of the function or import site of the
// module. Does nothing if the combination of the call site and the message has
// been shown before.
//
// ```elvish-transcript
// ~> deprecate msg
// deprecation: msg
// ~> fn f { deprecate msg }
// ~> f
// deprecation: msg
// [tty 19], line 1: f
// ~> exec
// ~> deprecate msg
// deprecation: msg
// ~> fn f { deprecate msg }
// ~> f
// deprecation: msg
// [tty 3], line 1: f
// ~> f # a different call site; shows deprecate message
// deprecation: msg
// [tty 4], line 1: f
// ~> fn g { f }
// ~> g
// deprecation: msg
// [tty 5], line 1: fn g { f }
// ~> g # same call site, no more deprecation message
// ```

func deprecate(fm *Frame, msg string) {
	var ctx *diag.Context
	if fm.traceback.Next != nil {
		ctx = fm.traceback.Next.Head
	}
	fm.Deprecate(msg, ctx, 0)
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
