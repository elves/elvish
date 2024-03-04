package md

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"src.elv.sh/pkg/wcwidth"
)

// FmtCodec is a codec that formats Markdown in a specific style.
//
// The only supported configuration option is the text width.
//
// The formatted text uses the following style:
//
//   - Blocks are always separated by a blank line.
//
//   - Thematic breaks use "***" where possible, falling back to "---" if using
//     the former is problematic.
//
//   - Code blocks are always fenced, never indented.
//
//   - Code fences use backquotes (like "```") wherever possible, falling back
//     to "~~~" if using the former is problematic.
//
//   - Continuation markers of container blocks ("> " for blockquotes and spaces
//     for list items) are never omitted; in other words, lazy continuation is
//     never used.
//
//   - Blockquotes use "> ", never omitting the space.
//
//   - Bullet lists use "-" as markers where possible, falling back to "*" if
//     using the former is problematic.
//
//   - Ordered lists use "X." (X being a number) where possible, falling back to
//     "X)" if using the former is problematic.
//
//   - Bullet lists and ordered lists are indented 4 spaces where possible.
//
//   - Emphasis always uses "*".
//
//   - Strong emphasis always uses "**".
//
//   - Hard line break always uses an explicit "\".
type FmtCodec struct {
	Width int

	sb strings.Builder

	unsupported *FmtUnsupported

	// Current active container blocks.
	containers stack[*fmtContainer]
	// The value of sb.Len() when the last container block was started. Used to
	// determine whether a container is empty, in which case a blank line is
	// needed to preserve the container.
	containerStart int

	// The punctuation of the just popped list container, only populated if the
	// last Op was OpBulletListEnd or OpOrderedListEnd. Used to alternate list
	// punctuation when a list follows directly after another of the same type.
	poppedListPunct rune
	// Value of op.Type of the last Do call.
	lastOpType OpType
}

// FmtUnsupported contains information about use of unsupported features.
type FmtUnsupported struct {
	// Input contains emphasis or strong emphasis nested in another emphasis or
	// strong emphasis (not necessarily of the same type).
	NestedEmphasisOrStrongEmphasis bool
	// Input contains emphasis or strong emphasis that follows immediately after
	// another emphasis or strong emphasis (not necessarily of the same type).
	ConsecutiveEmphasisOrStrongEmphasis bool
}

func (c *FmtCodec) String() string { return c.sb.String() }

// Unsupported returns information about use of unsupported features that may
// make the output incorrect. It returns nil if there is no use of unsupported
// features.
func (c *FmtCodec) Unsupported() *FmtUnsupported { return c.unsupported }

func (c *FmtCodec) setUnsupported() *FmtUnsupported {
	if c.unsupported == nil {
		c.unsupported = &FmtUnsupported{}
	}
	return c.unsupported
}

var (
	backquoteRunRegexp = regexp.MustCompile("`+")
	tildeRunRegexp     = regexp.MustCompile("~+")
)

