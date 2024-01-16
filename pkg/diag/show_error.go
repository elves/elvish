package diag

import (
	"fmt"
	"io"
)

// ShowError shows an error. It uses the Show method if the error
// implements Shower. Otherwise, it prints the error in bold and red, with a
// trailing newline.
func ShowError(w io.Writer, err error) {
	if shower, ok := err.(Shower); ok {
		fmt.Fprintln(w, shower.Show(""))
	} else {
		fmt.Fprintf(w, "\033[31;1m%s\033[m\n", err.Error())
	}
}
