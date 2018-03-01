package eval

import (
	"errors"
	"sync"

	"github.com/elves/elvish/util"
)

// Flow control.

func init() {
	addBuiltinFns(map[string]interface{}{
		"run-parallel": runParallel,
		// Exception and control
		"fail":        fail,
		"multi-error": multiErrorFn,
		"return":      returnFn,
		"break":       breakFn,
		"continue":    continueFn,
		// Iterations.
		"each":  each,
		"peach": peach,
	})
}

func runParallel(fm *Frame, functions ...Callable) error {
	var waitg sync.WaitGroup
	waitg.Add(len(functions))
	exceptions := make([]*Exception, len(functions))
	for i, function := range functions {
		go func(fm2 *Frame, function Callable, exception **Exception) {
			err := fm2.Call(function, NoArgs, NoOpts)
			if err != nil {
				*exception = err.(*Exception)
			}
			waitg.Done()
		}(fm.fork("[run-parallel function]"), function, &exceptions[i])
	}

	waitg.Wait()
	return ComposeExceptionsFromPipeline(exceptions)
}

// each takes a single closure and applies it to all input values.
func each(fm *Frame, f Callable, inputs Inputs) error {
	broken := false
	var err error
	inputs(func(v interface{}) {
		if broken {
			return
		}
		newFm := fm.fork("closure of each")
		newFm.ports[0] = DevNullClosedChan
		ex := newFm.Call(f, []interface{}{v}, NoOpts)
		newFm.Close()

		if ex != nil {
			switch ex.(*Exception).Cause {
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

// peach takes a single closure and applies it to all input values in parallel.
func peach(fm *Frame, f Callable, inputs Inputs) error {
	var w sync.WaitGroup
	broken := false
	var err error
	inputs(func(v interface{}) {
		if broken || err != nil {
			return
		}
		w.Add(1)
		go func() {
			newFm := fm.fork("closure of peach")
			newFm.ports[0] = DevNullClosedChan
			ex := newFm.Call(f, []interface{}{v}, NoOpts)
			newFm.Close()

			if ex != nil {
				switch ex.(*Exception).Cause {
				case nil, Continue:
					// nop
				case Break:
					broken = true
				default:
					broken = true
					err = util.Errors(err, ex)
				}
			}
			w.Done()
		}()
	})
	w.Wait()
	return err
}

func fail(msg string) error {
	return errors.New(msg)
}

func multiErrorFn(excs ...*Exception) error {
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
