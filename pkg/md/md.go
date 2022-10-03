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
		`^ {0,3}((?:-[ \t]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})$`)

	// Capture group 1: heading opener
	atxHeadingRegexp       = regexp.MustCompile(`^ {0,3}(#{1,6})(?:[ \t]|$)`)
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

	indentedCodePrefix = "    "
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
		} else if len(p.paragraph) == 0 && strings.HasPrefix(line, indentedCodePrefix) {
			p.popParagraph(matchedContainers)
			p.popList()
			p.parseIndentedCodeBlock(line)
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
	blockquoteMarkerRegexp = regexp.MustCompile(`^ {0,3}> ?`)

	containerStartingMarkerRegexp = regexp.MustCompile(
		// Capture groups:
		// 1. bullet item punctuation
		// 2. ordered item start index
		// 3. ordered item punctuation
		// 4. trailing spaces
		`^ {0,3}(?:([-+*])|([0-9]{1,9})([.)]))( +)`)
	itemStartingMarkerBlankLineRegexp = regexp.MustCompile(
		// Capture groups are the same, with group 4 always empty.
		`^ {0,3}(?:([-+*])|([0-9]{1,9})([.)]))[ \t]*()$`)
)

// Parses starting markers of container blocks. Returns the line after removing
// all starting markers and new containers to create.
func parseStartingMarkers(line string, newParagraph bool) (string, []container) {
	var containers []container
	// Don't parse thematic breaks like "- - - " as three bullets.
	for !thematicBreakRegexp.MatchString(line) {
		if bqMarker := blockquoteMarkerRegexp.FindString(line); bqMarker != "" {
			line = line[len(bqMarker):]
			containers = append(containers, container{typ: blockquote})
			continue
		}

		m := containerStartingMarkerRegexp.FindStringSubmatch(line)
		blankFirst := false
		if m == nil && newParagraph {
			m = itemStartingMarkerBlankLineRegexp.FindStringSubmatch(line)
			blankFirst = true
		}
		if m == nil {
			break
		}
		marker, bulletPunct, orderedStart, orderedPunct, spaces := m[0], m[1], m[2], m[3], m[4]
		if len(spaces) >= 5 {
			marker = marker[:len(marker)-len(spaces)+1]
		}

		indent := len(marker)
		if strings.Trim(line[len(marker):], " \t") == "" {
			indent = len(strings.TrimRight(marker, " \t")) + 1
		}
		c := container{indent: strings.Repeat(" ", indent), blankFirst: blankFirst}
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

		line = line[len(marker):]
		containers = append(containers, c)
	}
	return line, containers
}

func isBlankLine(line string) bool {
	return strings.Trim(line, " \t") == ""
}

func (p *blockParser) parseFencedCodeBlock(indent int, opener, info string) {
	tags := p.syntax.CodeBlock(processCodeFenceInfo(strings.Trim(info, " \t")))
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

// Code fence info strings are mostly verbatim, but support backslash and
// entities. This mirrors part of (*inlineParser).render.
func processCodeFenceInfo(text string) string {
	pos := 0
	var sb strings.Builder
	for pos < len(text) {
		b := text[pos]
		if b == '&' {
			if entity := entityRegexp.FindString(text[pos:]); entity != "" {
				sb.WriteString(html.UnescapeString(entity))
				pos += len(entity)
				continue
			}
		} else if b == '\\' && pos+1 < len(text) && isASCIIPunct(text[pos+1]) {
			b = text[pos+1]
			pos++
		}
		sb.WriteByte(b)
		pos++
	}
	return sb.String()
}

func (p *blockParser) parseIndentedCodeBlock(line string) {
	tags := p.syntax.CodeBlock("")
	p.sb.WriteString(tags.Start)
	p.sb.WriteString(p.syntax.Escape(strings.TrimPrefix(line, indentedCodePrefix)))
	p.sb.WriteByte('\n')
	var savedBlankLines []string
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
			if strings.HasPrefix(line, indentedCodePrefix) {
				line = strings.TrimPrefix(line, indentedCodePrefix)
			} else {
				line = ""
			}
			savedBlankLines = append(savedBlankLines, line)
			continue
		} else if matchedContainers < len(p.containers) || !strings.HasPrefix(line, indentedCodePrefix) {
			p.lines.backup()
			break
		}
		for _, blankLine := range savedBlankLines {
			p.sb.WriteString(blankLine)
			p.sb.WriteByte('\n')
		}
		savedBlankLines = savedBlankLines[:0]
		p.sb.WriteString(p.syntax.Escape(strings.TrimPrefix(line, indentedCodePrefix)))
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
