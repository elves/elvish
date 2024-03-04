package md

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// InlineOp represents an inline operation.
type InlineOp struct {
	Type InlineOpType
	// OpText, OpCodeSpan, OpRawHTML, OpAutolink: Text content
	// OpLinkStart, OpLinkEnd, OpImage: title text
	Text string
	// OpLinkStart, OpLinkEnd, OpImage, OpAutolink
	Dest string
	// ForOpImage
	Alt string
}

// InlineOpType enumerates possible types of an InlineOp.
type InlineOpType uint

const (
	// Text elements. Embedded newlines in OpText are turned into OpNewLine, but
	// OpRawHTML can contain embedded newlines. OpCodeSpan never contains
	// embedded newlines.
	OpText InlineOpType = iota
	OpCodeSpan
	OpRawHTML
	OpNewLine

	// Inline markup elements.
	OpEmphasisStart
	OpEmphasisEnd
	OpStrongEmphasisStart
	OpStrongEmphasisEnd
	OpLinkStart
	OpLinkEnd
	OpImage
	OpAutolink
	OpHardLineBreak
)

// String returns the text content of the InlineOp
func (op InlineOp) String() string {
	switch op.Type {
	case OpText, OpCodeSpan, OpRawHTML, OpAutolink:
		return op.Text
	case OpNewLine:
		return "\n"
	case OpImage:
		return op.Alt
	}
	return ""
}

func renderInline(text string) []InlineOp {
	p := inlineParser{text, 0, makeDelimStack(), buffer{}}
	p.render()
	return p.buf.ops()
}

type inlineParser struct {
	text   string
	pos    int
	delims delimStack
	buf    buffer
}

