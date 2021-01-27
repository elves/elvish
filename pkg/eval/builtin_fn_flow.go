package eval

import (
	"sync"

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

func runParallel(fm *Frame, functions ...Callable) error {
	var waitg sync.WaitGroup
	waitg.Add(len(functions))
	exceptions := make([]Exception, len(functions))
	for i, function := range functions {
		go func(fm2 *Frame, function Callable, pexc *Exception) {
			err := function.Call(fm2, NoArgs, NoOpts)
			if err != nil {
				*pexc = err.(Exception)
			}
			waitg.Done()
		}(fm.fork("[run-parallel function]"), function, &exceptions[i])
	}

	waitg.Wait()
	return MakePipelineError(exceptions)
}

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

func each(fm *Frame, f Callable, inputs Inputs) error {
	broken := false
	var err error
	inputs(func(v interface{}) {
		if broken {
			return
		}
		newFm := fm.fork("closure of each")
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
			newFm.ports[0] = DummyInputPort
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
					err = diag.Errors(err, ex)
				}
			}
			w.Done()
		}()
	})
	w.Wait()
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
// (defined by the [`fn` keyword](../language.html#function-definition-fn)) it
// is captured by the function and causes the function to terminate. It is not
// captured by an anonymous function (aka [lambda](../language.html#lambda)).
//
// Because `return` raises an exception it can be caught by a
// [`try`](language.html#exception-control-try) block. If not caught, either
// implicitly by a named function or explicitly, it causes a failure like any
// other uncaught exception.
//
// See the discussion about [flow commands and
// exceptions](language.html#exception-and-flow-commands)
//
// **Note**: If you want to shadow the builtin `return` function with a local
// wrapper, do not define it with `fn` as `fn` swallows the special exception
// raised by return. Consider this example:
//
// ```elvish-transcript
// ~> fn return { put return; builtin:return }
// ~> fn test-return { put before; return; put after }
// ~> test-return
// ▶ before
// ▶ return
// ▶ after
// ```
//
// Instead, shadow the function by directly assigning to `local:return~`:
//
// ```elvish-transcript
// ~> local:return~ = { put return; builtin:return }
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
// [`try`](language.html#exception-control-try) block. If not caught, either
// implicitly by a loop or explicitly, it causes a failure like any other
// uncaught exception.
//
// See the discussion about [flow commands and exceptions](language.html#exception-and-flow-commands)
//
// **Note**: You can create a `break` function and it will shadow the builtin
// command. If you do so you should explicitly invoke the builtin. For example:
//
// ```elvish-transcript
// > fn break []{ put 'break'; builtin:break; put 'should not appear' }
// > for x [a b c] { put $x; break; put 'unexpected' }
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
// [`try`](language.html#exception-control-try) block. If not caught, either
// implicitly by a loop or explicitly, it causes a failure like any other
// uncaught exception.
//
// See the discussion about [flow commands and exceptions](language.html#exception-and-flow-commands)
//
// **Note**: You can create a `continue` function and it will shadow the builtin
// command. If you do so you should explicitly invoke the builtin. For example:
//
// ```elvish-transcript
// > fn break []{ put 'continue'; builtin:continue; put 'should not appear' }
// > for x [a b c] { put $x; continue; put 'unexpected' }
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
