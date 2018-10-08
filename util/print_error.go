package util

import (
	"fmt"
	"io"
	"os"
)

// Can be changed for testing.
var stderr io.Writer = os.Stderr

// PprintError pretty-prints an error. It uses the Pprint method if the error
// implements Pprinter, and uses Complain to print the error message otherwise.
func PprintError(err error) {
	if pprinter, ok := err.(Pprinter); ok {
		fmt.Fprintln(stderr, pprinter.Pprint(""))
	} else {
		Complain(err.Error())
	}
}

// Complain prints a message to stderr in bold and red, adding a trailing
// newline.
func Complain(msg string) {
	fmt.Fprintf(stderr, "\033[31;1m%s\033[m\n", msg)
}

// Complainf is like Complain, but accepts a format string and arguments.
func Complainf(format string, args ...interface{}) {
	Complain(fmt.Sprintf(format, args...))
}
