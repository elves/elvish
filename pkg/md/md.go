// Package md implements a Markdown renderer.
//
// This package implements most of the CommonMark spec, with the following
// omissions:
//
//   - "\r" and "\r\n" are not supported as line endings. This can be easily
//     worked around by converting them to "\n" first.
//
//   - Tabs are not supported for defining block structures; use spaces instead.
//     Tabs in other context are supported.
//
//   - Only entities that are necessary for writing valid HTML (&lt; &gt;
//     &quote; &apos; &amp;) are supported. This aspect can be controlled by
//     overriding the UnescapeEntities variable.
//
//   - Setext headings are not supported; use ATX headings instead.
//
//   - Reference links are not supported; use inline links instead.
//
//   - Lists are always considered loose.
//
// All other features are supported, with CommonMark spec tests passing; see
// test file for which tests are skipped. The spec tests are taken from the HEAD
// of the CommonMark spec in https://github.com/commonmark/commonmark-spec,
// which may differ slightly from the latest released version.
//
// This package is not used anywhere in Elvish right now. It is intended to be
// used for rendering the elvdoc of builtin modules inside terminals.
package md

//go:generate stringer -type=OpType -output=zstring.go

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// UnescapeEntities is used to unescape HTML entities and numeric character
// references.
//
// The default value only unescapes the entities that are necessary when writing
// valid HTML. It can be set to html.UnescapeString for better CommonMark
// compliance.
var UnescapeEntities = strings.NewReplacer(
	"&lt;", "<", "&gt;", ">", "&quote;", `"`, "&apos;", `'`, "&amp;", "&",
).Replace

// Codec is used to render output.
type Codec interface {
	Do(Op)
}

// Op represents an operation for the Codec.
type Op struct {
	Type OpType
	// For OpOrderedListStart (the start number) or OpHeadingStart/OpHeadingEnd
	// (as the heading level)
	Number int
	// For OpText, OpLine, OpRawHTML, OpCodeBlockStart (as the processed info
	// string), OpLinkStart, OpImage (as the title in the last two)
	Text string
	// For OpLinkStart, OpImage
	Dest string
	// ForOpImage
	Alt string
}

// OpType enumerates possible types of an Op.
type OpType uint

// Possible output operations.
const (
	// Text elements.
	OpText OpType = iota
	OpRawHTML

	// Leaf blocks.
	OpThematicBreak
	OpHeadingStart
	OpHeadingEnd
	OpCodeBlockStart
	OpCodeBlockEnd
	OpParagraphStart
	OpParagraphEnd

	// Container blocks.
	OpBlockquoteStart
	OpBlockquoteEnd
	OpListItemStart
	OpListItemEnd
	OpBulletListStart
	OpBulletListEnd
	OpOrderedListStart
	OpOrderedListEnd

	// Inline elements.
	OpCodeSpanStart
	OpCodeSpanEnd
	OpEmphasisStart
	OpEmphasisEnd
	OpStrongEmphasisStart
	OpStrongEmphasisEnd
	OpLinkStart
	OpLinkEnd
	OpImage
	OpHardLineBreak
)

// Render parses markdown and renders it according to the output syntax.
func Render(text string, codec Codec) {
	p := blockParser{lines: lineSplitter{text, 0}, codec: codec}
	p.render()
}

type blockParser struct {
	lines lineSplitter
	codec Codec
	tree  blockTree
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

	html6Regexp = regexp.MustCompile(`^ {0,3}</?(?i:address|article|aside|base|basefont|blockquote|body|caption|center|col|colgroup|dd|details|dialog|dir|div|dl|dt|fieldset|figcaption|figure|footer|form|frame|frameset|h1|h2|h3|h4|h5|h6|head|header|hr|html|iframe|legend|li|link|main|menu|menuitem|nav|noframes|ol|optgroup|option|p|param|section|summary|table|tbody|td|tfoot|th|thead|title|tr|track|ul)(?:[ \t>]|$|/>)`)
	html7Regexp = regexp.MustCompile(
		fmt.Sprintf(`^ {0,3}(?:%s|%s)[ \t]*$`, openTag, closingTag))
)

const indentedCodePrefix = "    "

