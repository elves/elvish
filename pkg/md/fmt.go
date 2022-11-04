package md

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
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
//   - Code blocks are always fenced, never idented.
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
//
//   - Internal spaces in paragraphs are coalesced.
type FmtCodec struct {
	pieces []string
	Width  string

	unsupported *FmtUnsupported

	// Current active container blocks.
	containers stack[*fmtContainer]
	// The value of len(pieces) when the last container block was started. Used
	// to determine whether a container is empty, in which case a blank line is
	// needed to preserve the container.
	containerStart int

	// The punctuation of the just popped list container, only populated if the
	// last Op was OpBulletListEnd or OpOrderedListEnd. Used to alternate list
	// punctuation when a list follows directly after another of the same type.
	poppedListPunct rune
	// Whether a new stanza was already started.
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

func (c *FmtCodec) String() string { return strings.Join(c.pieces, "") }

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
	// Unlike escapeHTML, double and single quotes in autolinks do not need
	// escaping.
	escapeAutolink = strings.NewReplacer(
		"&", "&amp;", "<", "&lt;", ">", "&gt;",
	).Replace
)

var (
	backquoteRunRegexp = regexp.MustCompile("`+")
	tildeRunRegexp     = regexp.MustCompile("~+")
)

func (c *FmtCodec) Do(op Op) {
	var poppedListPunct rune
	defer func() {
		c.poppedListPunct = poppedListPunct
	}()

	switch op.Type {
	case OpThematicBreak, OpHeading, OpCodeBlock, OpHTMLBlock, OpParagraph,
		OpBlockquoteStart, OpBulletListStart, OpOrderedListStart:
		if len(c.pieces) > 0 && c.lastOpType != OpBlockquoteStart && c.lastOpType != OpListItemStart {
			c.writeLine("")
		}
	case OpListItemStart:
		if len(c.pieces) > 0 && c.lastOpType != OpBulletListStart && c.lastOpType != OpOrderedListStart {
			c.writeLine("")
		}
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
		c.doInlineContent(op.Content, true)
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
			// in https://spec.commonmark.org/0.30/#list-items.
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
		c.doInlineContent(op.Content, false)
		c.finishLine()
	case OpBlockquoteStart:
		c.containerStart = len(c.pieces)
		c.containers.push(&fmtContainer{typ: fmtBlockquote, marker: "> "})
	case OpBlockquoteEnd:
		if c.containerStart == len(c.pieces) {
			c.writeLine("")
		}
		c.containers.pop()
	case OpListItemStart:
		c.containerStart = len(c.pieces)
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
		if c.containerStart == len(c.pieces) {
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
		ct.number++
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

func escapeCodeFenceInfo(info string) string {
	// Info strings of code fences are terminated by newlines, so newlines also
	// need to be escaped in addition.
	return strings.ReplaceAll(escapeText(info), "\n", "&NewLine;")
}

func allDashBullets(containers []*fmtContainer) bool {
	for _, ct := range containers {
		if ct.marker != "-   " {
			return false
		}
	}
	return true
}

var (
	leadingHashesRegexp  = regexp.MustCompile(`^#{1,6}`)
	trailingHashesRegexp = regexp.MustCompile(`#+$`)
	leadingNumberRegexp  = regexp.MustCompile(`^([0-9]{1,9})([.)])`)
)

func (c *FmtCodec) doInlineContent(ops []InlineOp, atxHeading bool) {
	emphasis := 0

	for i, op := range ops {
		startOfLine := i == 0 || (ops[i-1].Type == OpNewLine && (i-1 == 0 || ops[i-2].Type != OpNewLine))
		endOfLine := i == len(ops)-1 || ops[i+1].Type == OpNewLine
		prevIsEmphasisEnd := i > 0 && isEmphasisEnd(ops[i-1])

		switch op.Type {
		case OpText:
			text := op.Text
			if startOfLine && endOfLine && !atxHeading && !startsWithSpaceOrTab(text) {
				// If a line contains a single OpText, there is a danger for the
				// text to be parsed as a thematic break.
				//
				// The case where the text starts with a space or a tab is
				// handled elsewhere, so we can assume that the text does not
				// start with space or tab.
				line := text
				if i == 0 {
					// If we are the very beginning of the paragraph, we also
					// need to include bullet markers that can be merged with
					// the text to form a thematic break.
					//
					// The condition checking the first byte will also match
					// markers like "1." if the text starts with "1", but since
					// that will never match a thematic break, it doesn't
					// matter. It will also cause the loop to end at newline.
					//
					// The code here depends on the fact that bullet markers are
					// written as individual pieces. This is guaranteed by the
					// startLine method.
					for j := len(c.pieces) - 1; j >= 0 && c.pieces[j][0] == text[0]; j-- {
						line = c.pieces[j] + line
					}
				}
				if thematicBreakRegexp.MatchString(line) {
					switch text[len(text)-1] {
					case ' ':
						// If the line ends with a space, it has to be escaped
						// to be preserved. This has the side effect of making
						// the line no longer parse as thematic break.
						c.write(escapeText(text[:len(text)-1]))
						c.write("&#32;")
					case '\t':
						// Same for lines ending with a tab.
						c.write(escapeText(text[:len(text)-1]))
						c.write("&Tab;")
					default:
						if escaped := escapeText(text); escaped != text {
							// If the text needs escaping anyway, the escaping
							// anything extra.
							c.write(escaped)
						} else {
							// Otherwise, escape the first byte, which must be a
							// punctuation at this point.
							c.write(`\`)
							c.write(text[:1])
							c.write(escapeText(text[1:]))
						}
					}
					continue
				}
			}
			if startOfLine {
				switch text[0] {
				case ' ':
					c.write("&#32;")
					text = text[1:]
				case '\t':
					c.write("&Tab;")
					text = text[1:]
				case '-', '+':
					if !atxHeading {
						tail := text[1:]
						if startsWithSpaceOrTab(tail) || (tail == "" && endOfLine) {
							c.write(`\` + text[:1])
							text = tail
						}
					}
				case '>':
					if !atxHeading {
						c.write(`\>`)
						text = text[1:]
					}
				case '#':
					if !atxHeading {
						if hashes := leadingHashesRegexp.FindString(text); hashes != "" {
							tail := text[len(hashes):]
							if startsWithSpaceOrTab(tail) || (tail == "" && endOfLine) {
								c.write(`\` + hashes)
								text = tail
							}
						}
					}
				default:
					if !atxHeading {
						if m := leadingNumberRegexp.FindStringSubmatch(text); m != nil {
							tail := text[len(m[0]):]
							if startsWithSpaceOrTab(tail) || (tail == "" && endOfLine) {
								number, punct := m[1], m[2]
								if i == 0 || strings.TrimLeft(number, "0") == "1" {
									c.write(number)
									c.write(`\` + punct)
									text = tail
								}
							}
						} else if strings.HasPrefix(text, "~~~") {
							c.write(`\~~~`)
							text = text[3:]
						}
					}
				}
			} else if i > 0 && isEmphasisStart(ops[i-1]) {
				if r, l := utf8.DecodeRuneInString(text); l > 0 && unicode.IsSpace(r) {
					// Escape space immediately after emphasis start, since a *
					// before a space cannot open emphasis.
					c.write("&#" + strconv.Itoa(int(r)) + ";")
					text = text[l:]
				}
			} else if i > 1 && isEmphasisEnd(ops[i-1]) && emphasisOutputEndsWithPunct(ops[i-2]) {
				if r, l := utf8.DecodeRuneInString(text); isWord(r, l) {
					// Escape "other" (word character) immediately after
					// emphasis end if emphasis content ends with a punctuation.
					c.write("&#" + strconv.Itoa(int(r)) + ";")
					text = text[l:]
				}
			}

			suffix := ""
			if endOfLine && text != "" {
				switch text[len(text)-1] {
				case ' ':
					suffix = "&#32;"
					text = text[:len(text)-1]
				case '\t':
					suffix = "&Tab;"
					text = text[:len(text)-1]
				case '#':
					if atxHeading {
						if hashes := trailingHashesRegexp.FindString(text); hashes != "" {
							head := text[:len(text)-len(hashes)]
							if endsWithSpaceOrTab(head) || (head == "" && i == 0) {
								text = head
								suffix = `\` + hashes
							}
						}
					}
				}
			} else if strings.HasSuffix(text, "!") && i < len(ops)-1 && ops[i+1].Type == OpLinkStart {
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

			c.write(escapeText(text))
			c.write(suffix)
		case OpRawHTML:
			// Inline raw HTML may contain embedded newlines; write them
			// separately.
			lines := strings.Split(op.Text, "\n")
			if startOfLine && i > 0 && strings.HasPrefix(op.Text, "<") {
				// If the first line appears at the start of the line, check
				// whether it can also be parsed as an HTML block interrupting a
				// paragraph (type 1 to 6). The only way I have found to prevent
				// this is to make sure that it starts with at least 4 spaces;
				// in fact, this exact case came up in a fuzz test where inline
				// raw HTML appears at the start of the second line, preceded by
				// 4 spaces. This won't be parsed as an indented code block
				// since the latter can't interrupt a paragraph, but will
				// prevent an HTML block to be parsed.
				//
				// If this piece of inline raw HTML appears at the very start,
				// it means it can't be parsed as an HTML block, so there is no
				// need to prevent it.
				for _, r := range []*regexp.Regexp{html1Regexp, html2Regexp, html3Regexp, html4Regexp, html5Regexp, html6Regexp} {
					if r.MatchString(lines[0]) {
						c.write("    ")
						break
					}
				}
			}
			c.write(lines[0])
			for _, line := range lines[1:] {
				c.finishLine()
				c.startLine()
				c.write(line)
			}
		case OpNewLine:
			if atxHeading || i == 0 || i == len(ops)-1 || ops[i-1].Type == OpNewLine || isEmphasisStart(ops[i-1]) || isEmphasisEnd(ops[i+1]) {
				c.write("&NewLine;")
			} else {
				c.finishLine()
				c.startLine()
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
			c.write(delim)
			if addSpace {
				c.write(" ")
			}
			c.write(text)
			if addSpace {
				c.write(" ")
			}
			c.write(delim)
		case OpEmphasisStart:
			c.write("*")
			emphasis++
			if emphasis >= 2 {
				c.setUnsupported().NestedEmphasisOrStrongEmphasis = true
			}
			if prevIsEmphasisEnd {
				c.setUnsupported().ConsecutiveEmphasisOrStrongEmphasis = true
			}
		case OpEmphasisEnd:
			c.write("*")
			emphasis--
		case OpStrongEmphasisStart:
			c.write("**")
			emphasis++
			if emphasis >= 2 {
				c.setUnsupported().NestedEmphasisOrStrongEmphasis = true
			}
			if prevIsEmphasisEnd {
				c.setUnsupported().ConsecutiveEmphasisOrStrongEmphasis = true
			}
		case OpStrongEmphasisEnd:
			c.write("**")
			emphasis--
		case OpLinkStart:
			c.write("[")
		case OpLinkEnd:
			c.write("]")
			c.writeLinkTail(op.Dest, op.Text)
		case OpImage:
			c.write("![")
			c.write(escapeNewLines(escapeText(op.Alt)))
			c.write("]")
			c.writeLinkTail(op.Dest, op.Text)
		case OpAutolink:
			c.write("<")
			if op.Dest == "mailto:"+op.Text {
				// Don't escape email autolinks. This is because the regexp that
				// matches email autolinks does not allow ";", so escaping them
				// makes the output no longer an email autolink.
				c.write(op.Text)
			} else {
				c.write(escapeAutolink(op.Text))
			}
			c.write(">")
		case OpHardLineBreak:
			c.write("\\")
		}
	}
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

const asciiControlOrSpaceOrParens = "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f ()"

func (c *FmtCodec) writeLinkTail(dest, title string) {
	c.write("(")
	if strings.ContainsAny(dest, asciiControlOrSpaceOrParens) {
		c.write("<" + strings.ReplaceAll(escapeText(dest), ">", "&gt;") + ">")
	} else if dest == "" && title != "" {
		c.write("<>")
	} else {
		c.write(escapeText(dest))
	}
	if title != "" {
		c.write(" ")
		c.write(escapeNewLines(wrapAndEscapeLinkTitle(title)))
	}
	c.write(")")
}

var escapeParens = strings.NewReplacer("(", `\(`, ")", `\)`).Replace

func wrapAndEscapeLinkTitle(title string) string {
	doubleQuotes := strings.Count(title, "\"")
	if doubleQuotes == 0 {
		return "\"" + escapeText(title) + "\""
	}
	singleQuotes := strings.Count(title, "'")
	if singleQuotes == 0 {
		return "'" + escapeText(title) + "'"
	}
	parens := strings.Count(title, "(") + strings.Count(title, ")")
	if parens == 0 {
		return "(" + escapeText(title) + ")"
	}
	switch {
	case doubleQuotes <= singleQuotes && doubleQuotes <= parens:
		return `"` + strings.ReplaceAll(escapeText(title), `"`, `\"`) + `"`
	case singleQuotes <= parens:
		return "'" + strings.ReplaceAll(escapeText(title), "'", `\'`) + "'"
	default:
		return "(" + escapeParens(escapeText(title)) + ")"
	}
}

func (c *FmtCodec) startLine() {
	// Note: the fact that the container markers are written as individual
	// pieces is depended on by the part of doInlineContent escaping texts that
	// look like thematic breaks.
	for _, container := range c.containers {
		c.write(container.useMarker())
	}
}

func (c *FmtCodec) finishLine() {
	c.write("\n")
}

func (c *FmtCodec) writeLine(s string) {
	if s == "" {
		// When writing a blank line, trim trailing spaces from the markers.
		//
		// This duplicates startLine, but merges the markers for ease of
		// trimming.
		var markers strings.Builder
		for _, container := range c.containers {
			markers.WriteString(container.useMarker())
		}
		c.write(strings.TrimRight(markers.String(), " "))
		c.finishLine()
		return
	}
	c.startLine()
	c.write(s)
	c.finishLine()
}

func (c *FmtCodec) write(s string) {
	c.pieces = append(c.pieces, s)
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
		ct.marker = strings.Repeat(" ", len(m))
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
				sb.WriteString("\\_")
			}
		case '&':
			// Look ahead to next ";" to decide whether the ampersand can start
			// a character reference and thus needs to be escaped. Since
			// any inline markup will introduce a metacharacter that is not
			// allowed within character reference, it is sufficient to check
			// within the text.
			if end := strings.IndexByte(s[i:], ';'); end == -1 || !charRefRegexp.MatchString(s[i:i+end+1]) {
				sb.WriteByte('&')
			} else {
				sb.WriteString("&amp;")
			}
		case '<':
			if i < len(s)-1 && !canBeSpecialAfterLt(s[i+1]) {
				sb.WriteByte('<')
			} else {
				sb.WriteString("&lt;")
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

// Takes the result of utf8.Decode*, and returns whether the character is
// non-empty and a "word" character for the purpose of emphasis parsing.
func isWord(r rune, l int) bool {
	return l > 0 && !unicode.IsSpace(r) && !isUnicodePunct(r)
}

func canBeSpecialAfterLt(b byte) bool {
	return /* Can form raw HTML */ b == '!' || b == '?' || b != '/' || isASCIILetter(b) ||
		/* Can form email autolink */ '0' <= b && b <= '9' || strings.IndexByte(emailLocalPuncts, b) >= 0
}
