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
	// OpText, OpCodeSpan, OpRawHTML: Text content
	// OpLinkStart, OpLinkEnd, OpImage: title text
	Text string
	// OpLinkStart, OpLinkEnd, OpImage
	Dest string
	// ForOpImage
	Alt string
}

// InlineOpType enumerates possible types of an InlineOp.
type InlineOpType uint

const (
	// Text elements.
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
	OpHardLineBreak
)

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

var (
	// https://spec.commonmark.org/0.30/#entity-and-numeric-character-references
	entityRegexp = regexp.MustCompile(`^&(?:[a-zA-Z0-9]+|#[0-9]{1,7}|#[xX][0-9a-fA-F]{1,6});`)

	// https://spec.commonmark.org/0.30/#uri-autolink
	uriAutolinkRegexp = regexp.MustCompile(`^<` +
		`[a-zA-Z][a-zA-Z0-9+.-]{1,31}` + // scheme
		`:[^\x00-\x19 <>]*` +
		`>`)
	// https://spec.commonmark.org/0.30/#email-autolink
	emailAutolinkRegexp = regexp.MustCompile(
		`^<[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*>`)

	openTagRegexp    = regexp.MustCompile(`^` + openTag)
	closingTagRegexp = regexp.MustCompile(`^` + closingTag)
)

const (
	// https://spec.commonmark.org/0.30/#open-tag
	openTag = `<` +
		`[a-zA-Z][a-zA-Z0-9-]*` + // tag name
		(`(?:` +
			`[ \t\n]+` + // whitespace
			`[a-zA-Z_:][a-zA-Z0-9_\.:-]*` + // attribute name
			`(?:[ \t\n]*=[ \t\n]*(?:[^ \t\n"'=<>` + "`" + `]+|'[^']*'|"[^"]*"))?` + // attribute value specification
			`)*`) + // zero or more attributes
		`[ \t\n]*` + // whitespace
		`/?>`
	// https://spec.commonmark.org/0.30/#closing-tag
	closingTag = `</[a-zA-Z][a-zA-Z0-9-]*[ \t\n]*>`
)

func (p *inlineParser) render() {
	for p.pos < len(p.text) {
		b := p.text[p.pos]
		begin := p.pos
		p.pos++

		parseText := func() {
			for p.pos < len(p.text) && !isMeta(p.text[p.pos]) {
				p.pos++
			}
			p.buf.push(textPiece(p.text[begin:p.pos]))
		}

		switch b {
		// The 3 branches below implement the first part of
		// https://spec.commonmark.org/0.30/#an-algorithm-for-parsing-nested-emphasis-and-links.
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
			next, lNext := utf8.DecodeRuneInString(p.text[p.pos:])
			prev, lPrev := utf8.DecodeLastRuneInString(p.text[:begin])
			// Left-flanking, right-flanking, can-open and can-close are defined
			// in https://spec.commonmark.org/0.30/#emphasis-and-strong-emphasis
			leftFlanking := lNext > 0 && !unicode.IsSpace(next) &&
				(!isUnicodePunct(next) ||
					(lPrev == 0 || unicode.IsSpace(prev) || isUnicodePunct(prev)))
			rightFlanking := lPrev > 0 && !unicode.IsSpace(prev) &&
				(!isUnicodePunct(prev) ||
					(lNext == 0 || unicode.IsSpace(next) || isUnicodePunct(next)))
			canOpen := leftFlanking
			canClose := rightFlanking
			if b == '_' {
				canOpen = leftFlanking && (!rightFlanking || (lPrev > 0 && isUnicodePunct(prev)))
				canClose = rightFlanking && (!leftFlanking || (lNext > 0 && isUnicodePunct(next)))
			}
			bufIdx := p.buf.push(textPiece(p.text[begin:p.pos]))
			p.delims.push(
				&delim{typ: b, bufIdx: bufIdx,
					n: p.pos - begin, canOpen: canOpen, canClose: canClose})
		case ']':
			// https://spec.commonmark.org/0.30/#look-for-link-or-image.
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
				var altBuilder strings.Builder
				for _, piece := range p.buf.pieces[opener.bufIdx+1:] {
					altBuilder.WriteString(plainText(piece))
				}
				p.buf.pieces = p.buf.pieces[:opener.bufIdx]
				alt := altBuilder.String()
				p.buf.push(piece{
					main: InlineOp{Type: OpImage, Dest: dest, Alt: alt, Text: title}})
			}
		case '`':
			// https://spec.commonmark.org/0.30/#code-spans
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
			// https://spec.commonmark.org/0.30/#raw-html
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
						// Autolinks support entities but not backslashes, so
						// UnescapeEntities gives us the desired behavior.
						text := UnescapeEntities(autolink[1 : len(autolink)-1])
						dest := text
						if email {
							dest = "mailto:" + dest
						}
						p.buf.push(piece{
							before: []InlineOp{{Type: OpLinkStart, Dest: dest}},
							main:   InlineOp{Type: OpText, Text: text},
							after:  []InlineOp{{Type: OpLinkEnd, Dest: dest}},
						})
						continue
					}
				}
			}
			parseText()
		case '&':
			// https://spec.commonmark.org/0.30/#entity-and-numeric-character-references
			entity := entityRegexp.FindString(p.text[begin:])
			if entity != "" {
				p.buf.push(textPiece(UnescapeEntities(entity)))
				p.pos = begin + len(entity)
			} else {
				parseText()
			}
		case '\\':
			// https://spec.commonmark.org/0.30/#backslash-escapes
			if p.pos < len(p.text) && isASCIIPunct(p.text[p.pos]) {
				begin++
				p.pos++
			}
			parseText()
		case '\n':
			// https://spec.commonmark.org/0.30/#hard-line-breaks
			// https://spec.commonmark.org/0.30/#soft-line-breaks
			if len(p.buf.pieces) > 0 {
				last := &p.buf.pieces[len(p.buf.pieces)-1]
				if last.before == nil && last.after == nil && last.main.Type == OpText {
					text := &last.main.Text
					if p.pos == len(p.text) {
						*text = strings.TrimRight(*text, " ")
					} else {
						hardLineBreak := false
						if strings.HasSuffix(*text, "\\") {
							hardLineBreak = true
							*text = (*text)[:len(*text)-1]
						} else {
							hardLineBreak = strings.HasSuffix(*text, "  ")
							last.main.Text = strings.TrimRight(*text, " ")
						}
						if hardLineBreak {
							p.buf.push(piece{main: InlineOp{Type: OpHardLineBreak}})
						}
					}
				}
			}
			p.buf.push(piece{main: InlineOp{Type: OpNewLine}})
			// Remove spaces at the beginning of the next line per
			// https://spec.commonmark.org/0.30/#soft-line-breaks.
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

