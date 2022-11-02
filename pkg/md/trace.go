package md

import (
	"fmt"
	"strings"
)

// TraceCodec is a Codec that records all the Op's passed to its Do method.
type TraceCodec struct{ strings.Builder }

func (c *TraceCodec) Do(op Op) {
	if c.Len() > 0 {
		c.WriteByte('\n')
	}
	c.WriteString(op.Type.String())
	if op.Number != 0 {
		fmt.Fprintf(c, " Number=%d", op.Number)
	}
	if op.Info != "" {
		fmt.Fprintf(c, " Info=%q", op.Info)
	}
	if op.MissingCloser {
		fmt.Fprintf(c, " MissingCloser")
	}
	for _, line := range op.Lines {
		c.WriteString("\n  ")
		c.WriteString(line)
	}
	for _, inlineOp := range op.Content {
		c.WriteString("\n  ")
		c.WriteString(inlineOp.Type.String())
		if inlineOp.Text != "" {
			fmt.Fprintf(c, " Text=%q", inlineOp.Text)
		}
		if inlineOp.Dest != "" {
			fmt.Fprintf(c, " Dest=%q", inlineOp.Dest)
		}
		if inlineOp.Alt != "" {
			fmt.Fprintf(c, " Alt=%q", inlineOp.Alt)
		}
	}
}
