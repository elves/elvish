package diag

import (
	"fmt"
	"io"
	"os"

	"github.com/elves/elvish/pkg/util"
)

// Can be changed for testing.
var stderr io.Writer = os.Stderr

// PPrintError pretty-prints an error. It uses the PPrint method if the error
// implements PPrinter, and uses Complain to print the error message otherwise.
func PPrintError(err error) {
	if pprinter, ok := err.(util.PPrinter); ok {
		fmt.Fprintln(stderr, pprinter.PPrint(""))
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
