package md

import "strings"

// TextCodec is a codec that dumps the pure text content of Markdown.
type TextCodec struct {
	blocks []TextBlock
}

// TextBlock is a text block dumped by TextCodec.
type TextBlock struct {
	Text string
	Code bool
}

func (c *TextCodec) Do(op Op) {
	switch op.Type {
	case OpHeading, OpParagraph:
		var sb strings.Builder
		for _, op := range op.Content {
			switch op.Type {
			case OpText, OpCodeSpan, OpAutolink:
				sb.WriteString(op.Text)
			case OpNewLine:
				sb.WriteByte(' ')
			case OpHardLineBreak:
				sb.WriteByte('\n')
			}
		}
		c.add(TextBlock{sb.String(), false})
	case OpCodeBlock:
		c.add(TextBlock{strings.Join(op.Lines, "\n"), true})
	}
}

func (c *TextCodec) Blocks() []TextBlock { return c.blocks }

func (c *TextCodec) add(b TextBlock) { c.blocks = append(c.blocks, b) }
