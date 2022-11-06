package md

import (
	"fmt"
	"strings"
)

// TraceCodec is a Codec that records all the Op's passed to its Do method.
type TraceCodec struct {
	ops []Op
	strings.Builder
}

func (c *TraceCodec) Do(op Op) {
	c.ops = append(c.ops, op)
	c.WriteString(op.Type.String())
	if op.Number != 0 {
		fmt.Fprintf(c, " Number=%d", op.Number)
	}
	if op.Info != "" {
		fmt.Fprintf(c, " Info=%q", op.Info)
	}
	c.WriteByte('\n')
	for _, line := range op.Lines {
		c.WriteString("  ")
		c.WriteString(line)
		c.WriteByte('\n')
	}
	for _, inlineOp := range op.Content {
		c.WriteString("  ")
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
		c.WriteString("\n")
	}
}

func (c *TraceCodec) Ops() []Op { return c.ops }
