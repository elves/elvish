package eval

import (
	"errors"
	"sync"
	"sync/atomic"

	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/eval/vals"
)

// Flow control.

// TODO(xiaq): Document "multi-error".

func init() {
	addBuiltinFns(map[string]interface{}{
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
// different commands in order to independently capture the output of each byte stream:
//
// ```elvish-transcript
// ~> fn capture {|f|
//      var pout = (file:pipe)
//      var perr = (file:pipe)
//      var out err
//      run-parallel {
//        $f > $pout[w] 2> $perr[w]
//        file:close $pout[w]
//        file:close $perr[w]
//      } {
//        set out = (slurp < $pout[r])
//        file:close $pout[r]
//      } {
//        set err = (slurp < $perr[r])
//        file:close $perr[r]
//      }
//      put $out $err
//    }
// ~> capture { echo stdout-test; echo stderr-test >&2 }
// ▶ "stdout-test\n"
// ▶ "stderr-test\n"
// ```
//
// This command is intended for doing a fixed number of heterogeneous things in
// parallel. If you need homogeneous parallel processing of possibly unbound data,
// use `peach` instead.
//
// @cf peach

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

//elvdoc:fn each
//
// ```elvish
// each $f $inputs?
// ```
//
// Calls `$f` on each [value input](#value-inputs).
//
// An exception raised from [`break`](#break) is caught by `each`, and will
// cause it to terminate early.
//
// An exception raised from [`continue`](#continue) is swallowed and can be used
// to terminate a single iteration early.
//
// Examples:
//
// ```elvish-transcript
// ~> range 5 8 | each {|x| * $x $x }
// ▶ 25
// ▶ 36
// ▶ 49
// ~> each {|x| put $x[:3] } [lorem ipsum]
// ▶ lor
// ▶ ips
// ```
//
// @cf peach
//
// Etymology: Various languages, as `for each`. Happens to have the same name as
// the iteration construct of
// [Factor](http://docs.factorcode.org/content/word-each,sequences.html).

func each(fm *Frame, f Callable, inputs Inputs) error {
	broken := false
	var err error
	inputs(func(v interface{}) {
		if broken {
			return
		}
		newFm := fm.Fork("closure of each")
		ex := f.Call(newFm, []interface{}{v}, NoOpts)
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

//elvdoc:fn peach
//
// ```elvish
// peach $f $inputs?
// ```
//
// Calls `$f` for each [value input](#value-inputs), possibly in parallel.
//
// Like `each`, an exception raised from [`break`](#break) will cause `peach`
// to terminate early. However due to the parallel nature of `peach`, the exact
// time of termination is non-deterministic, and termination is not guaranteed.
//
// An exception raised from [`continue`](#continue) is swallowed and can be used
// to terminate a single iteration early.
//
// Example (your output will differ):
//
// ```elvish-transcript
// ~> range 1 10 | peach {|x| + $x 10 }
// ▶ (num 12)
// ▶ (num 13)
// ▶ (num 11)
// ▶ (num 16)
// ▶ (num 18)
// ▶ (num 14)
// ▶ (num 17)
// ▶ (num 15)
// ▶ (num 19)
// ~> range 1 101 |
//    peach {|x| if (== 50 $x) { break } else { put $x } } |
//    + (all) # 1+...+49 = 1225; 1+...+100 = 5050
// ▶ (num 1328)
// ```
//
// This command is intended for homogeneous processing of possibly unbound data. If
// you need to do a fixed number of heterogeneous things in parallel, use
// `run-parallel`.
//
// @cf each run-parallel

func peach(fm *Frame, f Callable, inputs Inputs) error {
	var wg sync.WaitGroup
	var broken int32
	var errMu sync.Mutex
	var err error

	inputs(func(v interface{}) {
		if atomic.LoadInt32(&broken) != 0 {
			return
		}
		wg.Add(1)
		go func() {
			newFm := fm.Fork("closure of peach")
			newFm.ports[0] = DummyInputPort
			ex := f.Call(newFm, []interface{}{v}, NoOpts)
			newFm.Close()

			if ex != nil {
				switch Reason(ex) {
				case nil, Continue:
					// nop
				case Break:
					atomic.StoreInt32(&broken, 1)
				default:
					errMu.Lock()
					err = diag.Errors(err, ex)
					defer errMu.Unlock()
					atomic.StoreInt32(&broken, 1)
				}
			}
			wg.Done()
		}()
	})
	wg.Wait()
	return err
}

// FailError is an error returned by the "fail" command.
type FailError struct{ Content interface{} }

// Error returns the string representation of the cause.
func (e FailError) Error() string { return vals.ToString(e.Content) }

// Fields returns a structmap for accessing fields from Elvish.
func (e FailError) Fields() vals.StructMap { return failFields{e} }

type failFields struct{ e FailError }

func (failFields) IsStructMap() {}

func (f failFields) Type() string         { return "fail" }
func (f failFields) Content() interface{} { return f.e.Content }

//elvdoc:fn fail
//
// ```elvish
// fail $v
// ```
//
// Throws an exception; `$v` may be any type. If `$v` is already an exception,
// `fail` rethrows it.
//
// ```elvish-transcript
// ~> fail bad
// Exception: bad
// [tty 9], line 1: fail bad
// ~> put ?(fail bad)
// ▶ ?(fail bad)
// ~> fn f { fail bad }
// ~> fail ?(f)
// Exception: bad
// Traceback:
//   [tty 7], line 1:
//     fn f { fail bad }
//   [tty 8], line 1:
//     fail ?(f)
// ```

func fail(v interface{}) error {
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

//elvdoc:fn return
//
// Raises the special "return" exception. When raised inside a named function
// (defined by the [`fn` keyword](language.html#fn)) it is captured by the
// function and causes the function to terminate. It is not captured by an
// ordinary anonymous function.
//
// Because `return` raises an exception it can be caught by a
// [`try`](language.html#try) block. If not caught, either implicitly by a
// named function or explicitly, it causes a failure like any other uncaught
// exception.
//
// See the discussion about [flow commands and
// exceptions](language.html#exception-and-flow-commands)
//
// **Note**: If you want to shadow the builtin `return` function with a local
// wrapper, do not define it with `fn` as `fn` swallows the special exception
// raised by return. Consider this example:
//
// ```elvish-transcript
// ~> use builtin
// ~> fn return { put return; builtin:return }
// ~> fn test-return { put before; return; put after }
// ~> test-return
// ▶ before
// ▶ return
// ▶ after
// ```
//
// Instead, shadow the function by directly assigning to `return~`:
//
// ```elvish-transcript
// ~> use builtin
// ~> var return~ = { put return; builtin:return }
// ~> fn test-return { put before; return; put after }
// ~> test-return
// ▶ before
// ▶ return
// ```

func returnFn() error {
	return Return
}

//elvdoc:fn break
//
// Raises the special "break" exception. When raised inside a loop it is
// captured and causes the loop to terminate.
//
// Because `break` raises an exception it can be caught by a
// [`try`](language.html#try) block. If not caught, either implicitly by a loop
// or explicitly, it causes a failure like any other uncaught exception.
//
// See the discussion about [flow commands and exceptions](language.html#exception-and-flow-commands)
//
// **Note**: You can create a `break` function and it will shadow the builtin
// command. If you do so you should explicitly invoke the builtin. For example:
//
// ```elvish-transcript
// ~> use builtin
// ~> fn break { put 'break'; builtin:break; put 'should not appear' }
// ~> for x [a b c] { put $x; break; put 'unexpected' }
// ▶ a
// ▶ break
// ```

func breakFn() error {
	return Break
}

//elvdoc:fn continue
//
// Raises the special "continue" exception. When raised inside a loop it is
// captured and causes the loop to begin its next iteration.
//
// Because `continue` raises an exception it can be caught by a
// [`try`](language.html#try) block. If not caught, either implicitly by a loop
// or explicitly, it causes a failure like any other uncaught exception.
//
// See the discussion about [flow commands and exceptions](language.html#exception-and-flow-commands)
//
// **Note**: You can create a `continue` function and it will shadow the builtin
// command. If you do so you should explicitly invoke the builtin. For example:
//
// ```elvish-transcript
// ~> use builtin
// ~> fn continue { put 'continue'; builtin:continue; put 'should not appear' }
// ~> for x [a b c] { put $x; continue; put 'unexpected' }
// ▶ a
// ▶ continue
// ▶ b
// ▶ continue
// ▶ c
// ▶ continue
// ```

func continueFn() error {
	return Continue
}

//elvdoc:fn defer
//
// ```elvish
// defer $fn
// ```
//
// Schedules a function to be called when execution reaches the end of the
// current closure. The function is called with no arguments or options, and any
// exception it throws gets propagated.
//
// Examples:
//
// ```elvish-transcript
// ~> { defer { put foo }; put bar }
// ▶ bar
// ▶ foo
// ~> defer { put foo }
// Exception: defer must be called from within a closure
// [tty 2], line 1: defer { put foo }
// ```

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
