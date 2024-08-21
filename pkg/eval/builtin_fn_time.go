package eval

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"time"

	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/parse"
)

func init() {
	addBuiltinFns(map[string]any{
		"sleep":     sleep,
		"time":      timeCmd,
		"benchmark": benchmark,
	})
}

var (
	// Reference to [time.After] that can be mutated for testing. Takes an
	// additional Frame argument to allow inspection of the value of d in tests.
	timeAfter = func(fm *Frame, d time.Duration) <-chan time.Time { return time.After(d) }
	// Reference to [time.Now] that can be overridden in tests.
	timeNow = time.Now
)

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
	case <-fm.Context().Done():
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
		newFm := fm.Fork()
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
}

func (o *benchmarkOpts) SetDefaultOptions() {
	o.MinRuns = 5
}

func (opts *benchmarkOpts) parse() (time.Duration, error) {
	if opts.MinRuns < 0 {
		return 0, errs.BadValue{What: "min-runs option",
			Valid: "non-negative integer", Actual: strconv.Itoa(opts.MinRuns)}
	}

	if opts.MinTime != "" {
		d, err := time.ParseDuration(opts.MinTime)
		if err != nil {
			return 0, errs.BadValue{What: "min-time option",
				Valid: "duration string", Actual: parse.Quote(opts.MinTime)}
		}
		if d < 0 {
			return 0, errs.BadValue{What: "min-time option",
				Valid: "non-negative duration", Actual: parse.Quote(opts.MinTime)}
		}
		return d, nil
	}

	// Use 1s as the default minTime.
	return time.Second, nil
}

func benchmark(fm *Frame, opts benchmarkOpts, f Callable) error {
	minTime, err := opts.parse()
	if err != nil {
		return err
	}

	// Standard deviation is calculated using https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance#Welford's_online_algorithm
	var (
		min   = time.Duration(math.MaxInt64)
		max   = time.Duration(math.MinInt64)
		runs  int64
		total time.Duration
		m2    float64
	)
	for {
		t0 := timeNow()
		err = f.Call(fm, NoArgs, NoOpts)
		if err != nil {
			break
		}
		dt := timeNow().Sub(t0)

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
			newFm := fm.Fork()
			err = opts.OnRunEnd.Call(newFm, []any{dt.Seconds()}, NoOpts)
			if err != nil {
				break
			}
		}

		if runs >= int64(opts.MinRuns) && total >= minTime {
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
		newFm := fm.Fork()
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