func (c *FmtCodec) Do(op Op) {
	var poppedListPunct rune
	defer func() {
		c.poppedListPunct = poppedListPunct
	}()

	if c.sb.Len() > 0 && needNewStanza(op.Type, c.lastOpType) {
		c.writeLine("")
	}
	defer func() {
		c.lastOpType = op.Type
	}()

	switch op.Type {
	case OpThematicBreak:
		if len(c.containers) > 0 && strings.TrimSpace(c.containers[len(c.containers)-1].marker) == "*" {
			// If the last marker to write is "*", using "***" will swallow the
			// marker.
			c.writeLine("---")
		} else {
			c.writeLine("***")
		}
	case OpHeading:
		c.startLine()
		c.write(strings.Repeat("#", op.Number) + " ")
		c.writeSegmentsATXHeading(c.buildSegments(op.Content))
		if op.Info != "" {
			c.write(" {" + op.Info + "}")
		}
		c.finishLine()
	case OpCodeBlock:
		startFence, endFence := codeFences(op.Info, op.Lines)
		c.writeLine(startFence)
		for _, line := range op.Lines {
			c.writeLine(line)
		}
		c.writeLine(endFence)
	case OpHTMLBlock:
		if c.lastOpType == OpListItemStart && strings.HasPrefix(op.Lines[0], " ") {
			// HTML blocks can contain 1 to 3 leading spaces. When it appears at
			// the first line of a list item, following "-   " or "*   ", those
			// spaces will either get merged into the list marker (in case of 1
			// leading space) or turn the HTML block into an indented code block
			// (in case of 2 or 3 leading spaces).
			//
			// To fix this, use a blank line as the first line, and start the
			// HTML block on the second line. The marker needs to be shortened
			// to contain exactly one trailing space, as is required by rule 3
			// in https://spec.commonmark.org/0.31.2/#list-items.
			//
			// Note that this only matters for HTML blocks. Indented code blocks
			// has the same behavior regarding leading spaces, but we always
			// turn them into fenced code blocks, moving the content to the
			// second line and avoiding this problem. Other types of blocks
			// either don't allow leading spaces, or don't preserve them.
			lastMarker := &c.containers[len(c.containers)-1].marker
			*lastMarker = strings.TrimRight(*lastMarker, " ") + " "
			c.writeLine("")
		}
		for _, line := range op.Lines {
			c.writeLine(line)
		}
	case OpParagraph:
		c.startLine()
		segs := c.buildSegments(op.Content)
		if c.Width > 0 {
			c.writeSegmentsParagraphReflow(segs, c.Width)
		} else {
			c.writeSegmentsParagraph(segs)
		}
		c.finishLine()
	case OpBlockquoteStart:
		c.containerStart = c.sb.Len()
		c.containers.push(&fmtContainer{typ: fmtBlockquote, marker: "> "})
	case OpBlockquoteEnd:
		if c.containerStart == c.sb.Len() {
			c.writeLine("")
		}
		c.containers.pop()
	case OpListItemStart:
		c.containerStart = c.sb.Len()
		// Set marker to start marker
		if ct := c.containers.peek(); ct.typ == fmtBulletItem {
			ct.marker = fmt.Sprintf("%c   ", ct.punct)
		} else {
			ct.marker = fmt.Sprintf("%d%c ", ct.number, ct.punct)
			if len(ct.marker) < 4 {
				ct.marker += strings.Repeat(" ", 4-len(ct.marker))
			}
		}
	case OpListItemEnd:
		if c.containerStart == c.sb.Len() {
			// When a list item is empty, we will write a line consisting of
			// bullet punctuations and spaces only. When there are at least 3
			// instances of the same punctuation, this line will be become a
			// thematic break instead. Avoid this by varying the punctuation.
			//
			// We use "-" whenever possible. If there are 3 consecutive
			// identical starter marks, they can only be all "-    ".
			for i := 2; i < len(c.containers); i++ {
				ct := c.containers[i]
				if allDashBullets(c.containers[i-2 : i+1]) {
					ct.punct = pickPunct('-', '*', ct.punct)
					ct.marker = fmt.Sprintf("%c   ", ct.punct)
				}
			}
			c.writeLine("")
		}
		ct := c.containers.peek()
		ct.marker = ""
		// If the current number is 9 9's, incrementing it will make the number
		// 10 digits; CommonMark requires the number in the ordered list to be
		// at most 9 digits. So just stop incrementing at this number.
		if ct.number < 999999999 {
			ct.number++
		}
	case OpBulletListStart:
		c.containers.push(&fmtContainer{
			typ:   fmtBulletItem,
			punct: pickPunct('-', '*', c.poppedListPunct)})
	case OpBulletListEnd:
		poppedListPunct = c.containers.pop().punct
	case OpOrderedListStart:
		c.containers.push(&fmtContainer{
			typ:    fmtOrderedItem,
			punct:  pickPunct('.', ')', c.poppedListPunct),
			number: op.Number})
	case OpOrderedListEnd:
		poppedListPunct = c.containers.pop().punct
	}
}

func needNewStanza(cur, last OpType) bool {
	switch cur {
	case OpThematicBreak, OpHeading, OpCodeBlock, OpHTMLBlock, OpParagraph,
		OpBlockquoteStart, OpBulletListStart, OpOrderedListStart:
		// Start of new block that does not coincide with the start of an outer
		// block.
		return last != OpBlockquoteStart && last != OpListItemStart
	case OpListItemStart:
		// A list item that is not the first in the list. The first item is
		// already handled when OpBulletListStart or OpOrderedListStart is seen.
		return last != OpBulletListStart && last != OpOrderedListStart
	}
	return false
}

func codeFences(info string, lines []string) (string, string) {
	var fenceRune rune
	var runLens map[int]bool
	if strings.ContainsRune(info, '`') {
		fenceRune = '~'
		runLens = matchLens(lines, tildeRunRegexp)
	} else {
		fenceRune = '`'
		runLens = matchLens(lines, backquoteRunRegexp)
	}
	l := 3
	for x := range runLens {
		if l < x+1 {
			l = x + 1
		}
	}
	fence := strings.Repeat(string(fenceRune), l)
	if fenceRune == '~' && strings.HasPrefix(info, "~") {
		return fence + " " + escapeCodeFenceInfo(info), fence
	}
	return fence + escapeCodeFenceInfo(info), fence
}

