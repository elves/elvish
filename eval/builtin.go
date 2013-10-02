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

type builtinFunc func([]string, [3]*io) string

type builtin struct {
	f builtinFunc
	ioTypes [3]ioType
}

var builtins = map[string]builtin {
	"put": builtin{implPut, [3]ioType{unusedIO, chanIO}},
	"print": builtin{implPrint, [3]ioType{unusedIO}},
	"println": builtin{implPrintln, [3]ioType{unusedIO}},
}

func implPut(args []string, ios [3]*io) string {
	out := ios[1].ch
	for i := 1; i < len(args); i++ {
		out <- args[i]
	}
	return ""
}

func implPrint(args []string, ios [3]*io) string {
	out := ios[1].f

	args = args[1:]
	args_if := make([]interface{}, len(args))
	for i, a := range args {
		args_if[i] = a
	}
	fmt.Fprint(out, args_if...)
	return ""
}

func implPrintln(args []string, ios [3]*io) string {
	args = append(args, "\n")
	return implPrint(args, ios)
}
