// Package md implements a Markdown renderer.
//
// This package is incomplete and not used anywhere in Elvish right now. When it
// is complete, it will be used for rendering the elvdoc of builtin modules
// inside terminals.
//
// All inline features of CommonMark have been implemented, with almost all
// relevant spec tests passing. Block features not used in any elvdoc will
// probably be omitted.
//
// The implementation targets the HEAD of the CommonMark spec in
// https://github.com/commonmark/commonmark-spec, which may differ slightly
// from the latest released version of the spec.
package md

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// OutputSyntax specifies the output syntax.
type OutputSyntax struct {
	ThematicBreak  func(original string) string
	Heading        func(level int) TagPair
	CodeBlock      func(info string) TagPair
	Paragraph      TagPair
	Blockquote     TagPair
	BulletList     TagPair
	BulletItem     TagPair
	OrderedList    func(start int) TagPair
	OrderedItem    TagPair
	CodeSpan       TagPair
	Emphasis       TagPair
	StrongEmphasis TagPair
	Link           func(dest, title string) TagPair
	Image          func(dest, alt, title string) string

	Escape      func(string) string
	ConvertHTML func(string) string
}

// TagPair specifies a pair of "tags" to enclose a construct in the output.
type TagPair struct {
	Start, End string
}

// Render parses markdown and renders it according to the output syntax.
func Render(text string, syntax OutputSyntax) string {
	p := blockParser{
		lines:  lineSplitter{text, 0},
		syntax: syntax,
	}
	p.render()
	return p.sb.String()
}

type blockParser struct {
	lines      lineSplitter
	syntax     OutputSyntax
	containers []container
	paragraph  []string
	sb         strings.Builder
}

var (
	thematicBreakRegexp = regexp.MustCompile(
		`^[ \t]*((?:-[ \t]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})$`)

	// Capture group 1: heading opener
	atxHeadingRegexp       = regexp.MustCompile(`^ *(#{1,6})(?:[ \t]|$)`)
	atxHeadingCloserRegexp = regexp.MustCompile(`[ \t]#+[ \t]*$`)

	// Capture groups:
	// 1. Indent
	// 2. Fence punctuations (backquote fence)
	// 3. Untrimmed info string (backquote fence)
	// 4. Fence punctuations (tilde fence)
	// 5. Untrimmed info string (tilde fence)
	codeFenceRegexp = regexp.MustCompile("(^ {0,3})(?:(`{3,})([^`]*)|(~{3,})(.*))$")
	// Capture group 1: fence punctuations
	codeFenceCloserRegexp = regexp.MustCompile("(?:^ {0,3})(`{3,}|~{3,})[ \t]*$")

	html1Regexp       = regexp.MustCompile(`^ {0,3}<(?i:pre|script|style|textarea)`)
	html1CloserRegexp = regexp.MustCompile(`</(?i:pre|script|style|textarea)`)
	html2Regexp       = regexp.MustCompile(`^ {0,3}<!--`)
	html2CloserRegexp = regexp.MustCompile(`-->`)
	html3Regexp       = regexp.MustCompile(`^ {0,3}<\?`)
	html3CloserRegexp = regexp.MustCompile(`\?>`)
	html4Regexp       = regexp.MustCompile(`^ {0,3}<![a-zA-Z]`)
	html4CloserRegexp = regexp.MustCompile(`>`)
	html5Regexp       = regexp.MustCompile(`^ {0,3}<!\[CDATA\[`)
	html5CloserRegexp = regexp.MustCompile(`\]\]>`)

	html6Regexp = regexp.MustCompile(`^ {0,3}</?(?i:address|article|aside|base|basefont|blockquote|body|caption|center|col|colgroup|dd|details|dialog|dir|div|dl|dt|fieldset|figcaption|figure|footer|form|frame|frameset|h1|h2|h3|h4|h5|h6|head|header|hr|html|iframe|legend|li|link|main|menu|menuitem|nav|noframes|ol|optgroup|option|p|param|section|source|summary|table|tbody|td|tfoot|th|thead|title|tr|track|ul)(?:[ \t>]|$|/>)`)
	html7Regexp = regexp.MustCompile(
		fmt.Sprintf(`^ {0,3}(?:%s|%s)[ \t]*$`, openTag, closingTag))
)

