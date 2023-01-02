package md

import (
	"fmt"
	"regexp"
	"strings"

	"src.elv.sh/pkg/ui"
	"src.elv.sh/pkg/wcwidth"
)

// TTYCodec renders Markdown in a terminal.
//
// The rendered text uses the following style:
//
//   - Adjacent blocks are always separated with one blank line.
//
//   - Thematic breaks are rendered as "────" (four U+2500 "box drawing light
//     horizontal").
//
//   - Headings are rendered like "# Heading" in bold, with the same number of
//     hashes as in Markdown
//
//   - Code blocks are indented two spaces. The HighlightCodeBlock callback can
//     be supplied to highlight the content of the code block.
//
//   - HTML blocks are ignored.
//
//   - Paragraphs are always reflowed to fit the given width.
//
//   - Blockquotes start with "│ " (U+2502 "box drawing light vertical", then a
//     space) on each line.
//
//   - Bullet list items start with "• " (U+2022 "bullet", then a space) on the
//     first line. Continuation lines are indented two spaces.
//
//   - Ordered list items start with "X. " (where X is a number) on the first
//     line. Continuation lines are indented three spaces.
//
//   - Code spans are underlined.
//
//   - Emphasis makes the text italic. (Some terminal emulators turn italic text
//     into inverse text, which is not ideal but fine.)
//
//   - Strong emphasis makes the text bold.
//
//   - Links are rendered with their text content underlined. If the link is
//     absolute (starts with scheme:), the destination is rendered like "
//     (https://example.com)" after the text content.
//
//     Relative link destinations are not shown by default, since they are
//     usually not useful in a terminal. If the ConvertRelativeLink callback is
//     non-nil, it is called for each relative links and non-empty return values
//     are shown.
//
//     The link description is ignored for now since Elvish's Markdown sources
//     never use them.
//
//   - Images are rendered like "Image: alt text (https://example.com/a.png)".
//
//   - Autolinks have their text content rendered.
//
//   - Raw HTML is mostly ignored, except that text between <kbd> and </kbd>
//     becomes inverse video.
//
//   - Hard line breaks are respected.
//
// The structure of the implementation closely mirrors [FmtCodec] in a lot of
// places, without the complexity of handling all edge cases correctly, but with
// the slight complexity of handling styles.
type TTYCodec struct {
	Width int
	// If non-nil, will be called to highlight the content of code blocks.
	HighlightCodeBlock func(info, code string) ui.Text
	// If non-nil, will be called for each relative link destination.
	ConvertRelativeLink func(dest string) string

	buf ui.TextBuilder

	// Current active container blocks. The punct field is not used; the
	// TTYCodec uses fixed punctuations for each type.
	containers stack[*fmtContainer]
	// Value of op.Type of the last Do call.
	lastOpType OpType
}

// Text returns the rendering result as a [ui.Text].
func (c *TTYCodec) Text() ui.Text { return c.buf.Text() }

// String returns the rendering result as a string with ANSI escape sequences.
func (c *TTYCodec) String() string { return c.buf.Text().String() }

// Do processes an Op.
func (c *TTYCodec) Do(op Op) {
	defer func() {
		c.lastOpType = op.Type
	}()
	if !c.buf.Empty() && op.Type != OpHTMLBlock && needNewStanza(op.Type, c.lastOpType) {
		c.writeLine("")
	}

	switch op.Type {
	case OpThematicBreak:
		c.writeLine("────")
	case OpHeading:
		c.startLine()
		c.writeStyled(ui.T(strings.Repeat("#", op.Number)+" ", ui.Bold))
		c.doInlineContent(op.Content, true)
		c.finishLine()
	case OpCodeBlock:
		if c.HighlightCodeBlock != nil {
			t := c.HighlightCodeBlock(op.Info, strings.Join(op.Lines, "\n")+"\n")
			lines := t.SplitByRune('\n')
			for i, line := range lines {
				if i == len(lines)-1 && len(line) == 0 {
					// If t ends in a newline, the newline terminates the
					// element before it; don't write an empty line for it.
					break
				}
				c.startLine()
				c.write("  ")
				c.writeStyled(line)
				c.finishLine()
			}
		} else {
			for _, line := range op.Lines {
				c.writeLine("  " + line)
			}
		}
	case OpHTMLBlock:
		// Do nothing
	case OpParagraph:
		c.startLine()
		c.doInlineContent(op.Content, false)
		c.finishLine()
	case OpBlockquoteStart:
		c.containers.push(&fmtContainer{typ: fmtBlockquote, marker: "│ "})
	case OpBlockquoteEnd:
		c.containers.pop()
	case OpListItemStart:
		if ct := c.containers.peek(); ct.typ == fmtBulletItem {
			ct.marker = "• "
		} else {
			ct.marker = fmt.Sprintf("%d. ", ct.number)
		}
	case OpListItemEnd:
		ct := c.containers.peek()
		ct.marker = ""
		ct.number++
	case OpBulletListStart:
		c.containers.push(&fmtContainer{typ: fmtBulletItem})
	case OpBulletListEnd:
		c.containers.pop()
	case OpOrderedListStart:
		c.containers.push(&fmtContainer{typ: fmtOrderedItem, number: op.Number})
	case OpOrderedListEnd:
		c.containers.pop()
	}
}

