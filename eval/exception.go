package eval

import (
	"bytes"
	"fmt"

	"github.com/elves/elvish/util"
)

// Exception represents an elvish exception.
type Exception struct {
	Cause     error
	Traceback *util.SourceContext
}

func (exc *Exception) Error() string {
	return exc.Cause.Error()
}

func (exc *Exception) Pprint() string {
	buf := new(bytes.Buffer)
	// Error message
	fmt.Fprintf(buf, "Exception: \033[31;1m%s\033[m\n", exc.Cause.Error())
	buf.WriteString("Traceback:")

	for tb := exc.Traceback; tb != nil; tb = tb.Next {
		buf.WriteString("\n  ")
		tb.Pprint(buf, "    ")
	}

	return buf.String()
}
