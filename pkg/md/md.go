// Package md implements a Markdown parser.
//
// To use this package, call [Render] with one of the [Codec] implementations:
//
//   - [HTMLCodec] converts Markdown to HTML. This is used in
//     [src.elv.sh/website/cmd/md2html], part of Elvish's website toolchain.
//
//   - [FmtCodec] formats Markdown. This is used in [src.elv.sh/cmd/elvmdfmt],
//     used for formatting Markdown files in the Elvish repo.
//
//   - [TTYCodec] renders Markdown in the terminal. This will be used in a help
//     system that can used directly from Elvish to render documentation of
//     Elvish modules.
//
// # Why another Markdown implementation?
//
// The Elvish project uses Markdown in the documentation ("[elvdoc]") for the
// functions and variables defined in builtin modules. These docs are then
// converted to HTML as part of the website; for example, you can read the docs
// for builtin functions and variables at https://elv.sh/ref/builtin.html.
//
// We used to use [Pandoc] to convert the docs from their Markdown sources to
// HTML. However, we would also like to expand the elvdoc system in two ways:
//
//   - We would like to support elvdocs in user-defined modules, not just
//     builtin modules.
//
//   - We would like to allow users to read elvdocs directly from the Elvish
//     program, in the terminal, without needing a browser or an Internet
//     connection.
//
// With these requirements, Elvish itself needs to know how to parse Markdown
// sources and render them in the terminal, so we need a Go implementation
// instead. There is a good Go implementation, [github.com/yuin/goldmark], but
// it is quite large: linking it into Elvish will increase the binary size by
// more than 1MB. (There is another popular Markdown implementation,
// [github.com/russross/blackfriday/v2], but it doesn't support CommonMark.)
//
// By having a more narrow focus, this package is much smaller than goldmark,
// and can be easily optimized for Elvish's use cases. In contrast to goldmark's
// 1MB, including [Render] and [HTMLCodec] in Elvish only increases the binary
// size by 150KB. That said, the functionalities provided by this package still
// try to be as general as possible, and can potentially be used by other people
// interested in a small Markdown implementation.
//
// Besides elvdocs, Pandoc was also used to convert all the other content on the
// Elvish website (https://elv.sh) to HTML. Additionally, [Prettier] used to be
// used to format all the Markdown files in the repo. Now that Elvish has its
// own Markdown implementation, we can use it not just for rendering elvdocs in
// the terminal, but also replace the use of Pandoc and Prettier. These external
// tools are decent, but using them still came with some frictions:
//
//   - Even though both are relatively easy to set up, they can still be a
//     hindrance to casual contributors.
//
//   - Since behavior of these tools can change with version, we explicit
//     specify their versions in both CI configurations and [contributing
//     instructions]. But this creates another problem: every time these tools
//     release new versions, we have to manually bump the versions, and every
//     contributor also needs to manually update them in their development
//     environments.
//
// Replacing external tools with this package removes these frictions.
//
// Additionally, this package is very easy to extend and optimize to suit
// Elvish's needs:
//
//   - We used to custom Pandoc using a mix of shell scripts, templates and Lua
//     scripts. While these customization options of Pandoc are well documented,
//     they are not something people are likely to be familiar with.
//
//     With this implementation, everything is now done with Go code.
//
//   - The Markdown formatter is much faster than Prettier, so it's now feasible
//     to run the formatter every time when saving a Markdown file.
//
// # Which Markdown variant does this package implement?
//
// This package implements a large subset of the [CommonMark] spec, with the
// following omissions:
//
//   - "\r" and "\r\n" are not supported as line endings. This can be easily
//     worked around by converting them to "\n" first.
//
//   - Tabs are not supported for defining block structures; use spaces instead.
//     Tabs in other context are supported.
//
//   - Among HTML entities, only a few are supported: &lt; &gt; &quote; &apos;
//     &amp;. This is because the full list of HTML entities is very large and
//     will inflate the binary size.
//
//     If full support for HTML entities are desirable, this can be done by
//     overriding the [UnescapeHTML] variable with [html.UnescapeString].
//
//     (Numeric character references like &#9; and &#x20; are fully supported.)
//
//   - [Setext headings] are not supported; use [ATX headings] instead.
//
//   - [Reference links] are not supported; use [inline links] instead.
//
//   - Lists are always considered [loose].
//
// The package also supports the following extensions:
//
//   - ATX headers may be followed by [Pandoc header attributes] {...}.
//
// These omitted features are never used in Elvish's Markdown sources.
//
// All implemented features pass their relevant CommonMark spec tests, currently
// targeting [CommonMark 0.31.2]. See [testutils_test.go] for a complete list of
// which spec tests are skipped.
//
// # Is this package useful outside Elvish?
//
// Yes! Well, hopefully. Assuming you don't use the features this package omits,
// it can be useful in at least the following ways:
//
//   - The implementation is quite lightweight, so you can use it instead of a
//     more full-features Markdown library if small binary size is important.
//
//     As shown above, the increase in binary size when using this package in
//     Elvish is about 150KB, compared to more than 1MB when using
//     [github.com/yuin/goldmark]. You mileage may vary though, since the binary
//     size increase depends on which packages the binary is already including.
//
//   - The formatter implemented by [FmtCodec] is heavily fuzz-tested to ensure
//     that it does not alter the semantics of the Markdown.
//
//     Markdown formatting is fraught with tricky edge cases. For example, if a
//     formatter standardizes all bullet markers to "-", it might reformat "*
//     --" to "- ---", but the latter will now be parsed as a thematic break.
//
//     Thanks to Go's builtin [fuzzing support], the formatter is able to handle
//     many such corner cases (at least [all the corner cases found by the
//     fuzzer]; take a look and try them on other formatters!). There are two
//     areas - namely nested and consecutive emphasis or strong emphasis - that
//     are just too tricky to get 100% right that the formatter is not
//     guaranteed to be correct; the fuzz test explicitly skips those cases.
//
//     Nonetheless, if you are writing a Markdown formatter and care about
//     correctness, the corner cases will be interesting, regardless of which
//     language you are using to implement the formatter.
//
// [Pandoc header attributes]: https://pandoc.org/MANUAL.html#extension-header_attributes
// [all the corner cases found by the fuzzer]: https://github.com/elves/elvish/tree/master/pkg/md/testdata/fuzz/FuzzFmtPreservesHTMLRender
// [fuzzing support]: https://go.dev/security/fuzz/
// [loose]: https://spec.commonmark.org/0.31.2/#loose
// [Setext headings]: https://spec.commonmark.org/0.31.2/#setext-headings
// [ATX headings]: https://spec.commonmark.org/0.31.2/#atx-headings
// [testutils_test.go]: https://github.com/elves/elvish/blob/master/pkg/md/testutils_test.go
// [elvdoc]: https://github.com/elves/elvish/blob/master/CONTRIBUTING.md#reference-docs
// [Pandoc]: https://pandoc.org
// [Prettier]: https://prettier.io
// [CommonMark]: https://spec.commonmark.org
// [contributing instructions]: https://github.com/elves/elvish/blob/master/CONTRIBUTING.md
// [inline links]: https://spec.commonmark.org/0.31.2/#inline-link
// [Reference links]: https://spec.commonmark.org/0.31.2/#reference-link
// [CommonMark 0.31.2]: https://spec.commonmark.org/0.31.2/
package md

