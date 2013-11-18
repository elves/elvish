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
	// TODO Support `fn f a b c { cmd }` as sugar for `fn f { | a b c | cmd }`
	if len(args) != 2 {
		return "args error"
	}
	if _, ok := args[1].(*Closure); !ok {
		return "args error"
	}
	return doSet(ev, args[0], args[1])
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
