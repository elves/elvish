package eval

import "github.com/elves/elvish/eval/types"

// Basic predicate commands.

func init() {
	addToBuiltinFns([]*BuiltinFn{
		{"bool", boolFn},
		{"not", not},
		{"is", is},
		{"eq", eq},
		{"not-eq", notEq},
	})
}

func boolFn(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var v interface{}
	ScanArgs(args, &v)
	TakeNoOpt(opts)

	ec.OutputChan() <- types.Bool(v)
}

func not(ec *Frame, args []interface{}, opts map[string]interface{}) {
	var v interface{}
	ScanArgs(args, &v)
	TakeNoOpt(opts)

	ec.OutputChan() <- !types.Bool(v)
}

func is(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoOpt(opts)
	result := true
	for i := 0; i+1 < len(args); i++ {
		if args[i] != args[i+1] {
			result = false
			break
		}
	}
	ec.OutputChan() <- types.Bool(result)
}

func eq(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoOpt(opts)
	result := true
	for i := 0; i+1 < len(args); i++ {
		if !types.Equal(args[i], args[i+1]) {
			result = false
			break
		}
	}
	ec.OutputChan() <- types.Bool(result)
}

func notEq(ec *Frame, args []interface{}, opts map[string]interface{}) {
	TakeNoOpt(opts)
	result := true
	for i := 0; i+1 < len(args); i++ {
		if types.Equal(args[i], args[i+1]) {
			result = false
			break
		}
	}
	ec.OutputChan() <- types.Bool(result)
}