//go:generate stringer -type=OpType,InlineOpType -output=zstring.go

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// UnescapeHTML is used by the parser to unescape HTML entities and numeric
// character references.
//
// The default implementation supports numeric character references, plus a
// minimal set of entities that are necessary for writing valid HTML or can
// appear in the output of FmtCodec. It can be set to html.UnescapeString for
// better CommonMark compliance.
var UnescapeHTML = unescapeHTML

// https://spec.commonmark.org/0.31.2/#entity-and-numeric-character-references
const charRefPattern = `&(?:[a-zA-Z0-9]+|#[0-9]{1,7}|#[xX][0-9a-fA-F]{1,6});`

var charRefRegexp = regexp.MustCompile(charRefPattern)

var entities = map[string]rune{
	// Necessary for writing valid HTML
	"lt": '<', "gt": '>', "quote": '"', "apos": '\'', "amp": '&',
	// Not strictly necessary, but could be output by FmtCodec for slightly
	// nicer text
	"Tab": '\t', "NewLine": '\n', "nbsp": '\u00A0',
}

func unescapeHTML(s string) string {
	return charRefRegexp.ReplaceAllStringFunc(s, func(entity string) string {
		body := entity[1 : len(entity)-1]
		if r, ok := entities[body]; ok {
			return string(r)
		} else if body[0] == '#' {
			if body[1] == 'x' || body[1] == 'X' {
				if num, err := strconv.ParseInt(body[2:], 16, 32); err == nil {
					return string(rune(num))
				}
			} else {
				if num, err := strconv.ParseInt(body[1:], 10, 32); err == nil {
					return string(rune(num))
				}
			}
		}
		return entity
	})
}