func (p *inlineParser) render() {
	for p.pos < len(p.text) {
		b := p.text[p.pos]
		begin := p.pos
		p.pos++

		parseText := func() {
			for p.pos < len(p.text) && !isMeta(p.text[p.pos]) {
				p.pos++
			}
			text := p.text[begin:p.pos]
			hardLineBreak := false
			if p.pos < len(p.text) && p.text[p.pos] == '\n' {
				// https://spec.commonmark.org/0.31.2/#hard-line-break
				//
				// The input to renderInline never ends in a newline, so all
				// newlines are internal ones, thus subject to the hard line
				// break rules
				hardLineBreak = strings.HasSuffix(text, "  ")
				text = strings.TrimRight(text, " ")
			}
			p.buf.push(textPiece(text))
			if hardLineBreak {
				p.buf.push(piece{main: InlineOp{Type: OpHardLineBreak}})
			}
		}

		switch b {
		// The 3 branches below implement the first part of
		// https://spec.commonmark.org/0.31.2/#an-algorithm-for-parsing-nested-emphasis-and-links.
		case '[':
			bufIdx := p.buf.push(textPiece("["))
			p.delims.push(&delim{typ: '[', bufIdx: bufIdx})
		case '!':
			if p.pos < len(p.text) && p.text[p.pos] == '[' {
				p.pos++
				bufIdx := p.buf.push(textPiece("!["))
				p.delims.push(&delim{typ: '!', bufIdx: bufIdx})
			} else {
				parseText()
			}
		case '*', '_':
			p.consumeRun(b)
			canOpen, canClose := canOpenCloseEmphasis(rune(b),
				emptyToNewline(utf8.DecodeLastRuneInString(p.text[:begin])),
				emptyToNewline(utf8.DecodeRuneInString(p.text[p.pos:])))
			bufIdx := p.buf.push(textPiece(p.text[begin:p.pos]))
			p.delims.push(
				&delim{typ: b, bufIdx: bufIdx,
					n: p.pos - begin, canOpen: canOpen, canClose: canClose})
		case ']':
			// https://spec.commonmark.org/0.31.2/#look-for-link-or-image.
			var opener *delim
			for d := p.delims.top.prev; d != p.delims.bottom; d = d.prev {
				if d.typ == '[' || d.typ == '!' {
					opener = d
					break
				}
			}
			if opener == nil || opener.inactive {
				if opener != nil {
					unlink(opener)
				}
				p.buf.push(textPiece("]"))
				continue
			}
			n, dest, title := parseLinkTail(p.text[p.pos:])
			if n == -1 {
				unlink(opener)
				p.buf.push(textPiece("]"))
				continue
			}
			p.pos += n
			p.processEmphasis(opener)
			if opener.typ == '[' {
				for d := opener.prev; d != p.delims.bottom; d = d.prev {
					if d.typ == '[' {
						d.inactive = true
					}
				}
			}
			unlink(opener)
			if opener.typ == '[' {
				p.buf.pieces[opener.bufIdx] = piece{
					before: []InlineOp{{Type: OpLinkStart, Dest: dest, Text: title}}}
				p.buf.push(piece{
					after: []InlineOp{{Type: OpLinkEnd, Dest: dest, Text: title}}})
			} else {
				// Use the pieces after "![" to build the image alt text.
				var altBuilder strings.Builder
				for _, piece := range p.buf.pieces[opener.bufIdx+1:] {
					altBuilder.WriteString(piece.main.String())
				}
				p.buf.pieces = p.buf.pieces[:opener.bufIdx]
				alt := altBuilder.String()
				p.buf.push(piece{
					main: InlineOp{Type: OpImage, Dest: dest, Alt: alt, Text: title}})
			}
		case '`':
			// https://spec.commonmark.org/0.31.2/#code-spans
			p.consumeRun('`')
			closer := findBacktickRun(p.text, p.text[begin:p.pos], p.pos)
			if closer == -1 {
				// No matching closer, don't parse as code span.
				parseText()
				continue
			}
			p.buf.push(piece{
				main: InlineOp{Type: OpCodeSpan,
					Text: normalizeCodeSpanContent(p.text[p.pos:closer])}})
			p.pos = closer + (p.pos - begin)
		case '<':
			// https://spec.commonmark.org/0.31.2/#raw-html
			if p.pos == len(p.text) {
				parseText()
				continue
			}
			parseWithRegexp := func(pattern *regexp.Regexp) bool {
				html := pattern.FindString(p.text[begin:])
				if html == "" {
					return false
				}
				p.buf.push(htmlPiece(html))
				p.pos = begin + len(html)
				return true
			}
			parseWithCloser := func(closer string) bool {
				i := strings.Index(p.text[p.pos:], closer)
				if i == -1 {
					return false
				}
				p.pos += i + len(closer)
				p.buf.push(htmlPiece(p.text[begin:p.pos]))
				return true
			}
			switch p.text[p.pos] {
			case '!':
				switch {
				case strings.HasPrefix(p.text[p.pos:], "!--"):
					// Try parsing a comment.
					if parseWithCloser("-->") {
						continue
					}
				case strings.HasPrefix(p.text[p.pos:], "![CDATA["):
					// Try parsing a CDATA section
					if parseWithCloser("]]>") {
						continue
					}
				case p.pos+1 < len(p.text) && isASCIILetter(p.text[p.pos+1]):
					// Try parsing a declaration.
					if parseWithCloser(">") {
						continue
					}
				}
			case '?':
				// Try parsing a processing instruction.
				closer := strings.Index(p.text[p.pos:], "?>")
				if closer != -1 {
					p.buf.push(htmlPiece(p.text[begin : p.pos+closer+2]))
					p.pos += closer + 2
					continue
				}
			case '/':
				// Try parsing a closing tag.
				if parseWithRegexp(closingTagRegexp) {
					continue
				}
			default:
				// Try parsing a open tag.
				if parseWithRegexp(openTagRegexp) {
					continue
				} else {
					// Try parsing an autolink.
					autolink := uriAutolinkRegexp.FindString(p.text[begin:])
					email := false
					if autolink == "" {
						autolink = emailAutolinkRegexp.FindString(p.text[begin:])
						email = true
					}
					if autolink != "" {
						p.pos = begin + len(autolink)
						// Autolinks support character references but not
						// backslashes, so UnescapeHTML gives us the desired
						// behavior.
						text := UnescapeHTML(autolink[1 : len(autolink)-1])
						dest := text
						if email {
							dest = "mailto:" + dest
						}
						p.buf.push(piece{
							main: InlineOp{Type: OpAutolink, Text: text, Dest: dest},
						})
						continue
					}
				}
			}
			parseText()
		case '&':
			// https://spec.commonmark.org/0.31.2/#entity-and-numeric-character-references
			if entity := leadingCharRef(p.text[begin:]); entity != "" {
				p.buf.push(textPiece(UnescapeHTML(entity)))
				p.pos = begin + len(entity)
			} else {
				parseText()
			}
		case '\\':
			// https://spec.commonmark.org/0.31.2/#backslash-escapes
			if p.pos < len(p.text) {
				if p.text[p.pos] == '\n' {
					// https://spec.commonmark.org/0.31.2/#hard-line-break
					//
					// Do *not* consume the newline; "\\\n" is a hard line break
					// plus a (soft) line break.
					p.buf.push(piece{main: InlineOp{Type: OpHardLineBreak}})
					continue
				} else if isASCIIPunct(p.text[p.pos]) {
					// Valid backslash escape: handle this by just discarding
					// the backslash. The parseText call below will consider the
					// next byte to be already included in the text content.
					begin++
					p.pos++
				}
			}
			parseText()
		case '\n':
			// Hard line breaks are already inserted using lookahead in
			// parseText and the case '\\' branch.

			p.buf.push(piece{main: InlineOp{Type: OpNewLine}})
			// Remove spaces at the beginning of the next line per
			// https://spec.commonmark.org/0.31.2/#soft-line-breaks.
			for p.pos < len(p.text) && p.text[p.pos] == ' ' {
				p.pos++
			}
		default:
			parseText()
		}
	}
	p.processEmphasis(p.delims.bottom)
}