const (
	openTag = `<` +
		`[a-zA-Z][a-zA-Z0-9-]*` + // tag name
		(`(?:` +
			`[ \t\n]+` + // whitespace
			`[a-zA-Z_:][a-zA-Z0-9_\.:-]*` + // attribute name
			`(?:[ \t\n]*=[ \t\n]*(?:[^ \t\n"'=<>` + "`" + `]+|'[^']*'|"[^"]*"))?` + // attribute value specification
			`)*`) + // zero or more attributes
		`[ \t\n]*` + // whitespace
		`/?>`
	closingTag = `</[a-zA-Z][a-zA-Z0-9-]*[ \t\n]*>`
)

func (p *blockParser) render() {
	for p.lines.more() {
		line := p.lines.next()
		line, matchedContainers := matchContinuationMarkers(line, p.containers)
		newParagraph := matchedContainers != len(p.containers) || len(p.paragraph) == 0
		line, newContainers := parseStartingMarkers(line, newParagraph)
		if len(newContainers) > 0 {
			p.popParagraph(matchedContainers)
			for _, c := range newContainers {
				p.appendContainer(c)
			}
			matchedContainers = len(p.containers)
		}

		if isBlankLine(line) {
			for i := matchedContainers; i < len(p.containers); i++ {
				if p.containers[i].typ == blockquote {
					p.popParagraph(i)
					p.popList()
					continue
				}
			}
			if len(newContainers) == 0 && len(p.paragraph) == 0 &&
				(p.lastContainerIs(bulletItem) || p.lastContainerIs(orderedItem)) {
				if p.containers[len(p.containers)-1].blankFirst {
					p.popLastContainer()
				}
			}
			p.popParagraph(len(p.containers))
		} else if thematicBreakRegexp.MatchString(line) {
			p.popParagraph(matchedContainers)
			p.popList()
			p.sb.WriteString(p.syntax.ThematicBreak(line))
			p.sb.WriteByte('\n')
		} else if m := atxHeadingRegexp.FindStringSubmatchIndex(line); m != nil {
			p.popParagraph(matchedContainers)
			p.popList()
			openerStart, openerEnd := m[2], m[3]
			opener := line[openerStart:openerEnd]
			line = strings.TrimRight(line[openerEnd:], " \t")
			if closer := atxHeadingCloserRegexp.FindString(line); closer != "" {
				line = line[:len(line)-len(closer)]
			}
			p.renderLeaf(p.syntax.Heading(len(opener)), strings.Trim(line, " \t"))
		} else if m := codeFenceRegexp.FindStringSubmatch(line); m != nil {
			p.popParagraph(matchedContainers)
			p.popList()
			indent, opener, info := len(m[1]), m[2], m[3]
			if opener == "" {
				opener, info = m[4], m[5]
			}
			p.parseFencedCodeBlock(indent, opener, info)
		} else if html1Regexp.MatchString(line) {
			p.popParagraph(matchedContainers)
			p.popList()
			p.parseHTMLBlock(line, html1CloserRegexp.MatchString)
		} else if html2Regexp.MatchString(line) {
			p.popParagraph(matchedContainers)
			p.popList()
			p.parseHTMLBlock(line, html2CloserRegexp.MatchString)
		} else if html3Regexp.MatchString(line) {
			p.popParagraph(matchedContainers)
			p.popList()
			p.parseHTMLBlock(line, html3CloserRegexp.MatchString)
		} else if html4Regexp.MatchString(line) {
			p.popParagraph(matchedContainers)
			p.popList()
			p.parseHTMLBlock(line, html4CloserRegexp.MatchString)
		} else if html5Regexp.MatchString(line) {
			p.popParagraph(matchedContainers)
			p.popList()
			p.parseHTMLBlock(line, html5CloserRegexp.MatchString)
		} else if html6Regexp.MatchString(line) || (newParagraph && html7Regexp.MatchString(line)) {
			p.popParagraph(matchedContainers)
			p.popList()
			p.parseBlankLineTerminatedHTMLBlock(line)
		} else {
			if len(p.paragraph) == 0 {
				p.popParagraph(matchedContainers)
				p.popList()
			}
			p.addParagraphLine(line)
		}
	}
	p.popParagraph(0)
}

