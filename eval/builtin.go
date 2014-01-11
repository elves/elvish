package eval

import (
	"os"
	"fmt"
)

type BuiltinFunc func(*Evaluator, []Value, [3]*io) string

type builtin struct {
	fn BuiltinFunc
	ioTypes [3]IOType
}

var builtins = map[string]builtin {
	"var": builtin{var_, [3]IOType{unusedIO, unusedIO}},
	"set": builtin{set, [3]IOType{unusedIO, unusedIO}},
	"fn": builtin{fn, [3]IOType{unusedIO, unusedIO}},
	"put": builtin{put, [3]IOType{unusedIO, chanIO}},
	"print": builtin{print, [3]IOType{unusedIO}},
	"println": builtin{println, [3]IOType{unusedIO}},
	"printchan": builtin{printchan, [3]IOType{chanIO, fileIO}},
	"cd": builtin{cd, [3]IOType{unusedIO, unusedIO}},
}

func doSet(ev *Evaluator, names []string, values []Value) string {
	// TODO Support assignment of mismatched arity in some restricted way -
	// "optional" and "rest" arguments and the like
	if len(names) != len(values) {
		return "arity mismatch"
	}

	for i, name := range names {
		// TODO Prevent overriding builtin variables e.g. $pid $env
		ev.locals[name] = values[i]
	}

	return ""
}

func var_(ev *Evaluator, args []Value, ios [3]*io) string {
	var names []string
	var values []Value
	for i, nameVal := range args {
		name := nameVal.String(ev)
		if name == "=" {
			values = args[i+1:]
			break
		}
		if _, ok := ev.locals[name]; ok {
			return fmt.Sprintf("Variable %q already exists", name)
		}
		names = append(names, name)
	}

	for _, name := range names {
		ev.locals[name] = nil
	}
	if values != nil {
		return doSet(ev, names, values)
	}
	return ""
}

func set(ev *Evaluator, args []Value, ios [3]*io) string {
	var names []string
	var values []Value
	for i, nameVal := range args {
		name := nameVal.String(ev)
		if name == "=" {
			values = args[i+1:]
			break
		}
		if _, ok := ev.locals[name]; !ok {
			return fmt.Sprintf("Variable %q doesn't exists", name)
		}
		names = append(names, name)
	}

	if values == nil {
		return "missing equal sign"
	}
	return doSet(ev, names, values)
}

func fn(ev *Evaluator, args []Value, ios [3]*io) string {
	n := len(args)
	if n < 2 {
		return "args error"
	}
	closure, ok := args[n-1].(*Closure)
	if !ok {
		return "args error"
	}
	if n > 2 && len(closure.ArgNames) != 0 {
		return "can't define arg names list twice"
	}
	// XXX Should either make a copy of closure or forbid the following:
	// var f; set f = { }
	// fn g a b $f // Changes arity of $f!
	for i := 1; i < n-1; i++ {
		closure.ArgNames = append(closure.ArgNames, args[i].String(ev))
	}
	// TODO Warn about redefining fn?
	ev.locals["fn-" + args[0].String(ev)] = closure
	return ""
}

func put(ev *Evaluator, args []Value, ios [3]*io) string {
	out := ios[1].ch
	for _, a := range args {
		out <- a
	}
	return ""
}

func print(ev *Evaluator, args []Value, ios [3]*io) string {
	out := ios[1].f
	for _, a := range args {
		fmt.Fprint(out, a.String(ev))
	}
	return ""
}

func println(ev *Evaluator, args []Value, ios [3]*io) string {
	args = append(args, NewScalar("\n"))
	return print(ev, args, ios)
}

func printchan(ev *Evaluator, args []Value, ios [3]*io) string {
	if len(args) > 0 {
		return "args error"
	}
	in := ios[0].ch
	out := ios[1].f

	for s := range in {
		fmt.Fprintln(out, s.String(ev))
	}
	return ""
}

func cd(ev *Evaluator, args []Value, ios [3]*io) string {
	var dir string
	if len(args) == 0 {
		dir = ""
	} else if len(args) == 1 {
		dir = args[0].String(ev)
	} else {
		return "args error"
	}
	err := os.Chdir(dir)
	if err != nil {
		return err.Error()
	}
	return ""
}
