package eval

import (
	"fmt"
)

type ioType byte

const (
	fileIO ioType = iota // Default IO type. Corresponds to io.f.
	chanIO // Corresponds to io.ch.
	unusedIO
)

type builtinFunc func(*Evaluator, []Value, [3]*io) string

type builtin struct {
	fn builtinFunc
	ioTypes [3]ioType
}

var builtins = map[string]builtin {
	"var": builtin{var_, [3]ioType{unusedIO, unusedIO}},
	"set": builtin{set, [3]ioType{unusedIO, unusedIO}},
	"fn": builtin{fn, [3]ioType{unusedIO, unusedIO}},
	"put": builtin{put, [3]ioType{unusedIO, chanIO}},
	"print": builtin{print, [3]ioType{unusedIO}},
	"println": builtin{println, [3]ioType{unusedIO}},
	"printchan": builtin{printchan, [3]ioType{chanIO, fileIO}},
}

func doSet(ev *Evaluator, nameVal Value, value Value) string {
	name := nameVal.String(ev)
	// TODO Prevent overriding builtin variables e.g. $pid $env
	if _, ok := ev.locals[name]; !ok {
		return fmt.Sprintf("Variable %q doesn't exist", name)
	}
	ev.locals[name] = value
	return ""
}

func var_(ev *Evaluator, args []Value, ios [3]*io) string {
	var names []string
	for _, nameVal := range args {
		name := nameVal.String(ev)
		if _, ok := ev.locals[name]; ok {
			return fmt.Sprintf("Variable %q already exists", name)
		}
		names = append(names, name)
	}

	for _, name := range names {
		ev.locals[name] = nil
	}
	return ""
}

func set(ev *Evaluator, args []Value, ios [3]*io) string {
	if len(args) != 3 || args[1].String(ev) != "=" {
		return "args error"
	}
	return doSet(ev, args[0], args[2])
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
	close(out)
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
		fmt.Fprintf(out, "%q\n", s)
	}
	return ""
}
