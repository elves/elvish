package util

import (
	"bytes"
	"fmt"
)

// PosError is an error associated with a position range.
type PosError struct {
	Err  error
	Type string
	Traceback
}

func (pe *PosError) Error() string {
	return fmt.Sprintf("%d-%d: %s", pe.Traceback.Begin, pe.Traceback.End, pe.msg())
}

// Pprint pretty-prints a PosError.
func (pe *PosError) Pprint() string {
	buf := new(bytes.Buffer)
	// Error message
	fmt.Fprintf(buf, "%s: \033[31;1m%s\033[m\n", pe.Type, pe.msg())
	// Position
	pe.Traceback.Pprint(buf, "  ")

	return buf.String()
}

func (pe *PosError) msg() string {
	if pe.Err != nil {
		return pe.Err.Error()
	} else {
		return "<nil>"
	}
}