func (p *blockParser) render() {
	for p.lines.more() {
		line := p.lines.next()
		line, matchedContainers, newItem := p.tree.processContainerMarkers(line, p.codec)

		if isBlankLine(line) {
			// Blank lines terminate blockquote if the continuation marker is
			// absent.
			if i, unmatched := p.tree.unmatchedBlockquote(matchedContainers); unmatched {
				p.tree.closeBlocks(i, p.codec)
				continue
			}
			if newItem && p.lines.more() {
				// A list item can start with at most one blank line; the second
				// blank closes it.
				nextLine := p.lines.next()
				nextLine, _ = p.tree.matchContinuationMarkers(nextLine)
				p.lines.backup()
				if isBlankLine(nextLine) {
					p.tree.closeBlocks(len(p.tree.containers)-1, p.codec)
				}
			}
			p.tree.closeParagraph(p.codec)
		} else if thematicBreakRegexp.MatchString(line) {
			p.tree.closeBlocks(matchedContainers, p.codec)
			p.codec.Do(Op{Type: OpThematicBreak, Text: line})
		} else if m := atxHeadingRegexp.FindStringSubmatchIndex(line); m != nil {
			p.tree.closeBlocks(matchedContainers, p.codec)
			openerStart, openerEnd := m[2], m[3]
			opener := line[openerStart:openerEnd]
			line = strings.TrimRight(line[openerEnd:], " \t")
			if closer := atxHeadingCloserRegexp.FindString(line); closer != "" {
				line = line[:len(line)-len(closer)]
			}
			level := len(opener)
			p.codec.Do(Op{Type: OpHeadingStart, Number: level})
			renderInline(strings.Trim(line, " \t"), p.codec)
			p.codec.Do(Op{Type: OpHeadingEnd, Number: level})
		} else if m := codeFenceRegexp.FindStringSubmatch(line); m != nil {
			p.tree.closeBlocks(matchedContainers, p.codec)
			indent, opener, info := len(m[1]), m[2], m[3]
			if opener == "" {
				opener, info = m[4], m[5]
			}
			p.parseFencedCodeBlock(indent, opener, info)
		} else if len(p.tree.paragraph) == 0 && strings.HasPrefix(line, indentedCodePrefix) {
			p.tree.closeBlocks(matchedContainers, p.codec)
			p.parseIndentedCodeBlock(line)
		} else if html1Regexp.MatchString(line) {
			p.tree.closeBlocks(matchedContainers, p.codec)
			p.parseHTMLBlock(line, html1CloserRegexp.MatchString)
		} else if html2Regexp.MatchString(line) {
			p.tree.closeBlocks(matchedContainers, p.codec)
			p.parseHTMLBlock(line, html2CloserRegexp.MatchString)
		} else if html3Regexp.MatchString(line) {
			p.tree.closeBlocks(matchedContainers, p.codec)
			p.parseHTMLBlock(line, html3CloserRegexp.MatchString)
		} else if html4Regexp.MatchString(line) {
			p.tree.closeBlocks(matchedContainers, p.codec)
			p.parseHTMLBlock(line, html4CloserRegexp.MatchString)
		} else if html5Regexp.MatchString(line) {
			p.tree.closeBlocks(matchedContainers, p.codec)
			p.parseHTMLBlock(line, html5CloserRegexp.MatchString)
		} else if html6Regexp.MatchString(line) || (len(p.tree.paragraph) == 0 && html7Regexp.MatchString(line)) {
			p.tree.closeBlocks(matchedContainers, p.codec)
			p.parseBlankLineTerminatedHTMLBlock(line)
		} else {
			if len(p.tree.paragraph) == 0 {
				// This is not lazy continuation, so close all unmatched
				// containers.
				p.tree.closeBlocks(matchedContainers, p.codec)
			}
			p.tree.paragraph = append(p.tree.paragraph, line)
		}
	}
	p.tree.closeBlocks(0, p.codec)
}

func isBlankLine(line string) bool {
	return strings.Trim(line, " \t") == ""
}