var absoluteDest = regexp.MustCompile(`^` + scheme + `:`)

func (c *TTYCodec) doInlineContent(ops []InlineOp, heading bool) {
	var stylings stack[ui.Styling]
	if heading {
		stylings.push(ui.Bold)
	}

	var (
		write         func(string)
		hardLineBreak func()
	)
	if heading || c.Width == 0 {
		write = func(s string) {
			c.writeStyled(ui.T(s, stylings...))
		}
		// When writing heading, ignore hard line break.
		//
		// When writing paragraph without reflowing, a hard line break will be
		// followed by an OpNewline, which will result in a line break.
		hardLineBreak = func() {}
	} else {
		maxWidth := c.Width
		for _, ct := range c.containers {
			maxWidth -= wcwidth.Of(ct.marker)
		}
		// The reflowing algorithm below is very similar to
		// [FmtCodec.writeSegmentsParagraphReflow], except that the step to
		// build spans and the step to arrange spans on lines are combined, and
		// the span is a ui.Text rather than a strings.Builder.
		currentLineWidth := 0
		var currentSpan ui.Text
		var prefixSpace ui.Text
		writeSpan := func(t ui.Text) {
			if len(t) == 0 {
				return
			}
			w := wcwidthOfText(t)
			if currentLineWidth == 0 {
				c.writeStyled(t)
				currentLineWidth = w
			} else if currentLineWidth+1+w <= maxWidth {
				c.writeStyled(prefixSpace)
				c.writeStyled(t)
				currentLineWidth += w + 1
			} else {
				c.finishLine()
				c.startLine()
				c.writeStyled(t)
				currentLineWidth = w
			}
		}
		write = func(s string) {
			parts := whitespaceRunRegexp.Split(s, -1)
			currentSpan = append(currentSpan, ui.T(parts[0], stylings...)...)
			if len(parts) > 1 {
				writeSpan(currentSpan)
				prefixSpace = ui.T(" ", stylings...)
				for _, s := range parts[1 : len(parts)-1] {
					writeSpan(ui.T(s, stylings...))
				}
				currentSpan = ui.T(parts[len(parts)-1], stylings...)
			}
		}
		hardLineBreak = func() {
			writeSpan(currentSpan)
			currentSpan = nil
			currentLineWidth = 0
			c.finishLine()
			c.startLine()
		}
		defer func() {
			writeSpan(currentSpan)
		}()
	}
	writeLinkDest := func(dest string) {
		show := absoluteDest.MatchString(dest)
		if !show && c.ConvertRelativeLink != nil {
			dest = c.ConvertRelativeLink(dest)
			show = dest != ""
		}
		if show {
			write(" (")
			write(dest)
			write(")")
		}
	}

	for _, op := range ops {
		switch op.Type {
		case OpText:
			write(op.Text)
		case OpRawHTML:
			switch op.Text {
			case "<kbd>":
				stylings.push(ui.Inverse)
			case "</kbd>":
				stylings.pop()
			}
		case OpNewLine:
			if heading || c.Width > 0 {
				write(" ")
			} else {
				c.finishLine()
				c.startLine()
			}
		case OpCodeSpan:
			stylings.push(ui.Underlined)
			write(op.Text)
			stylings.pop()
		case OpEmphasisStart:
			stylings.push(ui.Italic)
		case OpEmphasisEnd:
			stylings.pop()
		case OpStrongEmphasisStart:
			stylings.push(ui.Bold)
		case OpStrongEmphasisEnd:
			stylings.pop()
		case OpLinkStart:
			stylings.push(ui.Underlined)
		case OpLinkEnd:
			stylings.pop()
			writeLinkDest(op.Dest)
		case OpImage:
			write("Image: ")
			write(op.Alt)
			writeLinkDest(op.Dest)
		case OpAutolink:
			write(op.Text)
		case OpHardLineBreak:
			hardLineBreak()
		}
	}
}

func wcwidthOfText(t ui.Text) int {
	w := 0
	for _, seg := range t {
		w += wcwidth.Of(seg.Text)
	}
	return w
}

func (c *TTYCodec) startLine()            { startLine(c, c.containers) }
func (c *TTYCodec) writeLine(s string)    { writeLine(c, c.containers, s) }
func (c *TTYCodec) finishLine()           { c.write("\n") }
func (c *TTYCodec) write(s string)        { c.writeStyled(ui.T(s)) }
func (c *TTYCodec) writeStyled(t ui.Text) { c.buf.WriteText(t) }
