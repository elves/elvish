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

type builtinFunc func([]Value, [3]*io) string

type builtin struct {
	fn builtinFunc
	ioTypes [3]ioType
}

var builtins = map[string]builtin {
	"put": builtin{implPut, [3]ioType{unusedIO, chanIO}},
	"print": builtin{implPrint, [3]ioType{unusedIO}},
	"println": builtin{implPrintln, [3]ioType{unusedIO}},
	"printchan": builtin{implPrintchan, [3]ioType{chanIO, fileIO}},
}

func implPut(args []Value, ios [3]*io) string {
	out := ios[1].ch
	for _, a := range args {
		out <- a
	}
	close(out)
	return ""
}

func implPrint(args []Value, ios [3]*io) string {
	out := ios[1].f

	args_if := make([]interface{}, len(args))
	for i, a := range args {
		args_if[i] = a
	}
	fmt.Fprint(out, args_if...)
	return ""
}

func implPrintln(args []Value, ios [3]*io) string {
	args = append(args, NewScalar("\n"))
	return implPrint(args, ios)
}

func implPrintchan(args []Value, ios [3]*io) string {
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