// Codec is used to render output.
type Codec interface {
	Do(Op)
}

// Op represents an operation for the Codec.
type Op struct {
	Type OpType
	// 1-based line number. If the Op spans multiple lines, this identifies the
	// first line. For the *End types, this identifies the first line that
	// causes the block to be terminated, which can be the first line of another
	// block.
	LineNo int
	// For OpOrderedListStart (the start number) or OpHeading (as the heading
	// level)
	Number int
	// For OpHeading (attributes inside { }) and OpCodeBlock (text after opening
	// fence)
	Info string
	// For OpCodeBlock and OpHTMLBlock
	Lines []string
	// For OpParagraph and OpHeading
	Content []InlineOp
}

// OpType enumerates possible types of an Op.
type OpType uint

// Possible output operations.
const (
	// Leaf blocks.
	OpThematicBreak OpType = iota
	OpHeading
	OpCodeBlock
	OpHTMLBlock
	OpParagraph

	// Container blocks.
	OpBlockquoteStart
	OpBlockquoteEnd
	OpListItemStart
	OpListItemEnd
	OpBulletListStart
	OpBulletListEnd
	OpOrderedListStart
	OpOrderedListEnd
)

var initRegexpsOnce sync.Once

// Render parses markdown and renders it with a [Codec].
func Render(text string, codec Codec) {
	// Compiled regular expressions live on the heap. Compiling them lazily
	// saves memory if this function is never called.
	initRegexpsOnce.Do(initRegexps)
	p := blockParser{lines: lineSplitter{text, 0, 0}, codec: codec}
	p.render()
}

// StringerCodec is a [Codec] that also implements the String method.
type StringerCodec interface {
	Codec
	String() string
}

// Render calls Render(text, codec) and returns codec.String(). This can be a
// bit more convenient to use than [Render].
func RenderString(text string, codec StringerCodec) string {
	Render(text, codec)
	return codec.String()
}

type blockParser struct {
	lines lineSplitter
	codec Codec
	tree  blockTree
}

// Block regexps.
var thematicBreakRegexp,
	atxHeadingRegexp,
	atxHeadingCloserRegexp,
	atxHeadingAttributeRegexp,
	codeFenceRegexp,
	codeFenceCloserRegexp,
	html1Regexp,
	html1CloserRegexp,
	html2Regexp,
	html2CloserRegexp,
	html3Regexp,
	html3CloserRegexp,
	html4Regexp,
	html4CloserRegexp,
	html5Regexp,
	html5CloserRegexp,
	html6Regexp,
	html7Regexp *regexp.Regexp

// Inline regexps.
var uriAutolinkRegexp,
	emailAutolinkRegexp,
	openTagRegexp,
	closingTagRegexp *regexp.Regexp

// Building blocks for regexps.
const (
	scheme           = `[a-zA-Z][a-zA-Z0-9+.-]{1,31}`
	emailLocalPuncts = ".!#$%&'*+/=?^_`{|}~-"

	// https://spec.commonmark.org/0.31.2/#open-tag
	openTag = `<` +
		`[a-zA-Z][a-zA-Z0-9-]*` + // tag name
		(`(?:` +
			`[ \t\n]+` + // whitespace
			`[a-zA-Z_:][a-zA-Z0-9_\.:-]*` + // attribute name
			`(?:[ \t\n]*=[ \t\n]*(?:[^ \t\n"'=<>` + "`" + `]+|'[^']*'|"[^"]*"))?` + // attribute value specification
			`)*`) + // zero or more attributes
		`[ \t\n]*` + // whitespace
		`/?>`
	// https://spec.commonmark.org/0.31.2/#closing-tag
	closingTag = `</[a-zA-Z][a-zA-Z0-9-]*[ \t\n]*>`
)

