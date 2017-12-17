package eval

import "errors"

var NotSupportedOnWindows = errors.New("not supported on Windows")

func execFn(ec *EvalCtx, args []Value, opts map[string]Value) {
	throw(NotSupportedOnWindows)
}

func fg(ec *EvalCtx, args []Value, opts map[string]Value) {
	throw(NotSupportedOnWindows)
}
