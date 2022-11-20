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

var nopGoFn = NewGoFn("nop", nop)

func nop(opts RawOptions, args ...any) {
	// Do nothing
}

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

func useMod(fm *Frame, spec string) (*Ns, error) {
	return use(fm, spec, nil)
}

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
