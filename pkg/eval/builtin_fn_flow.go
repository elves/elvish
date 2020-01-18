package eval

import (
	"errors"
	"sync"

	"github.com/elves/elvish/pkg/util"
)

// Flow control.

//elvdoc:fn run-parallel
//
// ```elvish
// run-parallel $callable ...
// ```
//
// Run several callables in parallel, and wait for all of them to finish.
//
// If one or more callables throw exceptions, the other callables continue running,
// and a composite exception is thrown when all callables finish execution.
//
// The behavior of `run-parallel` is consistent with the behavior of pipelines,
// except that it does not perform any redirections.
//
// Here is an example that lets you pipe the stdout and stderr of a command to two
// different commands:
//
// ```elvish
// pout = (pipe)
// perr = (pipe)
// run-parallel {
// foo > $pout 2> $perr
// pwclose $pout
// pwclose $perr
// } {
// bar < $pout
// prclose $pout
// } {
// bar2 < $perr
// prclose $perr
// }
// ```
//
// This command is intended for doing a fixed number of heterogeneous things in
// parallel. If you need homogeneous parallel processing of possibly unbound data,
// use `peach` instead.
//
// @cf peach

//elvdoc:fn fail
//
// ```elvish
// fail $message
// ```
//
// Throw an exception.
//
// ```elvish-transcript
// ~> fail bad
// Exception: bad
// Traceback:
// [interactive], line 1:
// fail bad
// ~> put ?(fail bad)
// ▶ ?(fail bad)
// ```
//
// **Note**: Exceptions are now only allowed to carry string messages. You cannot
// do `fail [&cause=xxx]` (this will, ironically, throw a different exception
// complaining that you cannot throw a map). This is subject to change. Builtins
// will likely also throw structured exceptions in future.

// TODO(xiaq): Document "multi-error".

// TODO(xiaq): Document "return", "break" and "continue".

//elvdoc:fn each
//
// ```elvish
// each $f $input-list?
// ```
//
// Call `$f` on all inputs. Examples:
//
// ```elvish-transcript
// ~> range 5 8 | each [x]{ ^ $x 2 }
// ▶ 25
// ▶ 36
// ▶ 49
// ~> each [x]{ put $x[:3] } [lorem ipsum]
// ▶ lor
// ▶ ips
// ```
//
// @cf peach
//
// Etymology: Various languages, as `for each`. Happens to have the same name as
// the iteration construct of
// [Factor](http://docs.factorcode.org/content/word-each,sequences.html).

//elvdoc:fn peach
//
// ```elvish
// peach $f $input-list?
// ```
//
// Call `$f` on all inputs, possibly in parallel.
//
// Example (your output will differ):
//
// ```elvish-transcript
// ~> range 1 7 | peach [x]{ + $x 10 }
// ▶ 12
// ▶ 11
// ▶ 13
// ▶ 16
// ▶ 15
// ▶ 14
// ```
//
// This command is intended for homogeneous processing of possibly unbound data. If
// you need to do a fixed number of heterogeneous things in parallel, use
// `run-parallel`.
//
// @cf each run-parallel

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
		ex := newFm.Call(f, []interface{}{v}, NoOpts)
		newFm.Close()

		if ex != nil {
			switch Cause(ex) {
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
				switch Cause(ex) {
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
