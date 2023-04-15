package eval

import (
	"context"
	"errors"
	"fmt"
	"math"
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
		}(fm.Fork("[run-parallel function]"), function, &exceptions[i])
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
		newFm := fm.Fork("closure of each")
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

type peachOpt struct{ NumWorkers int }

func (o *peachOpt) SetDefaultOptions() { o.NumWorkers = -1 }

func peach(fm *Frame, opts peachOpt, f Callable, inputs Inputs) error {
	var wg sync.WaitGroup
	var broken int32
	var errMu sync.Mutex
	var err error

	var workerSema *semaphore.Weighted
	switch {
	case opts.NumWorkers == -1:
		workerSema = semaphore.NewWeighted(math.MaxInt64)
	case opts.NumWorkers > 0:
		workerSema = semaphore.NewWeighted(int64(opts.NumWorkers))
	default:
		return errs.BadValue{
			What:   "peach &num-workers",
			Valid:  "-1 or > 0",
			Actual: fmt.Sprintf("%d", opts.NumWorkers),
		}
	}
	ctx := context.TODO() // workerSema needs a context so use a dummy context today

	inputs(func(v any) {
		if atomic.LoadInt32(&broken) != 0 {
			return
		}
		workerSema.Acquire(ctx, 1)
		wg.Add(1)
		go func() {
			newFm := fm.Fork("closure of peach")
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
			workerSema.Release(1)
		}()
	})
	wg.Wait()
	return err
}

// FailError is an error returned by the "fail" command.
type FailError struct{ Content any }

// Error returns the string representation of the cause.
func (e FailError) Error() string { return vals.ToString(e.Content) }

// Fields returns a structmap for accessing fields from Elvish.
func (e FailError) Fields() vals.StructMap { return failFields{e} }

type failFields struct{ e FailError }

func (failFields) IsStructMap() {}

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