func initRegexps() {
	thematicBreakRegexp = regexp.MustCompile(
		`^ {0,3}((?:-[ \t]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})$`)

	// Capture group 1: heading opener
	atxHeadingRegexp = regexp.MustCompile(`^ {0,3}(#{1,6})(?:[ \t]|$)`)
	atxHeadingCloserRegexp = regexp.MustCompile(`[ \t]#+[ \t]*$`)

	// Support the header_attributes extension
	// (https://pandoc.org/MANUAL.html#extension-header_attributes). Like
	// pandoc, attributes appear *after* the optional heading closer.
	//
	// Attributes are stored in the info string and interpreted by the Codec.
	atxHeadingAttributeRegexp = regexp.MustCompile(` {([^}]+)}$`)

	// Capture groups:
	// 1. Indent
	// 2. Fence punctuations (backquote fence)
	// 3. Untrimmed info string (backquote fence)
	// 4. Fence punctuations (tilde fence)
	// 5. Untrimmed info string (tilde fence)
	codeFenceRegexp = regexp.MustCompile("(^ {0,3})(?:(`{3,})([^`]*)|(~{3,})(.*))$")
	// Capture group 1: fence punctuations
	codeFenceCloserRegexp = regexp.MustCompile("(?:^ {0,3})(`{3,}|~{3,})[ \t]*$")

	// These corresponds to the bullet list in
	// https://spec.commonmark.org/0.31.2/#html-blocks.
	html1Regexp = regexp.MustCompile(`^ {0,3}<(?i:pre|script|style|textarea)`)
	html1CloserRegexp = regexp.MustCompile(`</(?i:pre|script|style|textarea)`)
	html2Regexp = regexp.MustCompile(`^ {0,3}<!--`)
	html2CloserRegexp = regexp.MustCompile(`-->`)
	html3Regexp = regexp.MustCompile(`^ {0,3}<\?`)
	html3CloserRegexp = regexp.MustCompile(`\?>`)
	html4Regexp = regexp.MustCompile(`^ {0,3}<![a-zA-Z]`)
	html4CloserRegexp = regexp.MustCompile(`>`)
	html5Regexp = regexp.MustCompile(`^ {0,3}<!\[CDATA\[`)
	html5CloserRegexp = regexp.MustCompile(`\]\]>`)

	html6Regexp = regexp.MustCompile(`^ {0,3}</?(?i:address|article|aside|base|basefont|blockquote|body|caption|center|col|colgroup|dd|details|dialog|dir|div|dl|dt|fieldset|figcaption|figure|footer|form|frame|frameset|h1|h2|h3|h4|h5|h6|head|header|hr|html|iframe|legend|li|link|main|menu|menuitem|nav|noframes|ol|optgroup|option|p|param|search|section|summary|table|tbody|td|tfoot|th|thead|title|tr|track|ul)(?:[ \t>]|$|/>)`)
	html7Regexp = regexp.MustCompile(
		fmt.Sprintf(`^ {0,3}(?:%s|%s)[ \t]*$`, openTag, closingTag))

	// https://spec.commonmark.org/0.31.2/#uri-autolink
	uriAutolinkRegexp = regexp.MustCompile(
		`^<` + scheme + `:[^\x00-\x19 <>]*` + `>`)
	// https://spec.commonmark.org/0.31.2/#email-autolink
	emailAutolinkRegexp = regexp.MustCompile(
		`^<[a-zA-Z0-9` + emailLocalPuncts + `]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*>`)

	openTagRegexp = regexp.MustCompile(`^` + openTag)
	closingTagRegexp = regexp.MustCompile(`^` + closingTag)
}

const indentedCodePrefix = "    "

