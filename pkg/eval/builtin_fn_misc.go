package eval

// Misc builtin functions.

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"net"
	"strconv"
	"sync"
	"time"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/parse"
)

var (
	ErrNegativeSleepDuration = errors.New("sleep duration must be >= zero")
	ErrInvalidSleepDuration  = errors.New("invalid sleep duration")
)

// Builtins that have not been put into their own groups go here.

func init() {
	addBuiltinFns(map[string]any{
		"kind-of":    kindOf,
		"constantly": constantly,

		// Introspection
		"call":    call,
		"resolve": resolve,
		"eval":    eval,
		"use-mod": useMod,

		"deprecate": deprecate,

		// Time
		"sleep":     sleep,
		"time":      timeCmd,
		"benchmark": benchmark,

		"-ifaddrs": _ifaddrs,
	})

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

var nopGoFn = NewGoFn("nop", nop)

func nop(opts RawOptions, args ...any) {
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

func kindOf(fm *Frame, args ...any) error {
	out := fm.ValueOutput()
	for _, a := range args {
		err := out.Put(vals.Kind(a))
		if err != nil {
			return err
		}
	}
	return nil
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
// ~> var f = (constantly lorem ipsum)
// ~> $f
// ▶ lorem
// ▶ ipsum
// ```
//
// The above example is equivalent to simply `var f = { put lorem ipsum }`;
// it is most useful when the argument is **not** a literal value, e.g.
//
// ```elvish-transcript
// ~> var f = (constantly (uname))
// ~> $f
// ▶ Darwin
// ~> $f
// ▶ Darwin
// ```
//
// The above code only calls `uname` once when defining `$f`. In contrast, if
// `$f` is defined as `var f = { put (uname) }`, every time you invoke `$f`,
// `uname` will be called.
//
// Etymology: [Clojure](https://clojuredocs.org/clojure.core/constantly).

func constantly(args ...any) Callable {
	// TODO(xiaq): Repr of this function is not right.
	return NewGoFn(
		"created by constantly",
		func(fm *Frame) error {
			out := fm.ValueOutput()
			for _, v := range args {
				err := out.Put(v)
				if err != nil {
					return err
				}
			}
			return nil
		},
	)
}

//elvdoc:fn call
//
// ```elvish
// call $fn $args $opts
// ```
//
// Calls `$fn` with `$args` as the arguments, and `$opts` as the option. Useful
// for calling a function with dynamic option keys.
//
// Example:
//
// ```elvish-transcript
// ~> var f = {|a &k1=v1 &k2=v2| put $a $k1 $k2 }
// ~> call $f [foo] [&k1=bar]
// ▶ foo
// ▶ bar
// ▶ v2
// ```

func call(fm *Frame, fn Callable, argsVal vals.List, optsVal vals.Map) error {
	args := make([]any, 0, argsVal.Len())
	for it := argsVal.Iterator(); it.HasElem(); it.Next() {
		args = append(args, it.Elem())
	}
	opts := make(map[string]any, optsVal.Len())
	for it := optsVal.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		ks, ok := k.(string)
		if !ok {
			return errs.BadValue{What: "option key",
				Valid: "string", Actual: vals.Kind(k)}
		}
		opts[ks] = v
	}
	return fn.Call(fm.Fork("-call"), args, opts)
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
// ~> var x = foo
// ~> eval 'put $x'
// ▶ foo
// ~> var ns = (ns [&x=bar])
// ~> eval &ns=$ns 'put $x'
// ▶ bar
// ```
//
// Examples that modify existing variables:
//
// ```elvish-transcript
// ~> var y = foo
// ~> eval 'set y = bar'
// ~> put $y
// ▶ bar
// ```
//
// Examples that creates new variables and uses the callback to access it:
//
// ```elvish-transcript
// ~> eval 'var z = lorem'
// ~> put $z
// compilation error: variable $z not found
// [ttz 2], line 1: put $z
// ~> var saved-ns = $nil
// ~> eval &on-end={|ns| set saved-ns = $ns } 'var z = lorem'
// ~> put $saved-ns[z]
// ▶ lorem
// ```
//
// Note that when using variables from an outer scope, only those
// that have been referenced are captured as upvalues (see [closure
// semantics](language.html#closure-semantics)) and thus accessible to `eval`:
//
// ```elvish-transcript
// ~> var a b
// ~> fn f {|code| nop $a; eval $code }
// ~> f 'echo $a'
// $nil
// ~> f 'echo $b'
// Exception: compilation error: variable $b not found
// [eval 2], line 1: echo $b
// Traceback: [... omitted ...]
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
		newFm := fm.Fork("on-end callback of eval")
		errCb := opts.OnEnd.Call(newFm, []any{newNs}, NoOpts)
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
// ~> echo 'var x = value' > a.elv
// ~> put (use-mod ./a)[x]
// ▶ value
// ```

func useMod(fm *Frame, spec string) (*Ns, error) {
	return use(fm, spec, nil)
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

// Reference to time.After, can be mutated for testing. Takes an additional
// Frame argument to allow inspection of the value of d in tests.
var timeAfter = func(fm *Frame, d time.Duration) <-chan time.Time { return time.After(d) }

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
// A duration can be a simple [number](language.html#number) (with optional
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

func sleep(fm *Frame, duration any) error {
	var f float64
	var d time.Duration

	if err := vals.ScanToGo(duration, &f); err == nil {
		d = time.Duration(f * float64(time.Second))
	} else {
		// See if it is a duration string rather than a simple number.
		switch duration := duration.(type) {
		case string:
			d, err = time.ParseDuration(duration)
			if err != nil {
				return ErrInvalidSleepDuration
			}
		default:
			return ErrInvalidSleepDuration
		}
	}

	if d < 0 {
		return ErrNegativeSleepDuration
	}

	select {
	case <-fm.Interrupts():
		return ErrInterrupted
	case <-timeAfter(fm, d):
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
// ~> var t = ''
// ~> time &on-end={|x| set t = $x } { sleep 1 }
// ~> put $t
// ▶ (num 1.000925004)
// ~> time &on-end={|x| set t = $x } { sleep 0.01 }
// ~> put $t
// ▶ (num 0.011030208)
// ```
//
// @cf benchmark

type timeOpt struct{ OnEnd Callable }

func (o *timeOpt) SetDefaultOptions() {}

func timeCmd(fm *Frame, opts timeOpt, f Callable) error {
	t0 := time.Now()
	err := f.Call(fm, NoArgs, NoOpts)
	t1 := time.Now()

	dt := t1.Sub(t0)
	if opts.OnEnd != nil {
		newFm := fm.Fork("on-end callback of time")
		errCb := opts.OnEnd.Call(newFm, []any{dt.Seconds()}, NoOpts)
		if err == nil {
			err = errCb
		}
	} else {
		_, errWrite := fmt.Fprintln(fm.ByteOutput(), dt)
		if err == nil {
			err = errWrite
		}
	}

	return err
}

//elvdoc:fn benchmark
//
// ```elvish
// benchmark &min-runs=5 &min-time=1s &on-end=$nil &on-run-end=$nil $callable
// ```
//
// Runs `$callable` repeatedly, and reports statistics about how long each run
// takes.
//
// If the `&on-end` callback is not given, `benchmark` prints the average,
// standard deviation, minimum and maximum of the time it took to run
// `$callback`, and the number of runs. If the `&on-end` callback is given,
// `benchmark` instead calls it with a map containing these metrics, keyed by
// `avg`, `stddev`, `min`, `max` and `runs`. Each duration value (i.e. all
// except `runs`) is given as the number of seconds.
//
// The number of runs is controlled by `&min-runs` and `&min-time`. The
// `$callable` is run at least `&min-runs` times, **and** when the total
// duration is at least `&min-time`.
//
// The `&min-runs` option must be a non-negative integer within the range of the
// machine word.
//
// The `&min-time` option must be a string representing a non-negative duration,
// specified as a sequence of decimal numbers with a unit suffix (the numbers
// may have fractional parts), such as "300ms", "1.5h" and "1h45m7s". Valid time
// units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
//
// If `&on-run-end` is given, it is called after each call to `$callable`, with
// the time that call took, given as the number of seconds.
//
// If `$callable` throws an exception, `benchmark` terminates and propagates the
// exception after the `&on-end` callback (or the default printing behavior)
// finishes. The duration of the call that throws an exception is not passed to
// `&on-run-end`, nor is it included when calculating the statistics for
// `&on-end`. If the first call to `$callable` throws an exception and `&on-end`
// is `$nil`, nothing is printed and any `&on-end` callback is not called.
//
// If `&on-run-end` is given and throws an exception, `benchmark` terminates and
// propagates the exception after the `&on-end` callback (or the default
// printing behavior) finishes, unless `$callable` has already thrown an
// exception
//
// If `&on-end` throws an exception, the exception is propagated, unless
// `$callable` or `&on-run-end` has already thrown an exception.
//
// Example:
//
// ```elvish-transcript
// ~> benchmark { }
// 98ns ± 382ns (min 0s, max 210.417µs, 10119226 runs)
// ~> benchmark &on-end={|m| put $m[avg]} { }
// ▶ (num 9.8e-08)
// ~> benchmark &on-run-end={|d| echo $d} { sleep 0.3 }
// 0.301123625
// 0.30123775
// 0.30119075
// 0.300629166
// 0.301260333
// 301.088324ms ± 234.298µs (min 300.629166ms, max 301.260333ms, 5 runs)
// ```
//
// @cf time

type benchmarkOpts struct {
	OnEnd    Callable
	OnRunEnd Callable
	MinRuns  int
	MinTime  string
	minTime  time.Duration
}

func (o *benchmarkOpts) SetDefaultOptions() {
	o.MinRuns = 5
	o.minTime = time.Second
}

func (opts *benchmarkOpts) parse() error {
	if opts.MinRuns < 0 {
		return errs.BadValue{What: "min-runs option",
			Valid: "non-negative integer", Actual: strconv.Itoa(opts.MinRuns)}
	}

	if opts.MinTime != "" {
		d, err := time.ParseDuration(opts.MinTime)
		if err != nil {
			return errs.BadValue{What: "min-time option",
				Valid: "duration string", Actual: parse.Quote(opts.MinTime)}
		}
		if d < 0 {
			return errs.BadValue{What: "min-time option",
				Valid: "non-negative duration", Actual: parse.Quote(opts.MinTime)}
		}
		opts.minTime = d
	}

	return nil
}

// TimeNow is a reference to [time.Now] that can be overridden in tests.
var TimeNow = time.Now

func benchmark(fm *Frame, opts benchmarkOpts, f Callable) error {
	if err := opts.parse(); err != nil {
		return err
	}

	// Standard deviation is calculated using https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance#Welford's_online_algorithm
	var (
		min   = time.Duration(math.MaxInt64)
		max   = time.Duration(math.MinInt64)
		runs  int64
		total time.Duration
		m2    float64
		err   error
	)
	for {
		t0 := TimeNow()
		err = f.Call(fm, NoArgs, NoOpts)
		if err != nil {
			break
		}
		dt := TimeNow().Sub(t0)

		if min > dt {
			min = dt
		}
		if max < dt {
			max = dt
		}
		var oldDelta float64
		if runs > 0 {
			oldDelta = float64(dt) - float64(total)/float64(runs)
		}
		runs++
		total += dt
		if runs > 0 {
			newDelta := float64(dt) - float64(total)/float64(runs)
			m2 += oldDelta * newDelta
		}

		if opts.OnRunEnd != nil {
			newFm := fm.Fork("on-run-end callback of benchmark")
			err = opts.OnRunEnd.Call(newFm, []any{dt.Seconds()}, NoOpts)
			if err != nil {
				break
			}
		}

		if runs >= int64(opts.MinRuns) && total >= opts.minTime {
			break
		}
	}

	if runs == 0 {
		return err
	}

	avg := total / time.Duration(runs)
	stddev := time.Duration(math.Sqrt(m2 / float64(runs)))
	if opts.OnEnd == nil {
		_, errOut := fmt.Fprintf(fm.ByteOutput(),
			"%v ± %v (min %v, max %v, %d runs)\n", avg, stddev, min, max, runs)
		if err == nil {
			err = errOut
		}
	} else {
		stats := vals.MakeMap(
			"avg", avg.Seconds(), "stddev", stddev.Seconds(),
			"min", min.Seconds(), "max", max.Seconds(), "runs", int64ToElv(runs))
		newFm := fm.Fork("on-end callback of benchmark")
		errOnEnd := opts.OnEnd.Call(newFm, []any{stats}, NoOpts)
		if err == nil {
			err = errOnEnd
		}
	}
	return err
}

func int64ToElv(i int64) any {
	if i <= int64(math.MaxInt) {
		return int(i)
	} else {
		return big.NewInt(i)
	}
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
	out := fm.ValueOutput()
	for _, addr := range addrs {
		err := out.Put(addr.String())
		if err != nil {
			return err
		}
	}
	return nil
}
