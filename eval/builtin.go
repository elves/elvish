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
	"set": builtin{implSet, [3]ioType{unusedIO, unusedIO}},
	"put": builtin{implPut, [3]ioType{unusedIO, chanIO}},
	"print": builtin{implPrint, [3]ioType{unusedIO}},
	"println": builtin{implPrintln, [3]ioType{unusedIO}},
	"printchan": builtin{implPrintchan, [3]ioType{chanIO, fileIO}},
}

func implSet(ev *Evaluator, args []Value, ios [3]*io) string {
	// TODO Support setting locals
	// TODO Prevent overriding builtin variables e.g. $pid $env
	if len(args) != 3 || args[1].String(ev) != "=" {
		return "args error"
	}
	ev.globals[args[0].String(ev)] = args[2]
	return ""
}

func implPut(ev *Evaluator, args []Value, ios [3]*io) string {
	out := ios[1].ch
	for _, a := range args {
		out <- a
	}
	close(out)
	return ""
}

func implPrint(ev *Evaluator, args []Value, ios [3]*io) string {
	out := ios[1].f
	for _, a := range args {
		fmt.Fprint(out, a.String(ev))
	}
	return ""
}

func implPrintln(ev *Evaluator, args []Value, ios [3]*io) string {
	args = append(args, NewScalar("\n"))
	return implPrint(ev, args, ios)
}

func implPrintchan(ev *Evaluator, args []Value, ios [3]*io) string {
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