func (p *blockParser) render() {
	for p.lines.more() {
		line, lineNo := p.lines.next()
		line, matchedContainers, newItem := p.tree.processContainerMarkers(line, lineNo, p.codec)

		if isBlankLine(line) {
			// Blank lines terminate blockquote if the continuation marker is
			// absent.
			if i, unmatched := p.tree.unmatchedBlockquote(matchedContainers); unmatched {
				p.tree.closeBlocks(i, lineNo, p.codec)
				continue
			}
			if newItem && p.lines.more() {
				// A list item can start with at most one blank line; the second
				// blank closes it.
				nextLine, _ := p.lines.next()
				nextLine, _ = p.tree.matchContinuationMarkers(nextLine)
				p.lines.backup()
				if isBlankLine(nextLine) {
					p.tree.closeBlocks(len(p.tree.containers)-1, lineNo, p.codec)
				}
			}
			p.tree.closeParagraph(lineNo, p.codec)
		} else if thematicBreakRegexp.MatchString(line) {
			p.tree.closeBlocks(matchedContainers, lineNo, p.codec)
			p.codec.Do(Op{Type: OpThematicBreak, LineNo: lineNo})
		} else if m := atxHeadingRegexp.FindStringSubmatchIndex(line); m != nil {
			p.tree.closeBlocks(matchedContainers, lineNo, p.codec)
			openerStart, openerEnd := m[2], m[3]
			opener := line[openerStart:openerEnd]
			line = strings.TrimRight(line[openerEnd:], " \t")
			if closer := atxHeadingCloserRegexp.FindString(line); closer != "" {
				line = strings.TrimRight(line[:len(line)-len(closer)], " \t")
			}
			attr := ""
			if m := atxHeadingAttributeRegexp.FindStringSubmatch(line); m != nil {
				attr = m[1]
				line = strings.TrimRight(line[:len(line)-len(m[0])], " \t")
			}
			level := len(opener)
			p.codec.Do(Op{
				Type: OpHeading, LineNo: lineNo, Number: level, Info: attr,
				Content: renderInline(strings.Trim(line, " \t"))})
		} else if m := codeFenceRegexp.FindStringSubmatch(line); m != nil {
			p.tree.closeBlocks(matchedContainers, lineNo, p.codec)
			indent, opener, info := len(m[1]), m[2], m[3]
			if opener == "" {
				opener, info = m[4], m[5]
			}
			p.parseFencedCodeBlock(indent, opener, info)
		} else if len(p.tree.paragraph) == 0 && strings.HasPrefix(line, indentedCodePrefix) {
			p.tree.closeBlocks(matchedContainers, lineNo, p.codec)
			p.parseIndentedCodeBlock(line)
		} else if html1Regexp.MatchString(line) {
			p.tree.closeBlocks(matchedContainers, lineNo, p.codec)
			p.parseCloserTerminatedHTMLBlock(line, html1CloserRegexp.MatchString)
		} else if html2Regexp.MatchString(line) {
			p.tree.closeBlocks(matchedContainers, lineNo, p.codec)
			p.parseCloserTerminatedHTMLBlock(line, html2CloserRegexp.MatchString)
		} else if html3Regexp.MatchString(line) {
			p.tree.closeBlocks(matchedContainers, lineNo, p.codec)
			p.parseCloserTerminatedHTMLBlock(line, html3CloserRegexp.MatchString)
		} else if html4Regexp.MatchString(line) {
			p.tree.closeBlocks(matchedContainers, lineNo, p.codec)
			p.parseCloserTerminatedHTMLBlock(line, html4CloserRegexp.MatchString)
		} else if html5Regexp.MatchString(line) {
			p.tree.closeBlocks(matchedContainers, lineNo, p.codec)
			p.parseCloserTerminatedHTMLBlock(line, html5CloserRegexp.MatchString)
		} else if html6Regexp.MatchString(line) || (len(p.tree.paragraph) == 0 && html7Regexp.MatchString(line)) {
			p.tree.closeBlocks(matchedContainers, lineNo, p.codec)
			p.parseBlankLineTerminatedHTMLBlock(line)
		} else {
			if len(p.tree.paragraph) == 0 {
				// This is not lazy continuation, so close all unmatched
				// containers.
				p.tree.closeBlocks(matchedContainers, lineNo, p.codec)
			}
			p.tree.paragraph = append(p.tree.paragraph, line)
		}
	}
	p.tree.closeBlocks(0, p.lines.lastLineNo+1, p.codec)
}

func isBlankLine(line string) bool {
	return strings.Trim(line, " \t") == ""
}