// Matches the continuation markers of existing container nodes. Returns the
// line after removing all matched continuation markers and the number of
// containers matched.
func matchContinuationMarkers(line string, containers []container) (string, int) {
	for i, container := range containers {
		markerLen, matched := container.findContinuationMarker(line)
		if !matched {
			return line, i
		}
		line = line[markerLen:]
	}
	return line, len(containers)
}

var (
	containerStartingMarkerRegexp = regexp.MustCompile(
		// Capture groups:
		// 1. blockquote marker
		// 2. bullet item punctuation
		// 3. ordered item start index
		// 4. ordered item punctuation
		`^ {0,3}(?:(> ?)|([-+*]) {1,4}|([0-9]{1,9})([.)]) {1,4})`)
	itemStartingMarkerBlankLineRegexp = regexp.MustCompile(
		// Capture groups are the same, with group 1 always empty.
		`^ {0,3}(?:()([-+*])|([0-9]{1,9})([.)]))[ \t]*$`)
)

// Parses starting markers of container blocks. Returns the line after removing
// all starting markers and new containers to create.
func parseStartingMarkers(line string, newParagraph bool) (string, []container) {
	var containers []container
	// Don't parse thematic breaks like "- - - " as three bullets.
	for !thematicBreakRegexp.MatchString(line) {
		m := containerStartingMarkerRegexp.FindStringSubmatch(line)
		blankFirst := false
		if m == nil && newParagraph {
			m = itemStartingMarkerBlankLineRegexp.FindStringSubmatch(line)
			blankFirst = true
		}
		if m == nil {
			break
		}
		marker, bq, bulletPunct, orderedStart, orderedPunct := m[0], m[1], m[2], m[3], m[4]
		var c container
		if bq != "" {
			c.typ = blockquote
		} else {
			indent := len(marker)
			if strings.Trim(line[len(marker):], " \t") == "" {
				indent = len(strings.TrimRight(marker, " \t")) + 1
			}
			c.indent = strings.Repeat(" ", indent)
			c.blankFirst = blankFirst
			if bulletPunct != "" {
				c.typ = bulletItem
				c.punct = bulletPunct[0]
			} else {
				c.typ = orderedItem
				c.punct = orderedPunct[0]
				c.start, _ = strconv.Atoi(orderedStart)
				if c.start != 1 && !newParagraph {
					break
				}
			}
		}
		line = line[len(marker):]
		containers = append(containers, c)
	}
	return line, containers
}

func isBlankLine(line string) bool {
	return strings.Trim(line, " \t") == ""
}

func (p *blockParser) parseFencedCodeBlock(indent int, opener, info string) {
	tags := p.syntax.CodeBlock(strings.Trim(info, " \t"))
	p.sb.WriteString(tags.Start)
	for p.lines.more() {
		line := p.lines.next()
		line, matchedContainers := matchContinuationMarkers(line, p.containers)
		if isBlankLine(line) {
			for i := matchedContainers; i < len(p.containers); i++ {
				if p.containers[i].typ == blockquote {
					p.sb.WriteString(tags.End)
					p.sb.WriteByte('\n')
					p.popParagraph(i)
					return
				}
			}
		} else if matchedContainers < len(p.containers) {
			p.sb.WriteString(tags.End)
			p.sb.WriteByte('\n')
			p.lines.backup()
			return
		}
		if m := codeFenceCloserRegexp.FindStringSubmatch(line); m != nil {
			closer := m[1]
			if closer[0] == opener[0] && len(closer) >= len(opener) {
				break
			}
		}
		for i := indent; i > 0 && line != "" && line[0] == ' '; i-- {
			line = line[1:]
		}
		p.sb.WriteString(p.syntax.Escape(line))
		p.sb.WriteByte('\n')
	}
	p.sb.WriteString(tags.End)
	p.sb.WriteByte('\n')
}

func (p *blockParser) parseHTMLBlock(line string, closer func(string) bool) {
	p.sb.WriteString(line)
	p.sb.WriteByte('\n')
	if closer(line) {
		return
	}
	for p.lines.more() {
		line := p.lines.next()
		line, matchedContainers := matchContinuationMarkers(line, p.containers)
		if isBlankLine(line) {
			for i := matchedContainers; i < len(p.containers); i++ {
				if p.containers[i].typ == blockquote {
					p.popParagraph(i)
					return
				}
			}
		} else if matchedContainers < len(p.containers) {
			p.lines.backup()
			return
		}
		p.sb.WriteString(line)
		p.sb.WriteByte('\n')
		if closer(line) {
			return
		}
	}
}

