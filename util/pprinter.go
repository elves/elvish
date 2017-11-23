package util

import (
	"fmt"
	"os"
)

type Pprinter interface {
	Pprint(indent string) string
}

// PprintError pretty-prints an error if it implements Pprinter, and prints it
// in bold and red otherwise.
func PprintError(err error) {
	if pprinter, ok := err.(Pprinter); ok {
		fmt.Fprintln(os.Stderr, pprinter.Pprint(""))
	} else {
		fmt.Fprintf(os.Stderr, "\033[31;1m%s\033[m", err.Error())
	}
}