func (p *blockParser) parseFencedCodeBlock(indent int, opener, info string) {
	// Escaped spaces and tabs (e.g. &Tab;) should also be trimmed, so process
	// the info string before trimming.
	info = strings.Trim(processCodeFenceInfo(info), " \t")
	var lines []string
	startLineNo := p.lines.lastLineNo
	doCodeBlock := func() {
		p.codec.Do(Op{Type: OpCodeBlock, LineNo: startLineNo, Info: info, Lines: lines})
	}

	for p.lines.more() {
		line, lineNo := p.lines.next()
		line, matchedContainers := p.tree.matchContinuationMarkers(line)
		if isBlankLine(line) {
			if i, unmatched := p.tree.unmatchedBlockquote(matchedContainers); unmatched {
				doCodeBlock()
				p.tree.closeBlocks(i, lineNo, p.codec)
				return
			}
		} else if matchedContainers < len(p.tree.containers) {
			p.lines.backup()
			doCodeBlock()
			return
		}
		if m := codeFenceCloserRegexp.FindStringSubmatch(line); m != nil {
			closer := m[1]
			if closer[0] == opener[0] && len(closer) >= len(opener) {
				doCodeBlock()
				return
			}
		}
		for i := indent; i > 0 && line != "" && line[0] == ' '; i-- {
			line = line[1:]
		}
		lines = append(lines, line)
	}
	doCodeBlock()
}