func (p *blockParser) parseBlankLineTerminatedHTMLBlock(line string) {
	p.sb.WriteString(line)
	p.sb.WriteByte('\n')
	for p.lines.more() {
		line := p.lines.next()
		line, matchedContainers := matchContinuationMarkers(line, p.containers)
		if isBlankLine(line) {
			for i := matchedContainers; i < len(p.containers); i++ {
				if p.containers[i].typ == blockquote {
					p.popParagraph(i)
					return
				}
			}
			return
		} else if matchedContainers < len(p.containers) {
			p.lines.backup()
			return
		}
		p.sb.WriteString(line)
		p.sb.WriteByte('\n')
	}
}

func (p *blockParser) appendContainer(c container) {
	if len(p.containers) > 0 {
		leaf := p.containers[len(p.containers)-1]
		if (leaf.typ == bulletList || leaf.typ == orderedList) && leaf.punct != c.punct {
			p.popLastContainer()
		}
	}

	if c.typ == bulletItem && !p.lastContainerIs(bulletList) {
		p.containers = append(p.containers, container{typ: bulletList, punct: c.punct})
		p.sb.WriteString(p.syntax.BulletList.Start)
		p.sb.WriteByte('\n')
	}
	if c.typ == orderedItem && !p.lastContainerIs(orderedList) {
		p.containers = append(p.containers, container{typ: orderedList, punct: c.punct})
		p.sb.WriteString(p.syntax.OrderedList(c.start).Start)
		p.sb.WriteByte('\n')
	}
	p.containers = append(p.containers, c)
	p.sb.WriteString(c.tagPair(&p.syntax).Start)
	p.sb.WriteByte('\n')
}

func (p *blockParser) lastContainerIs(t containerType) bool {
	return len(p.containers) > 0 && p.containers[len(p.containers)-1].typ == t
}

func (p *blockParser) popLastContainer() {
	p.sb.WriteString(p.containers[len(p.containers)-1].tagPair(&p.syntax).End)
	p.sb.WriteByte('\n')
	p.containers = p.containers[:len(p.containers)-1]
}

func (p *blockParser) addParagraphLine(line string) {
	p.paragraph = append(p.paragraph, line)
}

func (p *blockParser) popParagraph(keepContainers int) {
	if len(p.paragraph) > 0 {
		text := strings.Trim(strings.Join(p.paragraph, "\n"), " \t")
		p.renderLeaf(p.syntax.Paragraph, text)
		p.paragraph = p.paragraph[:0]
	}
	for len(p.containers) > keepContainers {
		p.popLastContainer()
	}
}

func (p *blockParser) popList() {
	if p.lastContainerIs(bulletList) || p.lastContainerIs(orderedList) {
		p.popLastContainer()
	}
}

func (p *blockParser) renderLeaf(tags TagPair, content string) {
	p.sb.WriteString(tags.Start)
	p.sb.WriteString(renderInline(content, p.syntax))
	p.sb.WriteString(tags.End)
	p.sb.WriteByte('\n')
}

type lineSplitter struct {
	text string
	pos  int
}

func (s *lineSplitter) more() bool {
	return s.pos < len(s.text)
}

func (s *lineSplitter) next() string {
	begin := s.pos
	delta := strings.IndexByte(s.text[begin:], '\n')
	if delta == -1 {
		s.pos = len(s.text)
		return s.text[begin:]
	}
	s.pos += delta + 1
	return s.text[begin : s.pos-1]
}

func (s *lineSplitter) backup() {
	if s.pos == 0 {
		return
	}
	s.pos = 1 + strings.LastIndexByte(s.text[:s.pos-1], '\n')
}

type container struct {
	typ        containerType
	punct      byte
	start      int
	indent     string
	blankFirst bool
}

type containerType uint8

const (
	blockquote containerType = iota
	bulletList
	bulletItem
	orderedList
	orderedItem
)

var blockquoteMarkerRegexp = regexp.MustCompile(`^ {0,3}> ?`)

