package util

import (
	"bytes"
	"fmt"
)

type TracebackError struct {
	Cause     error
	Traceback []*TracebackEntry
}

func (te *TracebackError) Error() string {
	return te.Cause.Error()
}

func (te *TracebackError) Pprint() string {
	buf := new(bytes.Buffer)
	// Error message
	fmt.Fprintf(buf, "Exception: \033[31;1m%s\033[m\n", te.Cause.Error())
	buf.WriteString("Traceback:\n")

	for _, tb := range te.Traceback {
		buf.WriteString("  ")
		tb.Pprint(buf, "    ")
	}

	return buf.String()
}