// Code fence info strings are mostly verbatim, but support backslash and
// entities. This mirrors part of (*inlineParser).render.
func processCodeFenceInfo(text string) string {
	pos := 0
	var sb strings.Builder
	for pos < len(text) {
		b := text[pos]
		if b == '&' {
			if entity := leadingCharRef(text[pos:]); entity != "" {
				sb.WriteString(UnescapeHTML(entity))
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
	lines := []string{strings.TrimPrefix(line, indentedCodePrefix)}
	startLineNo := p.lines.lastLineNo
	doCodeBlock := func() { p.codec.Do(Op{Type: OpCodeBlock, LineNo: startLineNo, Lines: lines}) }
	var savedBlankLines []string

	for p.lines.more() {
		line, lineNo := p.lines.next()
		line, matchedContainers := p.tree.matchContinuationMarkers(line)
		if isBlankLine(line) {
			if i, unmatched := p.tree.unmatchedBlockquote(matchedContainers); unmatched {
				doCodeBlock()
				p.tree.closeBlocks(i, lineNo, p.codec)
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
		lines = append(lines, savedBlankLines...)
		savedBlankLines = savedBlankLines[:0]
		lines = append(lines, strings.TrimPrefix(line, indentedCodePrefix))
	}
	doCodeBlock()
}

func (p *blockParser) parseCloserTerminatedHTMLBlock(line string, closer func(string) bool) {
	lines := []string{line}
	startLineNo := p.lines.lastLineNo
	doHTMLBlock := func() {
		p.codec.Do(Op{Type: OpHTMLBlock, LineNo: startLineNo, Lines: lines})
	}

	if closer(line) {
		doHTMLBlock()
		return
	}
	var savedBlankLines []string
	for p.lines.more() {
		line, lineNo := p.lines.next()
		line, matchedContainers := p.tree.matchContinuationMarkers(line)
		if isBlankLine(line) {
			if i, unmatched := p.tree.unmatchedBlockquote(matchedContainers); unmatched {
				doHTMLBlock()
				p.tree.closeBlocks(i, lineNo, p.codec)
				return
			}
			savedBlankLines = append(savedBlankLines, line)
			continue
		} else if matchedContainers < len(p.tree.containers) {
			p.lines.backup()
			doHTMLBlock()
			return
		}
		lines = append(lines, savedBlankLines...)
		savedBlankLines = savedBlankLines[:0]
		lines = append(lines, line)
		if closer(line) {
			doHTMLBlock()
			return
		}
	}
	doHTMLBlock()
}

func (p *blockParser) parseBlankLineTerminatedHTMLBlock(line string) {
	lines := []string{line}
	startLineNo := p.lines.lastLineNo
	doHTMLBlock := func() { p.codec.Do(Op{Type: OpHTMLBlock, LineNo: startLineNo, Lines: lines}) }

	for p.lines.more() {
		line, lineNo := p.lines.next()
		line, matchedContainers := p.tree.matchContinuationMarkers(line)
		if isBlankLine(line) {
			doHTMLBlock()
			if i, unmatched := p.tree.unmatchedBlockquote(matchedContainers); unmatched {
				p.tree.closeBlocks(i, lineNo, p.codec)
			}
			return
		} else if matchedContainers < len(p.tree.containers) {
			p.lines.backup()
			break
		}
		lines = append(lines, line)
	}
	doHTMLBlock()
}

// This struct corresponds to the block tree in
// https://spec.commonmark.org/0.31.2/#phase-1-block-structure.
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
func (t *blockTree) processContainerMarkers(line string, lineNo int, codec Codec) (string, int, bool) {
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

	t.closeBlocks(matched, lineNo, codec)
	for _, c := range newContainers {
		if c.typ.isItem() {
			if continueList {
				continueList = false
			} else {
				list := container{typ: c.typ.itemToList(), punct: c.punct, start: c.start}
				t.containers = append(t.containers, list)
				codec.Do(Op{Type: containerOpenOp[list.typ], LineNo: lineNo, Number: list.start})
			}
		}
		t.containers = append(t.containers, c)
		codec.Do(Op{Type: containerOpenOp[c.typ], LineNo: lineNo})
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
	// https://spec.commonmark.org/0.31.2/#block-quotes
	blockquoteMarkerRegexp = regexp.MustCompile(`^ {0,3}> ?`)

	// Rule #1 and #2 of https://spec.commonmark.org/0.31.2/#list-items
	itemStartingMarkerRegexp = regexp.MustCompile(
		// Capture groups:
		// 1. bullet item punctuation
		// 2. ordered item start index
		// 3. ordered item punctuation
		// 4. trailing spaces
		`^ {0,3}(?:([-+*])|([0-9]{1,9})([.)]))( +)`)

	// Rule #3 of https://spec.commonmark.org/0.31.2/#list-items
	itemStartingMarkerBlankLineRegexp = regexp.MustCompile(
		// Capture groups are the same, with group 4 always empty.
		`^ {0,3}(?:([-+*])|([0-9]{1,9})([.)]))[ \t]*()$`)
)

// Parses starting markers of container blocks. Returns the line after removing
// all starting markers and new containers to create.
//
// Blockquotes are simple to parse. Most of the code deals with list items,
// described in https://spec.commonmark.org/0.31.2/#list-items.
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
		// After parsing at least one starting marker, the rest of the line is
		// in a new paragraph. This means that bullet list marker can be
		// terminated by end of line or tab (instead of space), and ordered list
		// marker with number != 1 are allowed.
		newParagraph = true
	}
	return line, containers
}

func (t *blockTree) closeBlocks(keep, lineNo int, codec Codec) {
	t.closeParagraph(lineNo, codec)
	for i := len(t.containers) - 1; i >= keep; i-- {
		codec.Do(Op{Type: containerCloseOp[t.containers[i].typ], LineNo: lineNo})
	}
	t.containers = t.containers[:keep]
}

// lineNo identifies the first line not part of the paragraph.
func (t *blockTree) closeParagraph(lineNo int, codec Codec) {
	if len(t.paragraph) == 0 {
		return
	}
	startLineNo := lineNo - len(t.paragraph)
	text := strings.Trim(strings.Join(t.paragraph, "\n"), " \t")
	t.paragraph = t.paragraph[:0]
	codec.Do(Op{Type: OpParagraph, LineNo: startLineNo, Content: renderInline(text)})
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
	// Line number of the last line returned by next.
	lastLineNo int
}

func (s *lineSplitter) more() bool {
	return s.pos < len(s.text)
}

func (s *lineSplitter) next() (string, int) {
	begin := s.pos
	delta := strings.IndexByte(s.text[begin:], '\n')
	if delta == -1 {
		s.pos = len(s.text)
		s.lastLineNo++
		return s.text[begin:], s.lastLineNo
	}
	s.pos += delta + 1
	s.lastLineNo++
	return s.text[begin : s.pos-1], s.lastLineNo
}

func (s *lineSplitter) backup() {
	if s.pos == 0 {
		return
	}
	s.pos = 1 + strings.LastIndexByte(s.text[:s.pos-1], '\n')
	s.lastLineNo--
}

var leftAnchoredCharRefRegexp = regexp.MustCompile(`^` + charRefPattern)

func leadingCharRef(s string) string {
	return leftAnchoredCharRefRegexp.FindString(s)
}