// https://spec.commonmark.org/0.30/#process-emphasis
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
		for p := closer.prev; p != *openerBottom; p = p.prev {
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
			if op.Type == OpText || op.Type == OpRawHTML {
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
// https://spec.commonmark.org/0.30/#phase-2-inline-structure involves inserting
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

func plainText(p piece) string {
	switch p.main.Type {
	case OpText, OpCodeSpan, OpRawHTML:
		return p.main.Text
	case OpNewLine:
		return "\n"
	case OpImage:
		return p.main.Alt
	}
	return ""
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
// https://spec.commonmark.org/0.30/#delimiter-stack
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

// https://spec.commonmark.org/0.30/#links
func (p *linkTailParser) parse() (n int, dest, title string) {
	if len(p.text) < 2 || p.text[0] != '(' {
		return -1, "", ""
	}

	p.pos = 1
	p.skipWhitespaces()
	// Parse an optional link destination.
	var destBuilder strings.Builder
	if p.text[1] == '<' {
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
				destBuilder.WriteString(p.parseEntity())
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
				// https://spec.commonmark.org/0.30/#link-title
				return -1, "", ""
			case '\\':
				titleBuilder.WriteByte(p.parseBackslash())
			case '&':
				titleBuilder.WriteString(p.parseEntity())
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

func (p *linkTailParser) parseEntity() string {
	if entity := entityRegexp.FindString(p.text[p.pos:]); entity != "" {
		p.pos += len(entity)
		return UnescapeEntities(entity)
	}
	p.pos++
	return p.text[p.pos-1 : p.pos]
}

func isASCIILetter(b byte) bool { return ('a' <= b && b <= 'z') || ('A' <= b && b <= 'Z') }

func isASCIIControl(b byte) bool { return b < 0x20 }

const asciiPuncts = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"

func isASCIIPunct(b byte) bool { return strings.IndexByte(asciiPuncts, b) >= 0 }

// The CommonMark spec has its own definition of Unicode punctuation:
// https://spec.commonmark.org/0.30/#unicode-punctuation-character
//
// This definition includes all the ASCII punctuations above, some of which
// ("$+<=>^`|~" to be exact) are not considered to be punctuations by
// unicode.IsPunct.
func isUnicodePunct(r rune) bool {
	return unicode.IsPunct(r) || r <= 0x7f && isASCIIPunct(byte(r))
}

const metas = "![]*_`\\&<\n"

func isMeta(b byte) bool { return strings.IndexByte(metas, b) >= 0 }
