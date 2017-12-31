package eval

import (
	"errors"
	"sync"
)

// Flow control.

func init() {
	addToBuiltinFns([]*BuiltinFn{
		{"run-parallel", runParallel},

		// Iterations.
		{"each", each},
		{"peach", peach},

		// Exception and control
		{"fail", fail},
		{"multi-error", multiErrorFn},
		{"return", returnFn},
		{"break", breakFn},
		{"continue", continueFn},
	})
}

func runParallel(ec *Frame, args []Value, opts map[string]Value) {
	var functions []Fn
	ScanArgsVariadic(args, &functions)
	TakeNoOpt(opts)

	var waitg sync.WaitGroup
	waitg.Add(len(functions))
	exceptions := make([]*Exception, len(functions))
	for i, function := range functions {
		go func(ec *Frame, function Fn, exception **Exception) {
			err := ec.PCall(function, NoArgs, NoOpts)
			if err != nil {
				*exception = err.(*Exception)
			}
			waitg.Done()
		}(ec.fork("[run-parallel function]"), function, &exceptions[i])
	}

	waitg.Wait()
	maybeThrow(ComposeExceptionsFromPipeline(exceptions))
}

// each takes a single closure and applies it to all input values.
func each(ec *Frame, args []Value, opts map[string]Value) {
	var f Fn
	iterate := ScanArgsOptionalInput(ec, args, &f)
	TakeNoOpt(opts)

	broken := false
	iterate(func(v Value) {
		if broken {
			return
		}
		// NOTE We don't have the position range of the closure in the source.
		// Ideally, it should be kept in the Closure itself.
		newec := ec.fork("closure of each")
		newec.ports[0] = DevNullClosedChan
		ex := newec.PCall(f, []Value{v}, NoOpts)
		ClosePorts(newec.ports)

		if ex != nil {
			switch ex.(*Exception).Cause {
			case nil, Continue:
				// nop
			case Break:
				broken = true
			default:
				throw(ex)
			}
		}
	})
}

// peach takes a single closure and applies it to all input values in parallel.
func peach(ec *Frame, args []Value, opts map[string]Value) {
	var f Fn
	iterate := ScanArgsOptionalInput(ec, args, &f)
	TakeNoOpt(opts)

	var w sync.WaitGroup
	broken := false
	var err error
	iterate(func(v Value) {
		if broken || err != nil {
			return
		}
		w.Add(1)
		go func() {
			// NOTE We don't have the position range of the closure in the source.
			// Ideally, it should be kept in the Closure itself.
			newec := ec.fork("closure of each")
			newec.ports[0] = DevNullClosedChan
			ex := newec.PCall(f, []Value{v}, NoOpts)
			ClosePorts(newec.ports)

			if ex != nil {
				switch ex.(*Exception).Cause {
				case nil, Continue:
					// nop
				case Break:
					broken = true
				default:
					err = ex
				}
			}
			w.Done()
		}()
	})
	w.Wait()
	maybeThrow(err)
}

func fail(ec *Frame, args []Value, opts map[string]Value) {
	var msg String
	ScanArgs(args, &msg)
	TakeNoOpt(opts)

	throw(errors.New(string(msg)))
}

func multiErrorFn(ec *Frame, args []Value, opts map[string]Value) {
	var excs []*Exception
	ScanArgsVariadic(args, &excs)
	TakeNoOpt(opts)

	throw(PipelineError{excs})
}

func returnFn(ec *Frame, args []Value, opts map[string]Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	throw(Return)
}

func breakFn(ec *Frame, args []Value, opts map[string]Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	throw(Break)
}

func continueFn(ec *Frame, args []Value, opts map[string]Value) {
	TakeNoArg(args)
	TakeNoOpt(opts)

	throw(Continue)
}