func (c container) findContinuationMarker(line string) (int, bool) {
	switch c.typ {
	case blockquote:
		marker := blockquoteMarkerRegexp.FindString(line)
		return len(marker), marker != ""
	case bulletList, orderedList:
		return 0, true
	case bulletItem, orderedItem:
		if strings.HasPrefix(line, c.indent) {
			return len(c.indent), true
		}
		return 0, false
	}
	panic("unreachable")
}

func (c container) tagPair(syntax *OutputSyntax) TagPair {
	switch c.typ {
	case blockquote:
		return syntax.Blockquote
	case bulletList:
		return syntax.BulletList
	case bulletItem:
		return syntax.BulletItem
	case orderedList:
		return syntax.OrderedList(1)
	case orderedItem:
		return syntax.OrderedItem
	}
	panic("unreachable")
}

type buffer struct {
	pieces []piece
}

type piece struct {
	text          string
	altText       string
	prependMarkup []string
	appendMarkup  []string
}

func (p *piece) build(sb *strings.Builder) {
	for _, s := range p.prependMarkup {
		sb.WriteString(s)
	}
	sb.WriteString(p.text)
	for i := len(p.appendMarkup) - 1; i >= 0; i-- {
		sb.WriteString(p.appendMarkup[i])
	}
}

func (b *buffer) push(p piece) int {
	b.pieces = append(b.pieces, p)
	return len(b.pieces) - 1
}

func (b *buffer) String() string {
	var sb strings.Builder
	for _, p := range b.pieces {
		p.build(&sb)
	}
	return sb.String()
}

// A node in the delimiter "stack" (which is actually doubly linked list).
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

// A delimiter "stack" (actually a doubly linked list), with sentinels as bottom
// and top, with the bottom being the head of the list.
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

func renderInline(text string, syntax OutputSyntax) string {
	p := inlineParser{text, syntax, 0, makeDelimStack(), buffer{}}
	p.render()
	return p.buf.String()
}

type inlineParser struct {
	text   string
	syntax OutputSyntax
	pos    int
	delims delimStack
	buf    buffer
}

var isASCIIPunct = map[byte]bool{
	'!': true, '"': true, '#': true, '$': true, '%': true, '&': true,
	'\'': true, '(': true, ')': true, '*': true, '+': true, ',': true,
	'-': true, '.': true, '/': true, ':': true, ';': true, '<': true,
	'=': true, '>': true, '?': true, '@': true, '[': true, '\\': true,
	']': true, '^': true, '_': true, '`': true, '{': true, '|': true,
	'}': true, '~': true,
}

var (
	entityRegexp     = regexp.MustCompile(`^&(?:[a-zA-Z0-9]+|#[0-9]{1,7}|#[xX][0-9a-fA-F]{1,6});`)
	openTagRegexp    = regexp.MustCompile(`^` + openTag)
	closingTagRegexp = regexp.MustCompile(`^` + closingTag)
	autolinkRegexp   = regexp.MustCompile(`^<` +
		`[a-zA-Z][a-zA-Z0-9+.-]{1,31}` + // scheme
		`:[^\x00-\x19 <>]*` +
		`>`)
	emailAutolinkRegexp = regexp.MustCompile(fmt.Sprintf(`^<[a-zA-Z0-9.!#$%%&'*+/=?^_%s{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*>`, "`"))
)