func escapeCodeFenceInfo(s string) string {
	// We could just use escapNewLines(escapeText(info)) here and be correct.
	// However, some of Elvish website's Markdown files embeds Markdown inside
	// the info string, and we'd like to leave them as is as much as possible,
	// so be more conservative in what we escape.
	//
	// The info string is mostly verbatim, with only support for backslashes and
	// entities, so we only need to escape \ and &. Additionally, the info
	// string can't contain newlines, so escape that too.
	if !strings.ContainsAny(s, "\\\n&") {
		return s
	}
	var sb strings.Builder
	for i, r := range s {
		switch r {
		case '\\':
			sb.WriteString(`\\`)
		case '\n':
			sb.WriteString("&NewLine;")
		case '&':
			if leadingCharRef(s[i:]) == "" {
				sb.WriteByte('&')
			} else {
				sb.WriteString(`\&`)
			}
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func allDashBullets(containers []*fmtContainer) bool {
	for _, ct := range containers {
		if ct.marker != "-   " {
			return false
		}
	}
	return true
}

// A segment is a unit of intermediate output when formatting inline content.
type segment struct {
	typ  segmentType
	text string
}

type segmentType uint

const (
	segText segmentType = iota
	// Some texts cannot be reflowed because whitespace is significant. This
	// includes code spans and link tails.
	segTextNoReflow
	segHTML
	segNewLine
	segHardLineBreak
	segLinkOrImageStart
	segLinkOrImageEnd
)

func (c *FmtCodec) buildSegments(ops []InlineOp) []segment {
	var segs []segment
	write := func(s string) {
		if s != "" {
			segs = append(segs, segment{typ: segText, text: s})
		}
	}
	writeNoReflow := func(s string) {
		segs = append(segs, segment{typ: segTextNoReflow, text: s})
	}

	emphasis := 0

	for i, op := range ops {
		switch op.Type {
		case OpText:
			text := op.Text
			if i > 0 && isEmphasisStart(ops[i-1]) {
				if r, l := utf8.DecodeRuneInString(text); l > 0 && unicode.IsSpace(r) {
					// Escape space immediately after emphasis start, since a *
					// before a space cannot open emphasis.
					write("&#" + strconv.Itoa(int(r)) + ";")
					text = text[l:]
				}
			} else if i > 1 && isEmphasisEnd(ops[i-1]) && emphasisOutputEndsWithPunct(ops[i-2]) {
				if r, l := utf8.DecodeRuneInString(text); isWord(r, l) {
					// Escape "other" (word character) immediately after
					// emphasis end if emphasis content ends with a punctuation.
					write("&#" + strconv.Itoa(int(r)) + ";")
					text = text[l:]
				}
			}

			suffix := ""
			if strings.HasSuffix(text, "!") && i < len(ops)-1 && ops[i+1].Type == OpLinkStart {
				text = text[:len(text)-1]
				suffix = `\!`
			} else if i < len(ops)-1 && isEmphasisEnd(ops[i+1]) {
				if r, l := utf8.DecodeLastRuneInString(text); l > 0 && unicode.IsSpace(r) {
					// Escape space immediately before emphasis end, since a *
					// after a space cannot close emphasis.
					text = text[:len(text)-l]
					suffix = "&#" + strconv.Itoa(int(r)) + ";"
				}
			} else if i < len(ops)-2 && isEmphasisStart(ops[i+1]) && emphasisOutputStartsWithPunct(ops[i+2]) {
				if r, l := utf8.DecodeLastRuneInString(text); isWord(r, l) {
					// Escape "other" (word character) immediately before
					// emphasis start if the output of the emphasis content will
					// start with a punctuation.
					text = text[:len(text)-l]
					suffix = "&#" + strconv.Itoa(int(r)) + ";"
				}
			}

			write(escapeText(text))
			write(suffix)
		case OpRawHTML:
			segs = append(segs, segment{typ: segHTML, text: op.Text})
		case OpNewLine:
			if i > 0 && isEmphasisStart(ops[i-1]) || i < len(ops)-1 && isEmphasisEnd(ops[i+1]) {
				write("&NewLine;")
			} else {
				segs = append(segs, segment{typ: segNewLine, text: op.Text})
			}
		case OpCodeSpan:
			text := op.Text
			hasRunWithLen := matchLens([]string{text}, backquoteRunRegexp)
			l := 1
			for hasRunWithLen[l] {
				l++
			}
			delim := strings.Repeat("`", l)
			// Code span text is never empty
			first := text[0]
			last := text[len(text)-1]
			addSpace := first == '`' || last == '`' || (first == ' ' && last == ' ' && strings.Trim(text, " ") != "")

			var sb strings.Builder
			sb.WriteString(delim)
			if addSpace {
				sb.WriteByte(' ')
			}
			sb.WriteString(text)
			if addSpace {
				sb.WriteByte(' ')
			}
			sb.WriteString(delim)
			segs = append(segs, segment{typ: segTextNoReflow, text: sb.String()})
		case OpEmphasisStart:
			write("*")
			emphasis++
			if emphasis >= 2 {
				c.setUnsupported().NestedEmphasisOrStrongEmphasis = true
			}
			if i > 0 && isEmphasisEnd(ops[i-1]) {
				c.setUnsupported().ConsecutiveEmphasisOrStrongEmphasis = true
			}
		case OpEmphasisEnd:
			write("*")
			emphasis--
		case OpStrongEmphasisStart:
			write("**")
			emphasis++
			if emphasis >= 2 {
				c.setUnsupported().NestedEmphasisOrStrongEmphasis = true
			}
			if i > 0 && isEmphasisEnd(ops[i-1]) {
				c.setUnsupported().ConsecutiveEmphasisOrStrongEmphasis = true
			}
		case OpStrongEmphasisEnd:
			write("**")
			emphasis--
		case OpLinkStart:
			segs = append(segs, segment{typ: segLinkOrImageStart})
			write("[")
		case OpLinkEnd:
			write("]")
			writeNoReflow(formatLinkTail(op.Dest, op.Text))
			segs = append(segs, segment{typ: segLinkOrImageEnd})
		case OpImage:
			segs = append(segs, segment{typ: segLinkOrImageStart})
			write("![")
			write(escapeNewLines(escapeText(op.Alt)))
			write("]")
			writeNoReflow(formatLinkTail(op.Dest, op.Text))
			segs = append(segs, segment{typ: segLinkOrImageEnd})
		case OpAutolink:
			write("<")
			if op.Dest == "mailto:"+op.Text {
				// Don't escape email autolinks. This is because the regexp that
				// matches email autolinks does not allow ";", so escaping them
				// makes the output no longer an email autolink.
				write(op.Text)
			} else {
				write(escapeAutolink(op.Text))
			}
			write(">")
		case OpHardLineBreak:
			segs = append(segs, segment{typ: segHardLineBreak})
		}
	}
	return segs
}

var atxHeadingCloserLookalike = regexp.MustCompile(`#+$`)

func (c *FmtCodec) writeSegmentsATXHeading(segs []segment) {
	for i, seg := range segs {
		switch seg.typ {
		case segText:
			text := seg.text
			if i == 0 {
				text = escapeLeadingSpaceTab(text)
			}
			if i == len(segs)-1 {
				text = escapeTrailingSpaceTab(text)
				if text[len(text)-1] == '#' {
					if hashes := atxHeadingCloserLookalike.FindString(text); hashes != "" {
						head := text[:len(text)-len(hashes)]
						if endsWithSpaceOrTab(head) || (head == "" && i == 0) {
							text = head + `\` + hashes
						}
					}
				}
			}
			c.write(text)
		case segHTML, segTextNoReflow:
			// Raw HTML in ATX headings and code spans never contain embedded
			// newlines, so just write them as is.
			c.write(seg.text)
		case segNewLine:
			c.write("&NewLine;")
		}
	}
}

func (c *FmtCodec) writeSegmentsParagraph(segs []segment) {
	for i := 0; i < len(segs); i++ {
		seg := segs[i]
		startOfLine := i == 0 || (segs[i-1].typ == segNewLine && (i-1 == 0 || segs[i-2].typ != segNewLine))
		endOfLine := i == len(segs)-1 || segs[i+1].typ == segNewLine
		switch seg.typ {
		case segText:
			text := seg.text
			// Escape trailing space or tab first. This way text like "- - - "
			// alone on a line will already have the trailing space escaped and
			// can no longer be parsed as a thematic break.
			if endOfLine {
				text = escapeTrailingSpaceTab(text)
			}
			if startOfLine {
				text = c.escapeStartOfLine(text, i == 0, endOfLine)
			}
			c.write(text)
		case segHTML:
			// Inline raw HTML may contain embedded newlines; write them
			// separately.
			lines := strings.Split(seg.text, "\n")
			if startOfLine && canStartHTMLBlock(lines[0], i == 0) {
				// If the first line appears at the start of the line, check
				// whether it can also be parsed as an HTML block instead.
				if i > 0 {
					// If the raw HTML appears not at the start of a paragraph,
					// inserting 4 spaces will prevent an HTML block to be
					// parsed, and won't make the text parse as an indented code
					// block, since the latter can't interrupt a paragraph.
					c.write("    ")
				} else if len(lines) == 1 && i+1 < len(segs) && segs[i+1].typ == segNewLine {
					// If raw HTML does appear at the start of a paragraph, the
					// only way I have found (actually the fuzz test found) for
					// the raw HTML to not have get parsed as an HTML block is
					// when the raw HTML is one line and is followed by an
					// escaped newline.
					c.write(lines[0])
					c.write("&NewLine;")
					i++
					continue
				}
			}
			c.write(lines[0])
			for _, line := range lines[1:] {
				c.finishLine()
				c.startLine()
				c.write(line)
			}
		case segTextNoReflow:
			c.write(seg.text)
		case segNewLine:
			if i == 0 || i == len(segs)-1 || segs[i-1].typ == segNewLine {
				c.write("&NewLine;")
			} else {
				c.finishLine()
				c.startLine()
			}
		case segHardLineBreak:
			c.write("\\")
		}
	}
}

var (
	whitespaceRunRegexp = regexp.MustCompile(`[ \t\n]+`)
)

func (c *FmtCodec) writeSegmentsParagraphReflow(segs []segment, maxWidth int) {
	// Rearrange the segments into spans with the following properties:
	//
	//  - Each span must be written as a whole, with no changes it its internal
	//    whitespaces.
	//
	//  - Adjacent spans must be separated by whitespaces.
	var spans []string
	var currentSpan strings.Builder
	finishCurrentSpan := func() {
		if currentSpan.Len() > 0 {
			spans = append(spans, currentSpan.String())
			currentSpan.Reset()
		}
	}
	linkOrImage := 0
	for _, seg := range segs {
		switch seg.typ {
		case segText:
			parts := whitespaceRunRegexp.Split(seg.text, -1)
			if linkOrImage > 0 {
				// Coalesce spaces between segments
				if parts[0] == "" && strings.HasSuffix(currentSpan.String(), " ") {
					parts = parts[1:]
				}
				currentSpan.WriteString(strings.Join(parts, " "))
			} else {
				for i, part := range parts {
					if i > 0 {
						finishCurrentSpan()
					}
					currentSpan.WriteString(part)
				}
			}
		case segTextNoReflow:
			currentSpan.WriteString(seg.text)
		case segHTML:
			// Replacing a whitespace run in raw HTML with a single space is not
			// always correct. For example, the following two are different:
			//
			//   <span title="a
			//   b">foo</span>
			//
			//   <span title="a b">foo</span>
			//
			// But the few uses of raw inline HTML in Elvish's Markdown are
			// simple enough (for example just a simple "<kbd>" tag), so we do
			// the wrong but easy thing here. The result is accepted by the
			// automated test, whose simplistic whitespace coalescing algorithm
			// will coalesce whitespaces within raw HTML.
			//
			// The correct way to handle this is to preserve newlines inside raw
			// HTML. But this will make some spans multi-line and complicate the
			// process to arrange the spans into lines.
			currentSpan.WriteString(
				whitespaceRunRegexp.ReplaceAllLiteralString(seg.text, " "))
		case segNewLine:
			if linkOrImage > 0 {
				if !strings.HasSuffix(currentSpan.String(), " ") {
					currentSpan.WriteString(" ")
				}
			} else {
				finishCurrentSpan()
			}
		case segHardLineBreak:
			if linkOrImage > 0 {
				// Keep links and images on the same line.
				currentSpan.WriteString("<br />")
			} else {
				// A span ending in "\\\n" is handled specifically below.
				currentSpan.WriteString("\\\n")
				finishCurrentSpan()
			}
		case segLinkOrImageStart:
			linkOrImage++
		case segLinkOrImageEnd:
			linkOrImage--
		}
	}
	finishCurrentSpan()

	if len(spans) == 0 {
		// If there are no spans left, write an ampersand-escaped newline to
		// preserve the paragraph. A run of ampersand-escaped whitespaces seems
		// to be the only way to create an empty paragraph in Markdown in the
		// first place.
		c.write("&NewLine;")
		return
	}

	for _, ct := range c.containers {
		maxWidth -= len(ct.marker)
	}

	var currentLine strings.Builder
	currentLineWidth := 0
	startOfParagraph := true
	writeCurrentLine := func() {
		escaped := c.escapeStartOfLine(currentLine.String(), startOfParagraph, true)
		if canStartHTMLBlock(escaped, startOfParagraph) {
			if startOfParagraph {
				escaped = "&NewLine;" + escaped
			} else {
				escaped = "    " + escaped
			}
		}
		if !startOfParagraph {
			c.startLine()
		}
		c.write(escaped)
	}
	startNewLine := func() {
		c.finishLine()
		currentLine.Reset()
		currentLineWidth = 0
		startOfParagraph = false
	}
	for i, span := range spans {
		// Only spans ending in a hard line break ends in a newline
		hardLineBreak := strings.HasSuffix(span, "\n")
		if hardLineBreak {
			span = span[:len(span)-1]
		}
		w := wcwidth.Of(span)
		if currentLine.Len() == 0 {
			currentLine.WriteString(span)
			currentLineWidth = w
		} else {
			// Determine whether the current span fits onto the current line.
			//
			// One slightly tricky detail here is that c.escapeStartOfLine may
			// insert more text, making the line wider. In reflow mode, the line
			// never starts or ends with whitespaces, so the most we have to
			// worry about is one backslash.
			//
			// As a result, if the line's width is exactly maxWidth after
			// appending the current span, we need to be extra careful and only
			// consider the current span to fit if c.escapeStartOfLine won't
			// introduce an additional backslash.
			//
			// The current implementation of this check is rather inefficient,
			// but since the check is done at most once per line, the
			// performance might as well be good enough.
			fits := false
			if currentLineWidth+1+w < maxWidth {
				fits = true
			} else if currentLineWidth+1+w == maxWidth {
				line := currentLine.String() + " " + span
				fits = c.escapeStartOfLine(line, startOfParagraph, true) == line
			}
			if fits {
				currentLine.WriteByte(' ')
				currentLine.WriteString(span)
				currentLineWidth += 1 + w
			} else {
				writeCurrentLine()
				startNewLine()
				currentLine.WriteString(span)
				currentLineWidth = w
			}
		}
		if hardLineBreak {
			writeCurrentLine()
			startNewLine()
			if i == len(spans)-1 {
				// \ at the end of a paragraph becomes a literal \ instead of a
				// hard line break. Fix this with an ampersand-escaped
				// whitespace, which seems to be the only way to make a
				// paragraph end with a hard line break in the first place.
				currentLine.WriteString("&NewLine;")
			}
		}
	}
	if currentLine.Len() > 0 {
		writeCurrentLine()
	}
}

var (
	// Pattern for text that can be parsed as thematic break, possibly after
	// prepending the some bullet markers.
	//
	// - We don't need to consider leading spaces, since they will already be
	//   ampersand-escaped.
	//
	// - We don't need to consider "*", since it is always backslash-escaped.
	thematicBreakLookalike = regexp.MustCompile(`^((?:-[ \t]*)+|(?:_[ \t]*)+)$`)
	// Pattern for dash bullets at the end of the buffer.
	trailingDashes = regexp.MustCompile(`(?:- *)*$`)
	// Pattern for text that can be parsed as an ATX heading opener, if followed
	// by space, tab or end of line.
	atxHeadingOpenerLookalike = regexp.MustCompile(`^#{1,6}`)
	// Pattern for text that can be parsed as an ordered list opener, if
	// followed by space, tab or end of line.
	orderedListOpenerLookalike = regexp.MustCompile(`^([0-9]{1,9})([.)])`)
)

func (c *FmtCodec) escapeStartOfLine(s string, startOfParagraph, endOfLine bool) string {
	s = escapeLeadingSpaceTab(s)
	switch s[0] {
	case '-', '+':
		tail := s[1:]
		if startsWithSpaceOrTab(tail) || (tail == "" && startOfParagraph && endOfLine) {
			return `\` + s
		}
	case '>':
		return `\` + s
	case '#':
		if hashes := atxHeadingOpenerLookalike.FindString(s); hashes != "" {
			tail := s[len(hashes):]
			if startsWithSpaceOrTab(tail) || (tail == "" && endOfLine) {
				return `\` + s
			}
		}
	}
	if strings.HasPrefix(s, "~~~") {
		return `\` + s
	} else if m := orderedListOpenerLookalike.FindStringSubmatch(s); m != nil {
		tail := s[len(m[0]):]
		if startsWithSpaceOrTab(tail) || (tail == "" && endOfLine) {
			number, punct := m[1], m[2]
			if startOfParagraph || strings.TrimLeft(number, "0") == "1" {
				return number + `\` + punct + tail
			}
		}
	} else if endOfLine && thematicBreakLookalike.MatchString(s) {
		// If a line contains a single segment, there is a danger for
		// the text to be parsed as a thematic break.
		//
		// After the escaping above, the text cannot start of end with a
		// space or tab; the thematicBreakLookalikeRegexp match furthers
		// guarantees that the text starts with either "-" or "_".
		line := s
		if startOfParagraph && s[0] == '-' {
			// If we are the start of a paragraph, we also need to include
			// bullet markers that can be merged with the text to form a
			// thematic break.
			//
			// This can only happen for "-": "*" in the content is already
			// backslash-escaped at this point, while "_" is not a possible
			// bullet list marker.
			line = trailingDashes.FindString(c.sb.String()) + line
		}
		if thematicBreakRegexp.MatchString(line) {
			return `\` + s
		}
	}
	return s
}

// Whether an inline raw HTML element can be parsed as the first line of an HTML
// block.
func canStartHTMLBlock(s string, startOfParagraph bool) bool {
	return strings.HasPrefix(s, "<") && (html1Regexp.MatchString(s) ||
		html2Regexp.MatchString(s) ||
		html3Regexp.MatchString(s) ||
		html4Regexp.MatchString(s) ||
		html5Regexp.MatchString(s) ||
		html6Regexp.MatchString(s) ||
		html7Regexp.MatchString(s) && startOfParagraph)
}

func escapeLeadingSpaceTab(s string) string {
	switch s[0] {
	case ' ':
		return "&#32;" + s[1:]
	case '\t':
		return "&Tab;" + s[1:]
	}
	return s
}

func escapeTrailingSpaceTab(s string) string {
	switch s[len(s)-1] {
	case ' ':
		return s[:len(s)-1] + "&#32;"
	case '\t':
		return s[:len(s)-1] + "&Tab;"
	}
	return s
}

func startsWithSpaceOrTab(s string) bool {
	return s != "" && (s[0] == ' ' || s[0] == '\t')
}

func endsWithSpaceOrTab(s string) bool {
	return s != "" && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t')
}

func emphasisOutputStartsWithPunct(op InlineOp) bool {
	switch op.Type {
	case OpText:
		r, l := utf8.DecodeRuneInString(op.Text)
		// If the content starts with a space, it will be escaped into "&#32;"
		return l > 0 && unicode.IsSpace(r) || isUnicodePunct(r)
	default:
		return true
	}
}

func emphasisOutputEndsWithPunct(op InlineOp) bool {
	switch op.Type {
	case OpText:
		r, l := utf8.DecodeLastRuneInString(op.Text)
		// If the content starts with a space, it will be escaped into "&#32;"
		return l > 0 && unicode.IsSpace(r) || isUnicodePunct(r)
	default:
		return true
	}
}

func matchLens(pieces []string, pattern *regexp.Regexp) map[int]bool {
	hasRunWithLen := make(map[int]bool)
	for _, piece := range pieces {
		for _, run := range pattern.FindAllString(piece, -1) {
			hasRunWithLen[len(run)] = true
		}
	}
	return hasRunWithLen
}

const asciiControl = "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f"

const forbiddenInRawLinkDest = asciiControl + " "

func formatLinkTail(dest, title string) string {
	var sb strings.Builder
	sb.WriteString("(")
	if strings.ContainsAny(dest, forbiddenInRawLinkDest) || !balancedParens(dest) {
		// Angle-bracketed destinations recognize a few characters plus
		// character references as special and disallow newlines. The order of
		// function calls is important here to avoid double-escaping.
		sb.WriteString("<" + strings.ReplaceAll(
			escapeAmpersandBackslash(dest, "<>"), "\n", "&NewLine;") + ">")
	} else if dest == "" && title != "" {
		sb.WriteString("<>")
	} else {
		// Bare destinations only recognize backslash and character references
		// as special. The order of function calls is important here to avoid
		// double-escaping.
		escapedDest := escapeAmpersandBackslash(dest, "")
		// Also escape any leading < so that it won't be parsed as an
		// angle-bracketed destination.
		if strings.HasPrefix(escapedDest, "<") {
			escapedDest = `\` + escapedDest
		}
		sb.WriteString(escapedDest)
	}
	if title != "" {
		sb.WriteString(" ")
		sb.WriteString(escapeNewLines(wrapAndEscapeLinkTitle(title)))
	}
	sb.WriteString(")")
	return sb.String()
}

func balancedParens(s string) bool {
	balance := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(':
			balance++
		case ')':
			if balance == 0 {
				return false
			}
			balance--
		}
	}
	return balance == 0
}

func wrapAndEscapeLinkTitle(title string) string {
	doubleQuotes := strings.Count(title, "\"")
	if doubleQuotes == 0 {
		return "\"" + escapeAmpersandBackslash(title, "") + "\""
	}
	singleQuotes := strings.Count(title, "'")
	if singleQuotes == 0 {
		return "'" + escapeAmpersandBackslash(title, "") + "'"
	}
	parens := strings.Count(title, "(") + strings.Count(title, ")")
	if parens == 0 {
		return "(" + escapeAmpersandBackslash(title, "") + ")"
	}
	switch {
	case doubleQuotes <= singleQuotes && doubleQuotes <= parens:
		return `"` + escapeAmpersandBackslash(title, `"`) + `"`
	case singleQuotes <= parens:
		return "'" + escapeAmpersandBackslash(title, `'`) + "'"
	default:
		return "(" + escapeAmpersandBackslash(title, "()") + ")"
	}
}

// Backslash-escape ampersands, backslashes and bytes in the specified set.
func escapeAmpersandBackslash(s, set string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' || strings.IndexByte(set, s[i]) >= 0 || leadingCharRef(s[i:]) != "" {
			sb.WriteByte('\\')
		}
		sb.WriteByte(s[i])
	}
	return sb.String()
}

func (c *FmtCodec) startLine()         { startLine(c, c.containers) }
func (c *FmtCodec) writeLine(s string) { writeLine(c, c.containers, s) }
func (c *FmtCodec) finishLine()        { c.write("\n") }
func (c *FmtCodec) write(s string)     { c.sb.WriteString(s) }

type writer interface{ write(string) }

func startLine(w writer, containers stack[*fmtContainer]) {
	for _, container := range containers {
		w.write(container.useMarker())
	}
}

func writeLine(w writer, containers stack[*fmtContainer], s string) {
	if s == "" {
		// When writing a blank line, trim trailing spaces from the markers.
		//
		// This duplicates startLine, but merges the markers for ease of
		// trimming.
		var markers strings.Builder
		for _, container := range containers {
			markers.WriteString(container.useMarker())
		}
		w.write(strings.TrimRight(markers.String(), " "))
		w.write("\n")
		return
	}
	startLine(w, containers)
	w.write(s)
	w.write("\n")
}

type fmtContainer struct {
	typ    fmtContainerType
	punct  rune   // punctuation used to build the marker
	number int    // only used when typ == fmtOrderedItem
	marker string // starter or continuation marker
}

type fmtContainerType uint

const (
	fmtBlockquote fmtContainerType = iota
	fmtBulletItem
	fmtOrderedItem
)

func (ct *fmtContainer) useMarker() string {
	m := ct.marker
	if ct.typ != fmtBlockquote {
		ct.marker = strings.Repeat(" ", wcwidth.Of(m))
	}
	return m
}

func pickPunct(def, alt, banned rune) rune {
	if def != banned {
		return def
	}
	return alt
}

func isEmphasisStart(op InlineOp) bool {
	return op.Type == OpEmphasisStart || op.Type == OpStrongEmphasisStart
}

func isEmphasisEnd(op InlineOp) bool {
	return op.Type == OpEmphasisEnd || op.Type == OpStrongEmphasisEnd
}

func escapeNewLines(s string) string { return strings.ReplaceAll(s, "\n", "&NewLine;") }

func escapeText(s string) string {
	if !strings.ContainsAny(s, "[]*_`\\&<>\u00A0") {
		return s
	}
	var sb strings.Builder
	for i, r := range s {
		switch r {
		case '[', ']', '*', '`', '\\':
			sb.WriteByte('\\')
			sb.WriteRune(r)
		case '_':
			if isWord(utf8.DecodeLastRuneInString(s[:i])) && isWord(utf8.DecodeRuneInString(s[i+1:])) {
				sb.WriteByte('_')
			} else {
				sb.WriteString(`\_`)
			}
		case '&':
			// Look ahead decide whether the ampersand can start a character
			// reference and thus needs to be escaped. Since any inline markup
			// will introduce a metacharacter that is not allowed within
			// character reference, it is sufficient to check within the text.
			if leadingCharRef(s[i:]) == "" {
				sb.WriteByte('&')
			} else {
				sb.WriteString(`\&`)
			}
		case '<':
			if i < len(s)-1 && !canBeSpecialAfterLt(s[i+1]) {
				sb.WriteByte('<')
			} else {
				sb.WriteString(`\<`)
			}
		case '\u00A0':
			// This is by no means required, but it's nice to make non-breaking
			// spaces explicit.
			sb.WriteString("&nbsp;")
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

const forbiddenInAutolink = asciiControl + "& <>"

// The escape of autolinks need to be handled specifically, because they support
// character references, but don't support backslashes. Moreover, characters
// forbidden inside autolinks (see uriAutolinkRegexp) should also be escaped.
func escapeAutolink(s string) string {
	if !strings.ContainsAny(s, forbiddenInAutolink) {
		return s
	}
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] <= 0x20 {
			sb.WriteString("&#" + strconv.Itoa(int(s[i])) + ";")
		} else if s[i] == '&' {
			if leadingCharRef(s[i:]) == "" {
				sb.WriteByte('&')
			} else {
				sb.WriteString("&amp;")
			}
		} else if s[i] == '<' {
			sb.WriteString("&lt;")
		} else if s[i] == '>' {
			sb.WriteString("&gt;")
		} else {
			sb.WriteByte(s[i])
		}
	}
	return sb.String()
}

// Takes the result of utf8.Decode*, and returns whether the character is
// non-empty and a "word" character for the purpose of emphasis parsing.
func isWord(r rune, l int) bool {
	return l > 0 && !unicode.IsSpace(r) && !isUnicodePunct(r)
}

func canBeSpecialAfterLt(b byte) bool {
	return /* Can form raw HTML */ b == '!' || b == '?' || b != '/' || isASCIILetter(b) ||
		/* Can form email autolink */ '0' <= b && b <= '9' || strings.IndexByte(emailLocalPuncts, b) >= 0
}