func (p *inlineParser) consumeRun(b byte) {
	for p.pos < len(p.text) && p.text[p.pos] == b {
		p.pos++
	}
}

// Processes the (rune, int) result of utf8.Decode* so that an empty result is
// converted to '\n'.
func emptyToNewline(r rune, l int) rune {
	if l == 0 {
		return '\n'
	}
	return r
}

// Returns whether an emphasis punctuation can open or close an emphasis, when
// following prev and preceding next. Start and end of file should be
// represented by '\n'.
//
// The criteria are described in:
// https://spec.commonmark.org/0.31.2/#emphasis-and-strong-emphasis
//
// The algorithm is a bit complicated. Here is another way to describe the
// criteria:
//
//   - Every rune falls into one of three categories: space, punctuation and
//     other. "Other" is the category of word runes in "intraword emphasis".
//
//   - The following tables describe whether a punctuation can open or close
//     emphasis:
//
//     Can open emphasis:
//
//     |            | next space | next punct | next other |
//     | ---------- | ---------- | ---------- | ---------- |
//     | prev space |            |   _ or *   |   _ or *   |
//     | prev punct |            |   _ or *   |   _ or *   |
//     | prev other |            |            |   only *   |
//
//     Can close emphasis:
//
//     |            | next space | next punct | next other |
//     | ---------- | ---------- | ---------- | ---------- |
//     | prev space |            |            |            |
//     | prev punct |   _ or *   |   _ or *   |            |
//     | prev other |   _ or *   |   _ or *   |   only *   |
func canOpenCloseEmphasis(b, prev, next rune) (bool, bool) {
	leftFlanking := !unicode.IsSpace(next) &&
		(!isUnicodePunct(next) || unicode.IsSpace(prev) || isUnicodePunct(prev))
	rightFlanking := !unicode.IsSpace(prev) &&
		(!isUnicodePunct(prev) || unicode.IsSpace(next) || isUnicodePunct(next))
	if b == '*' {
		return leftFlanking, rightFlanking
	}
	return leftFlanking && (!rightFlanking || isUnicodePunct(prev)),
		rightFlanking && (!leftFlanking || isUnicodePunct(next))
}

// Returns the starting index of the next backtick run identical to the given
// run, starting from i. Returns -1 if no such run exists.
func findBacktickRun(s, run string, i int) int {
	for i < len(s) {
		j := strings.Index(s[i:], run)
		if j == -1 {
			return -1
		}
		j += i
		if j+len(run) == len(s) || s[j+len(run)] != '`' {
			return j
		}
		// Too many backticks; skip over the entire run.
		for j += len(run); j < len(s) && s[j] == '`'; j++ {
		}
		i = j
	}
	return -1
}

func normalizeCodeSpanContent(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > 1 && s[0] == ' ' && s[len(s)-1] == ' ' && strings.Trim(s, " ") != "" {
		return s[1 : len(s)-1]
	}
	return s
}