func (p *inlineParser) render() {
	if p.syntax.Escape == nil {
		p.syntax.Escape = func(s string) string { return s }
	}

	for p.pos < len(p.text) {
		b := p.text[p.pos]
		begin := p.pos
		p.pos++

		parseText := func() {
			for p.pos < len(p.text) && !isMeta(p.text[p.pos]) {
				p.pos++
			}
			p.buf.push(piece{text: p.syntax.Escape(p.text[begin:p.pos])})
		}

		switch b {
		case '[':
			bufIdx := p.buf.push(piece{text: "["})
			p.delims.push(&delim{typ: '[', bufIdx: bufIdx})
		case '!':
			if p.pos < len(p.text) && p.text[p.pos] == '[' {
				p.pos++
				bufIdx := p.buf.push(piece{text: "!["})
				p.delims.push(&delim{typ: '!', bufIdx: bufIdx})
			} else {
				parseText()
			}
		case ']':
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
				p.buf.push(piece{text: "]"})
				continue
			}
			n, dest, title := parseLinkTail(p.text[p.pos:])
			if n == -1 {
				unlink(opener)
				p.buf.push(piece{text: "]"})
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
				tags := p.syntax.Link(dest, title)
				p.buf.pieces[opener.bufIdx] = piece{appendMarkup: []string{tags.Start}}
				p.buf.push(piece{appendMarkup: []string{tags.End}})
			} else {
				var altBuilder strings.Builder
				for _, piece := range p.buf.pieces[opener.bufIdx+1:] {
					altBuilder.WriteString(piece.text)
					altBuilder.WriteString(piece.altText)
				}
				p.buf.pieces = p.buf.pieces[:opener.bufIdx]
				alt := altBuilder.String()
				p.buf.push(piece{
					altText:      alt,
					appendMarkup: []string{p.syntax.Image(dest, alt, title)}})
			}
		case '*', '_':
			// Consume the entire run of * or _.
			for p.pos < len(p.text) && p.text[p.pos] == b {
				p.pos++
			}
			next, lNext := utf8.DecodeRuneInString(p.text[p.pos:])
			prev, lPrev := utf8.DecodeLastRuneInString(p.text[:begin])
			leftFlanking := lNext > 0 && !unicode.IsSpace(next) &&
				(!unicode.IsPunct(next) ||
					(lPrev == 0 || unicode.IsSpace(prev) || unicode.IsPunct(prev)))
			rightFlanking := lPrev > 0 && !unicode.IsSpace(prev) &&
				(!unicode.IsPunct(prev) ||
					(lNext == 0 || unicode.IsSpace(next) || unicode.IsPunct(next)))
			canOpen := leftFlanking
			canClose := rightFlanking
			if b == '_' {
				canOpen = leftFlanking && (!rightFlanking || (lPrev > 0 && unicode.IsPunct(prev)))
				canClose = rightFlanking && (!leftFlanking || (lNext > 0 && unicode.IsPunct(next)))
			}
			bufIdx := p.buf.push(piece{text: p.text[begin:p.pos]})
			p.delims.push(
				&delim{typ: b, bufIdx: bufIdx,
					n: p.pos - begin, canOpen: canOpen, canClose: canClose})
		case '`':
			// Consume the entire run of `.
			for p.pos < len(p.text) && p.text[p.pos] == '`' {
				p.pos++
			}
			closer := findBacktickRun(p.text, p.text[begin:p.pos], p.pos)
			if closer == -1 {
				// No matching closer, don't parse as code span.
				parseText()
				continue
			}
			p.buf.push(piece{
				prependMarkup: []string{p.syntax.CodeSpan.Start},
				text:          p.syntax.Escape(normalizeCodeSpanContent(p.text[p.pos:closer])),
				appendMarkup:  []string{p.syntax.CodeSpan.End}})
			p.pos = closer + (p.pos - begin)
		case '<':
			if p.pos == len(p.text) {
				parseText()
				continue
			}
			parseWithRegexp := func(pattern *regexp.Regexp) bool {
				html := pattern.FindString(p.text[begin:])
				if html == "" {
					return false
				}
				p.buf.push(piece{prependMarkup: []string{html}})
				p.pos = begin + len(html)
				return true
			}
			parseWithCloser := func(closer string) bool {
				i := strings.Index(p.text[p.pos:], closer)
				if i == -1 {
					return false
				}
				p.pos += i + len(closer)
				p.buf.push(piece{prependMarkup: []string{p.text[begin:p.pos]}})
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
					p.buf.push(piece{prependMarkup: []string{p.text[begin : p.pos+closer+2]}})
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
					autolink := autolinkRegexp.FindString(p.text[begin:])
					email := false
					if autolink == "" {
						autolink = emailAutolinkRegexp.FindString(p.text[begin:])
						email = true
					}
					if autolink != "" {
						p.pos = begin + len(autolink)
						text := p.syntax.Escape(autolink[1 : len(autolink)-1])
						dest := text
						if email {
							dest = "mailto:" + dest
						}
						tags := p.syntax.Link(dest, "")
						p.buf.push(piece{
							prependMarkup: []string{tags.Start},
							text:          text,
							appendMarkup:  []string{tags.End},
						})
						continue
					}
				}
			}
			parseText()
		case '&':
			entity := entityRegexp.FindString(p.text[begin:])
			if entity != "" {
				p.buf.push(piece{text: p.syntax.Escape(html.UnescapeString(entity))})
				p.pos = begin + len(entity)
			} else {
				parseText()
			}
		case '\\':
			if p.pos < len(p.text) && isASCIIPunct[p.text[p.pos]] {
				begin++
				p.pos++
			}
			parseText()
		case '\n':
			if len(p.buf.pieces) > 0 {
				last := &p.buf.pieces[len(p.buf.pieces)-1]
				if last.prependMarkup == nil && last.appendMarkup == nil {
					if p.pos == len(p.text) {
						last.text = strings.TrimRight(last.text, " ")
					} else {
						hardLineBreak := false
						if strings.HasSuffix(last.text, "\\") {
							hardLineBreak = true
							last.text = last.text[:len(last.text)-1]
						} else {
							hardLineBreak = strings.HasSuffix(last.text, "  ")
							last.text = strings.TrimRight(last.text, " ")
						}
						if hardLineBreak {
							p.buf.push(piece{prependMarkup: []string{"<br />"}})
						}
					}
				}
			}
			p.buf.push(piece{text: "\n"})
			for p.pos < len(p.text) && p.text[p.pos] == ' ' {
				p.pos++
			}
		default:
			parseText()
		}
	}
	p.processEmphasis(p.delims.bottom)
}

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
		strong := len(openerPiece.text) >= 2 && len(closerPiece.text) >= 2
		if strong {
			openerPiece.text = openerPiece.text[2:]
			openerPiece.appendMarkup = append(openerPiece.appendMarkup, p.syntax.StrongEmphasis.Start)
			closerPiece.text = closerPiece.text[2:]
			closerPiece.prependMarkup = append(closerPiece.prependMarkup, p.syntax.StrongEmphasis.End)
		} else {
			openerPiece.text = openerPiece.text[1:]
			openerPiece.appendMarkup = append(openerPiece.appendMarkup, p.syntax.Emphasis.Start)
			closerPiece.text = closerPiece.text[1:]
			closerPiece.prependMarkup = append(closerPiece.prependMarkup, p.syntax.Emphasis.End)
		}
		opener.next = closer
		closer.prev = opener
		if openerPiece.text == "" {
			opener.prev.next = opener.next
			opener.next.prev = opener.prev
		}
		if closerPiece.text == "" {
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

type linkTailParser struct {
	text string
	pos  int
}

func parseLinkTail(text string) (n int, dest, title string) {
	p := linkTailParser{text, 0}
	return p.parse()
}

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
				return -1, "", ""
			case '\\':
				titleBuilder.WriteByte(p.parseBackslash())
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
	return p.pos + 1, html.UnescapeString(destBuilder.String()), html.UnescapeString(titleBuilder.String())
}