func (p *blockParser) parseFencedCodeBlock(indent int, opener, info string) {
	info = processCodeFenceInfo(strings.Trim(info, " \t"))
	p.codec.Do(Op{Type: OpCodeBlockStart, Text: info})
	for p.lines.more() {
		line := p.lines.next()
		line, matchedContainers := p.tree.matchContinuationMarkers(line)
		if isBlankLine(line) {
			if i, unmatched := p.tree.unmatchedBlockquote(matchedContainers); unmatched {
				do(p.codec, OpCodeBlockEnd)
				p.tree.closeBlocks(i, p.codec)
				return
			}
		} else if matchedContainers < len(p.tree.containers) {
			p.lines.backup()
			break
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
		doLine(p.codec, line)
	}
	do(p.codec, OpCodeBlockEnd)
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
				sb.WriteString(UnescapeEntities(entity))
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
	do(p.codec, OpCodeBlockStart)
	doLine(p.codec, strings.TrimPrefix(line, indentedCodePrefix))
	var savedBlankLines []string
	for p.lines.more() {
		line := p.lines.next()
		line, matchedContainers := p.tree.matchContinuationMarkers(line)
		if isBlankLine(line) {
			if i, unmatched := p.tree.unmatchedBlockquote(matchedContainers); unmatched {
				do(p.codec, OpCodeBlockEnd)
				p.tree.closeBlocks(i, p.codec)
				return
			}
			if strings.HasPrefix(line, indentedCodePrefix) {
				line = strings.TrimPrefix(line, indentedCodePrefix)
			} else {
				line = ""
			}
			savedBlankLines = append(savedBlankLines, line)
			continue
		} else if matchedContainers < len(p.tree.containers) || !strings.HasPrefix(line, indentedCodePrefix) {
			p.lines.backup()
			break
		}
		for _, blankLine := range savedBlankLines {
			doLine(p.codec, blankLine)
		}
		savedBlankLines = savedBlankLines[:0]
		doLine(p.codec, strings.TrimPrefix(line, indentedCodePrefix))
	}
	do(p.codec, OpCodeBlockEnd)
}

func (p *blockParser) parseHTMLBlock(line string, closer func(string) bool) {
	doRawHTMLLine(p.codec, line)
	if closer(line) {
		return
	}
	for p.lines.more() {
		line := p.lines.next()
		line, matchedContainers := p.tree.matchContinuationMarkers(line)
		if isBlankLine(line) {
			if i, unmatched := p.tree.unmatchedBlockquote(matchedContainers); unmatched {
				p.tree.closeBlocks(i, p.codec)
				return
			}
		} else if matchedContainers < len(p.tree.containers) {
			p.lines.backup()
			return
		}
		doRawHTMLLine(p.codec, line)
		if closer(line) {
			return
		}
	}
}

func (p *blockParser) parseBlankLineTerminatedHTMLBlock(line string) {
	doRawHTMLLine(p.codec, line)
	for p.lines.more() {
		line := p.lines.next()
		line, matchedContainers := p.tree.matchContinuationMarkers(line)
		if isBlankLine(line) {
			if i, unmatched := p.tree.unmatchedBlockquote(matchedContainers); unmatched {
				p.tree.closeBlocks(i, p.codec)
			}
			return
		} else if matchedContainers < len(p.tree.containers) {
			p.lines.backup()
			return
		}
		doRawHTMLLine(p.codec, line)
	}
}

// This struct corresponds to the block tree in
// https://spec.commonmark.org/0.30/#phase-1-block-structure.
//
// The spec describes a two-phased parsing strategy where the entire block tree
// is built before inline parsing is done. However, since we don't support
// setext headings and link reference definitions, and treats all lists as
// loose, the rendering result of closed blocks will never be impacted by future
// blocks. This enables us to render as we parse, and allows us to only track
// the path of currently open blocks, which is the same as the rightmost path in
// the full block tree at any given point in time.
//
// The path consists of zero or more container nodes, and an optional paragraph
// node. The paragraph node exists if and only if it contains at least 1 line;
// the spec prohibits paragraphs consisting of 0 lines.
//
// We don't need to track any other type of leaf blocks, because they all have
// simple termination conditions, so can be parsed in one iteration of the main
// parsing loop, as a nested loop that consumes lines until the block
// terminates.
//
// Paragraphs, however, don't have a simple termination condition. Other than
// the common condition of being terminated as part of the container block,
// paragraphs are always terminated by *another* type of leaf block. This means
// that the logic for deciding to continue or interrupt of a paragraph lives
// within the main parsing loop. This in turn makes it necessary to store the
// lines of the paragraph across iterations of the main parsing loop, hence part
// of the parser's state.
type blockTree struct {
	containers []container
	paragraph  []string
}

// Processes container markers at the start of the line, which consists of
// continuation markers of existing containers and starting markers of new
// containers.
//
// Returns the line after removing both types of markers, the number of markers
// matched or parsed, and whether the innermost container is a newly opened list
// item.
//
// The latter should be used to call t.closeContainers
// unless the remaining content of the line constitutes a blank line or
// paragraph continuation.
func (t *blockTree) processContainerMarkers(line string, codec Codec) (string, int, bool) {
	line, matched := t.matchContinuationMarkers(line)
	line, newContainers := t.parseStartingMarkers(line,
		// This argument tells parseStartingMarkers whether we are starting a
		// new paragraph. This seems straightforward enough: if the paragraph is
		// empty is the first place, or if we are going to terminate some
		// containers, we are starting a new paragraph.
		//
		// The second part of the condition is more subtle though. If the
		// remaining content of the line constitutes paragraph continuation, we
		// are not starting a new paragraph. We are only able to ignore this
		// case parseStartingMarkers only uses this condition when it actually
		// parses a starting marker, meaning that the line cannot be paragraph
		// continuation.
		len(t.paragraph) == 0 || matched != len(t.containers))

	continueList := false
	if matched > 0 && t.containers[matched-1].typ.isList() {
		// If the last matched container is a list (i.e. the first unmatched
		// container is a list item), keep it if and only if the first
		// container to add is a list item that can continue the list.
		continueList = len(newContainers) > 0 && newContainers[0].punct == t.containers[matched-1].punct
		if !continueList {
			matched--
		}
	}

	if len(newContainers) == 0 {
		return line, matched, false
	}

	t.closeBlocks(matched, codec)
	for _, c := range newContainers {
		if c.typ.isItem() {
			if continueList {
				continueList = false
			} else {
				list := container{typ: c.typ.itemToList(), punct: c.punct, start: c.start}
				t.containers = append(t.containers, list)
				codec.Do(Op{Type: containerOpenOp[list.typ], Number: list.start})
			}
		}
		t.containers = append(t.containers, c)
		do(codec, containerOpenOp[c.typ])
	}
	return line, len(t.containers), newContainers[len(newContainers)-1].typ.isItem()
}

// Matches the continuation markers of existing container nodes. Returns the
// line after removing all matched continuation markers and the number of
// containers matched.
func (t *blockTree) matchContinuationMarkers(line string) (string, int) {
	for i, container := range t.containers {
		markerLen, matched := container.matchContinuationMarker(line)
		if !matched {
			return line, i
		}
		line = line[markerLen:]
	}
	return line, len(t.containers)
}

// Finds the first blockquote container after skipping matched containers.
// Returns len(t.containers), false if not found.
//
// This is used for handling blank lines. Blank lines do not close list item
// blocks (except when a blank line follows a list item starting with a blank
// item), but they do close blockquote blocks if the continuation marker is
// missing.
func (t *blockTree) unmatchedBlockquote(matched int) (int, bool) {
	for i := matched; i < len(t.containers); i++ {
		if t.containers[i].typ == blockquote {
			return i, true
		}
	}
	return len(t.containers), false
}

var (
	// https://spec.commonmark.org/0.30/#block-quotes
	blockquoteMarkerRegexp = regexp.MustCompile(`^ {0,3}> ?`)

	// Rule #1 and #2 of https://spec.commonmark.org/0.30/#list-items
	itemStartingMarkerRegexp = regexp.MustCompile(
		// Capture groups:
		// 1. bullet item punctuation
		// 2. ordered item start index
		// 3. ordered item punctuation
		// 4. trailing spaces
		`^ {0,3}(?:([-+*])|([0-9]{1,9})([.)]))( +)`)

	// Rule #3 of https://spec.commonmark.org/0.30/#list-items
	itemStartingMarkerBlankLineRegexp = regexp.MustCompile(
		// Capture groups are the same, with group 4 always empty.
		`^ {0,3}(?:([-+*])|([0-9]{1,9})([.)]))[ \t]*()$`)
)

// Parses starting markers of container blocks. Returns the line after removing
// all starting markers and new containers to create.
//
// Blockquotes are simple to parse. Most of the code deals with list items,
// described in https://spec.commonmark.org/0.30/#list-items.
func (t *blockTree) parseStartingMarkers(line string, newParagraph bool) (string, []container) {
	var containers []container
	// Exception 2 of rule #1: Don't parse thematic breaks like "- - - " as
	// three bullets.
	for !thematicBreakRegexp.MatchString(line) {
		if bqMarker := blockquoteMarkerRegexp.FindString(line); bqMarker != "" {
			line = line[len(bqMarker):]
			containers = append(containers, container{typ: blockquote})
			continue
		}

		m := itemStartingMarkerRegexp.FindStringSubmatch(line)
		if m == nil && newParagraph {
			m = itemStartingMarkerBlankLineRegexp.FindStringSubmatch(line)
		}
		if m == nil {
			break
		}
		marker, bulletPunct, orderedStart, orderedPunct, spaces := m[0], m[1], m[2], m[3], m[4]
		if len(spaces) >= 5 {
			// Rule #2 applies; only the first space is as part of the marker.
			marker = marker[:len(marker)-len(spaces)+1]
		}

		indent := len(marker)
		if strings.Trim(line[len(marker):], " \t") == "" {
			// Rule #3 applies: indent is exactly one space, regardless of how
			// many spaces there actually are, which can be 0.
			indent = len(strings.TrimRight(marker, " \t")) + 1
		}
		c := container{continuation: strings.Repeat(" ", indent)}
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

func (t *blockTree) closeBlocks(keep int, codec Codec) {
	t.closeParagraph(codec)
	for i := len(t.containers) - 1; i >= keep; i-- {
		do(codec, containerCloseOp[t.containers[i].typ])
	}
	t.containers = t.containers[:keep]
}

func (t *blockTree) closeParagraph(codec Codec) {
	if len(t.paragraph) == 0 {
		return
	}
	do(codec, OpParagraphStart)
	renderInline(strings.Trim(strings.Join(t.paragraph, "\n"), " \t"), codec)
	t.paragraph = t.paragraph[:0]
	do(codec, OpParagraphEnd)
}

type container struct {
	typ          containerType
	punct        byte
	start        int
	continuation string
}

type containerType uint8

const (
	blockquote containerType = iota
	bulletList
	bulletItem
	orderedList
	orderedItem
)

func (t containerType) isList() bool { return t == bulletList || t == orderedList }

func (t containerType) isItem() bool { return t == bulletItem || t == orderedItem }

func (t containerType) itemToList() containerType {
	if t == bulletItem {
		return bulletList
	} else {
		return orderedList
	}
}

var (
	containerOpenOp = []OpType{
		blockquote:  OpBlockquoteStart,
		bulletList:  OpBulletListStart,
		bulletItem:  OpListItemStart,
		orderedList: OpOrderedListStart,
		orderedItem: OpListItemStart,
	}
	containerCloseOp = []OpType{
		blockquote:  OpBlockquoteEnd,
		bulletList:  OpBulletListEnd,
		bulletItem:  OpListItemEnd,
		orderedList: OpOrderedListEnd,
		orderedItem: OpListItemEnd,
	}
)

func (c container) matchContinuationMarker(line string) (int, bool) {
	switch c.typ {
	case blockquote:
		marker := blockquoteMarkerRegexp.FindString(line)
		return len(marker), marker != ""
	case bulletList, orderedList:
		return 0, true
	case bulletItem, orderedItem:
		if strings.HasPrefix(line, c.continuation) {
			return len(c.continuation), true
		}
		return 0, false
	}
	panic("unreachable")
}

// Provides support for consuming a string line by line.
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

// Codec shorthands.

func do(c Codec, t OpType) { c.Do(Op{Type: t}) }

func doLine(c Codec, s string) {
	c.Do(Op{Type: OpText, Text: s})
	c.Do(Op{Type: OpText, Text: "\n"})
}

func doRawHTMLLine(c Codec, s string) {
	c.Do(Op{Type: OpRawHTML, Text: s})
	c.Do(Op{Type: OpText, Text: "\n"})
}