// https://spec.commonmark.org/0.31.2/#process-emphasis
func (p *inlineParser) processEmphasis(bottom *delim) {
	var openersBottom [2][3][2]*delim
	for closer := bottom.next; closer != nil; {
		if !closer.canClose {
			closer = closer.next
			continue
		}
		openerBottom := &openersBottom[b2i(closer.typ == '_')][closer.n%3][b2i(closer.canOpen)]
		if *openerBottom == nil {
			*openerBottom = bottom
		}
		var opener *delim
		for p := closer.prev; p != *openerBottom && p != bottom; p = p.prev {
			if p.canOpen && p.typ == closer.typ &&
				((!p.canClose && !closer.canOpen) ||
					(p.n+closer.n)%3 != 0 || (p.n%3 == 0 && closer.n%3 == 0)) {
				opener = p
				break
			}
		}
		if opener == nil {
			*openerBottom = closer.prev
			if !closer.canOpen {
				closer.prev.next = closer.next
				closer.next.prev = closer.prev
			}
			closer = closer.next
			continue
		}
		openerPiece := &p.buf.pieces[opener.bufIdx]
		closerPiece := &p.buf.pieces[closer.bufIdx]
		strong := len(openerPiece.main.Text) >= 2 && len(closerPiece.main.Text) >= 2
		if strong {
			openerPiece.main.Text = openerPiece.main.Text[2:]
			openerPiece.append(InlineOp{Type: OpStrongEmphasisStart})
			closerPiece.main.Text = closerPiece.main.Text[2:]
			closerPiece.prepend(InlineOp{Type: OpStrongEmphasisEnd})
		} else {
			openerPiece.main.Text = openerPiece.main.Text[1:]
			openerPiece.append(InlineOp{Type: OpEmphasisStart})
			closerPiece.main.Text = closerPiece.main.Text[1:]
			closerPiece.prepend(InlineOp{Type: OpEmphasisEnd})
		}
		opener.next = closer
		closer.prev = opener
		if openerPiece.main.Text == "" {
			opener.prev.next = opener.next
			opener.next.prev = opener.prev
		}
		if closerPiece.main.Text == "" {
			closer.prev.next = closer.next
			closer.next.prev = closer.prev
			closer = closer.next
		}
	}
	bottom.next = p.delims.top
	p.delims.top.prev = bottom
}

func b2i(b bool) int {
	if b {
		return 1
	} else {
		return 0
	}
}

// Stores output of inline rendering.
type buffer struct {
	pieces []piece
}

func (b *buffer) push(p piece) int {
	b.pieces = append(b.pieces, p)
	return len(b.pieces) - 1
}

func (b *buffer) ops() []InlineOp {
	var ops []InlineOp
	for _, p := range b.pieces {
		p.iterate(func(op InlineOp) {
			if op.Type == OpText {
				// Convert any embedded newlines into OpNewLine, and merge
				// adjacent OpText's or OpRawHTML's.
				if op.Text == "" {
					return
				}
				lines := strings.Split(op.Text, "\n")
				if len(ops) > 0 && ops[len(ops)-1].Type == op.Type {
					ops[len(ops)-1].Text += lines[0]
				} else if lines[0] != "" {
					ops = append(ops, InlineOp{Type: op.Type, Text: lines[0]})
				}
				for _, line := range lines[1:] {
					ops = append(ops, InlineOp{Type: OpNewLine})
					if line != "" {
						ops = append(ops, InlineOp{Type: op.Type, Text: line})
					}
				}
			} else {
				ops = append(ops, op)
			}
		})
	}
	return ops
}

// The algorithm described in
// https://spec.commonmark.org/0.31.2/#phase-2-inline-structure involves inserting
// nodes before and after existing nodes in the output. The most natural choice
// is a doubly linked list; but for simplicity, we use a slice for output nodes,
// keep track of nodes that need to be prepended or appended to each node.
//
// TODO: Compare the performance of this data structure with doubly linked
// lists.
type piece struct {
	before []InlineOp
	main   InlineOp
	after  []InlineOp
}

func textPiece(text string) piece {
	return piece{main: InlineOp{Type: OpText, Text: text}}
}

func htmlPiece(html string) piece {
	return piece{main: InlineOp{Type: OpRawHTML, Text: html}}
}

func (p *piece) prepend(op InlineOp) { p.before = append(p.before, op) }
func (p *piece) append(op InlineOp)  { p.after = append(p.after, op) }

func (p *piece) iterate(f func(InlineOp)) {
	for _, op := range p.before {
		f(op)
	}
	f(p.main)
	for i := len(p.after) - 1; i >= 0; i-- {
		f(p.after[i])
	}
}

// A delimiter "stack" (actually a doubly linked list), with sentinels as bottom
// and top, with the bottom being the head of the list.
//
// https://spec.commonmark.org/0.31.2/#delimiter-stack
type delimStack struct {
	bottom, top *delim
}

func makeDelimStack() delimStack {
	bottom := &delim{}
	top := &delim{prev: bottom}
	bottom.next = top
	return delimStack{bottom, top}
}

func (s *delimStack) push(n *delim) {
	n.prev = s.top.prev
	n.next = s.top
	s.top.prev.next = n
	s.top.prev = n
}

// A node in the delimiter "stack".
type delim struct {
	typ    byte
	bufIdx int
	prev   *delim
	next   *delim
	// Only used when typ is '['
	inactive bool
	// Only used when typ is '_' or '*'.
	n        int
	canOpen  bool
	canClose bool
}

func unlink(n *delim) {
	n.next.prev = n.prev
	n.prev.next = n.next
}

type linkTailParser struct {
	text string
	pos  int
}

