package diag

import (
	"fmt"
	"io"
)

// ShowError shows an error. It uses the Show method if the error
// implements Shower, and uses Complain to print the error message otherwise.
func ShowError(w io.Writer, err error) {
	if shower, ok := err.(Shower); ok {
		fmt.Fprintln(w, shower.Show(""))
	} else {
		Complain(w, err.Error())
	}
}

// Complain prints a message to w in bold and red, adding a trailing newline.
func Complain(w io.Writer, msg string) {
	fmt.Fprintf(w, "\033[31;1m%s\033[m\n", msg)
}

// Complainf is like Complain, but accepts a format string and arguments.
func Complainf(w io.Writer, format string, args ...interface{}) {
	Complain(w, fmt.Sprintf(format, args...))
}