func (p *linkTailParser) skipWhitespaces() {
	for p.pos < len(p.text) && isWhitespace(p.text[p.pos]) {
		p.pos++
	}
}

func isWhitespace(b byte) bool {
	switch b {
	case ' ', '\t', '\r', '\n':
		return true
	default:
		return false
	}
}

func (p *linkTailParser) parseBackslash() byte {
	if p.pos+1 < len(p.text) && isASCIIPunct[p.text[p.pos+1]] {
		b := p.text[p.pos+1]
		p.pos += 2
		return b
	}
	p.pos++
	return '\\'
}

func isASCIILetter(b byte) bool {
	return ('a' <= b && b <= 'z') || ('A' <= b && b <= 'Z')
}

func isASCIIControl(b byte) bool {
	return b < 0x20
}

func isMeta(b byte) bool {
	switch b {
	case '!', '[', ']', '*', '_', '`', '\\', '&', '<', '\n':
		return true
	default:
		return false
	}
}

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
		for j < len(s) && s[j] == '`' {
			j++
		}
		i = j
	}
	return -1
}

var lineEndingToSpace = strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ")

func normalizeCodeSpanContent(s string) string {
	s = lineEndingToSpace.Replace(s)
	if len(s) > 1 && s[0] == ' ' && s[len(s)-1] == ' ' && strings.Trim(s, " ") != "" {
		return s[1 : len(s)-1]
	}
	return s
}
