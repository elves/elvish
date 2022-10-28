package md

import (
	"fmt"
	"strings"
)

// OpTraceCodec is a Codec that records all the Op's passed to its Do method.
type OpTraceCodec struct{ strings.Builder }

func (c *OpTraceCodec) Do(op Op) {
	if c.Len() > 0 {
		c.WriteByte('\n')
	}
	c.WriteString(op.Type.String())
	if op.Number != 0 {
		fmt.Fprintf(c, " Number=%d", op.Number)
	}
	if op.Text != "" {
		fmt.Fprintf(c, " Text=%q", op.Text)
	}
	if op.Dest != "" {
		fmt.Fprintf(c, " Dest=%q", op.Dest)
	}
	if op.Alt != "" {
		fmt.Fprintf(c, " Alt=%q", op.Alt)
	}
}