// Parses the link "tail", the part after the ] that closes the link text.
func parseLinkTail(text string) (n int, dest, title string) {
	p := linkTailParser{text, 0}
	return p.parse()
}

// https://spec.commonmark.org/0.31.2/#links
func (p *linkTailParser) parse() (n int, dest, title string) {
	if len(p.text) < 2 || p.text[0] != '(' {
		return -1, "", ""
	}

	p.pos = 1
	p.skipWhitespaces()
	if p.pos == len(p.text) {
		return -1, "", ""
	}
	// Parse an optional link destination.
	var destBuilder strings.Builder
	if p.text[p.pos] == '<' {
		p.pos++
		closed := false
	angleDest:
		for p.pos < len(p.text) {
			switch p.text[p.pos] {
			case '>':
				p.pos++
				closed = true
				break angleDest
			case '\n', '<':
				return -1, "", ""
			case '\\':
				destBuilder.WriteByte(p.parseBackslash())
			case '&':
				destBuilder.WriteString(p.parseCharRef())
			default:
				destBuilder.WriteByte(p.text[p.pos])
				p.pos++
			}
		}
		if !closed {
			return -1, "", ""
		}
	} else {
		parenBalance := 0
	bareDest:
		for p.pos < len(p.text) {
			if isASCIIControl(p.text[p.pos]) || p.text[p.pos] == ' ' {
				break
			}
			switch p.text[p.pos] {
			case '(':
				parenBalance++
				destBuilder.WriteByte('(')
				p.pos++
			case ')':
				if parenBalance == 0 {
					break bareDest
				}
				parenBalance--
				destBuilder.WriteByte(')')
				p.pos++
			case '\\':
				destBuilder.WriteByte(p.parseBackslash())
			case '&':
				destBuilder.WriteString(p.parseCharRef())
			default:
				destBuilder.WriteByte(p.text[p.pos])
				p.pos++
			}
		}
		if parenBalance != 0 {
			return -1, "", ""
		}
	}
	p.skipWhitespaces()

	var titleBuilder strings.Builder
	if p.pos < len(p.text) && strings.ContainsRune("'\"(", rune(p.text[p.pos])) {
		opener := p.text[p.pos]
		closer := p.text[p.pos]
		if closer == '(' {
			closer = ')'
		}
		p.pos++
	title:
		for p.pos < len(p.text) {
			switch p.text[p.pos] {
			case closer:
				p.pos++
				break title
			case opener:
				// Titles started with "(" does not allow unescaped "(":
				// https://spec.commonmark.org/0.31.2/#link-title
				return -1, "", ""
			case '\\':
				titleBuilder.WriteByte(p.parseBackslash())
			case '&':
				titleBuilder.WriteString(p.parseCharRef())
			default:
				titleBuilder.WriteByte(p.text[p.pos])
				p.pos++
			}
		}
	}

	p.skipWhitespaces()

	if p.pos == len(p.text) || p.text[p.pos] != ')' {
		return -1, "", ""
	}
	return p.pos + 1, destBuilder.String(), titleBuilder.String()
}

func (p *linkTailParser) skipWhitespaces() {
	for p.pos < len(p.text) && isWhitespace(p.text[p.pos]) {
		p.pos++
	}
}

func isWhitespace(b byte) bool { return b == ' ' || b == '\t' || b == '\n' }

func (p *linkTailParser) parseBackslash() byte {
	if p.pos+1 < len(p.text) && isASCIIPunct(p.text[p.pos+1]) {
		b := p.text[p.pos+1]
		p.pos += 2
		return b
	}
	p.pos++
	return '\\'
}

func (p *linkTailParser) parseCharRef() string {
	if entity := leadingCharRef(p.text[p.pos:]); entity != "" {
		p.pos += len(entity)
		return UnescapeHTML(entity)
	}
	p.pos++
	return p.text[p.pos-1 : p.pos]
}

func isASCIILetter(b byte) bool { return ('a' <= b && b <= 'z') || ('A' <= b && b <= 'Z') }

func isASCIIControl(b byte) bool { return b < 0x20 }

const asciiPuncts = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"

func isASCIIPunct(b byte) bool { return strings.IndexByte(asciiPuncts, b) >= 0 }

// The CommonMark spec's definition of Unicode punctuation includes both P and S
// categories: https://spec.commonmark.org/0.31.2/#unicode-punctuation-character
func isUnicodePunct(r rune) bool {
	return unicode.IsPunct(r) || unicode.IsSymbol(r)
}

const metas = "![]*_`\\&<\n"

func isMeta(b byte) bool { return strings.IndexByte(metas, b) >= 0 }
