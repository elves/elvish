package eval

import (
	"errors"
	"math"
	"math/big"
	"sync"
	"sync/atomic"

	"golang.org/x/sync/semaphore"

	"src.elv.sh/pkg/errutil"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
)

// Flow control.

// TODO(xiaq): Document "multi-error".

func init() {
	addBuiltinFns(map[string]any{
		"run-parallel": runParallel,
		// Exception and control
		"fail":        fail,
		"multi-error": multiErrorFn,
		"return":      returnFn,
		"break":       breakFn,
		"continue":    continueFn,
		"defer":       deferFn,
		// Iterations.
		"each":  each,
		"peach": peach,
	})
}

func runParallel(fm *Frame, functions ...Callable) error {
	var wg sync.WaitGroup
	wg.Add(len(functions))
	exceptions := make([]Exception, len(functions))
	for i, function := range functions {
		go func(fm2 *Frame, function Callable, pexc *Exception) {
			err := function.Call(fm2, NoArgs, NoOpts)
			if err != nil {
				*pexc = err.(Exception)
			}
			wg.Done()
		}(fm.Fork(), function, &exceptions[i])
	}

	wg.Wait()
	return MakePipelineError(exceptions)
}

func each(fm *Frame, f Callable, inputs Inputs) error {
	broken := false
	var err error
	inputs(func(v any) {
		if broken {
			return
		}
		newFm := fm.Fork()
		ex := f.Call(newFm, []any{v}, NoOpts)
		newFm.Close()

		if ex != nil {
			switch Reason(ex) {
			case nil, Continue:
				// nop
			case Break:
				broken = true
			default:
				broken = true
				err = ex
			}
		}
	})
	return err
}

type peachOpt struct{ NumWorkers vals.Num }

func (o *peachOpt) SetDefaultOptions() { o.NumWorkers = math.Inf(1) }

func peach(fm *Frame, opts peachOpt, f Callable, inputs Inputs) error {
	var wg sync.WaitGroup
	var broken int32
	var errMu sync.Mutex
	var err error

	var workerSema *semaphore.Weighted
	numWorkers, limited, err := parseNumWorkers(opts.NumWorkers)
	if err != nil {
		return err
	}
	if limited {
		workerSema = semaphore.NewWeighted(int64(numWorkers))
	}

	ctx := fm.Context()

	inputs(func(v any) {
		if atomic.LoadInt32(&broken) != 0 {
			return
		}
		if workerSema != nil {
			workerSema.Acquire(ctx, 1)
		}
		wg.Add(1)
		go func() {
			newFm := fm.Fork()
			newFm.ports[0] = DummyInputPort
			ex := f.Call(newFm, []any{v}, NoOpts)
			newFm.Close()

			if ex != nil {
				switch Reason(ex) {
				case nil, Continue:
					// nop
				case Break:
					atomic.StoreInt32(&broken, 1)
				default:
					errMu.Lock()
					err = errutil.Multi(err, ex)
					defer errMu.Unlock()
					atomic.StoreInt32(&broken, 1)
				}
			}
			wg.Done()
			if workerSema != nil {
				workerSema.Release(1)
			}
		}()
	})
	wg.Wait()
	return err
}

func parseNumWorkers(n vals.Num) (int, bool, error) {
	switch n := n.(type) {
	case int:
		if n >= 1 {
			return n, true, nil
		}
	case *big.Int:
		// A limit larger than MaxInt is equivalent to no limit.
		return 0, false, nil
	case float64:
		if math.IsInf(n, 1) {
			return 0, false, nil
		}
	}
	return 0, false, errs.BadValue{
		What:   "peach &num-workers",
		Valid:  "exact positive integer or +inf",
		Actual: vals.ToString(n),
	}
}

// FailError is an error returned by the "fail" command.
type FailError struct{ Content any }

var _ vals.PseudoMap = FailError{}

// Error returns the string representation of the cause.
func (e FailError) Error() string { return vals.ToString(e.Content) }

// Kind returns "fail-error".
func (FailError) Kind() string { return "fail-error" }

// Fields returns a [vals.MethodMap] for accessing fields from Elvish.
func (e FailError) Fields() vals.MethodMap { return failFields{e} }

type failFields struct{ e FailError }

func (f failFields) Type() string { return "fail" }
func (f failFields) Content() any { return f.e.Content }

func fail(v any) error {
	if e, ok := v.(error); ok {
		// MAYBE TODO: if v is an exception, attach a "rethrown" stack trace,
		// like Java
		return e
	}
	return FailError{v}
}

func multiErrorFn(excs ...Exception) error {
	return PipelineError{excs}
}

func returnFn() error {
	return Return
}

func breakFn() error {
	return Break
}

func continueFn() error {
	return Continue
}

var errDeferNotInClosure = errors.New("defer must be called from within a closure")

func deferFn(fm *Frame, fn Callable) error {
	if fm.defers == nil {
		return errDeferNotInClosure
	}
	deferTraceback := fm.traceback
	fm.addDefer(func(fm *Frame) Exception {
		err := fn.Call(fm, NoArgs, NoOpts)
		if exc, ok := err.(Exception); ok {
			return exc
		}
		return &exception{err, deferTraceback}
	})
	return nil
}
