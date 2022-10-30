package md

import (
	"fmt"
	"regexp"
	"strings"
)

// FmtCodec is a codec that formats Markdown in a specific style.
//
// The only supported configuration option is the text width.
//
// The formatted text uses the following style:
//
//   - Blocks are always separated by a blank line.
//
//   - Thematic breaks always use "***".
//
//   - Code blocks always use the fenced syntax; in other words, indented code
//     blocks are never used.
//
//   - Code fences use backquotes (like "```"), except when the info string
//     contains a backquote, in which case tildes are used (like "~~~").
//
//   - Container blocks never omit their markers; in other words, lazy
//     continuation is never used.
//
//   - Bullet lists use "-" as a marker, except when following immediately after
//     another bullet list that already uses "-", in which case "*" is used.
//
//   - Ordered lists use "X." (X being a number) as a marker, except when
//     following immediately after another ordered list that already uses "X.",
//     in which case "X)" is used.
//
//   - Emphasis uses "*", except when following immediately after another
//     emphasis that already uses "*", in which case "_" is used.
//
//   - Strong emphasis always uses "**".
//
//   - Hard line break always uses an explicit "\".
type FmtCodec struct {
	pieces []string
	Width  string

	// Number of trailing newlines in the currently written text. Used to
	// determine how many additional newlines are needed to start a new block.
	trailingNewlines int

	// Whether a code block/span span is active. Text in either is not escaped.
	code bool
	// The index of the piece that starts the current code block/span. Used to
	// determine the content of the code block/span, which can be used to fix
	// the starter and terminator if necessary.
	codeStart int

	// The index of the piece that starts the last emphasis. Used to alternate
	// between emphasis markers if one follows immediately after another.
	emphasisStart int
	// Stack of emphasis markers.
	emphasisMarkers stack[rune]

	// Current active container blocks.
	containers stack[*fmtContainer]
	// The value of len(pieces) when the last container block was started. Used
	// to determine whether a container is empty, in which case an empty line is
	// needed to preserve the container.
	containerStart int
	// The punctuation of the just popped list container, only populated if the
	// last Op was OpBulletListEnd or OpOrderedListEnd. Used to alternate list
	// punctuation when a list follows directly after another of the same type.
	poppedListPunct rune

	linkDest  string
	linkTitle string
}

var (
	escapeText = strings.NewReplacer(
		"[", `\[`, "]", `\]`, "*", `\*`, "`", "\\`", `\`, `\\`,
		"&", "&amp;", "<", "&lt;",
		// TODO: Don't always escape these
		// Don't escape _ when in-word
		"_", "\\_",
		// Only escape when followed by [
		"!", "\\!",
		// Only escape after a number at the beginning of line
		".", "\\.",
		// Only escape at the beginning of line and followed by space
		"#", "\\#",
		// Only escape at the beginning of line and followed by space
		"-", "\\-",
	).Replace
)

var (
	backquoteRunRegexp = regexp.MustCompile("`+")
	tildeRunRegexp     = regexp.MustCompile("~+")
)

func (c *FmtCodec) Do(op Op) {
	if (op.Type == OpText || op.Type == OpRawHTML) && op.Text == "" {
		return
	}

	var poppedListPunct rune
	defer func() {
		c.poppedListPunct = poppedListPunct
	}()

	switch op.Type {
	case OpText:
		if c.code {
			c.write(op.Text)
		} else {
			c.write(escapeText(op.Text))
		}
	case OpRawHTML:
		// TODO: Ensure stanza for HTML block
		c.write(op.Text)
	case OpThematicBreak:
		c.ensureNewStanza()
		c.write("***\n")
	case OpHeadingStart:
		c.ensureNewStanza()
		c.write(strings.Repeat("#", op.Number) + " ")
	case OpHeadingEnd:
	case OpCodeBlockStart:
		c.ensureNewStanza()
		if strings.ContainsRune(op.Text, '`') {
			c.codeStart = c.write("~~~")
			if strings.HasPrefix(op.Text, "~") {
				c.write(" ")
			}
		} else {
			c.codeStart = c.write("```")
		}
		c.write(op.Text)
		c.write("\n")
		c.code = true
	case OpCodeBlockEnd:
		var delimRune rune
		var runLens map[int]bool
		if c.pieces[c.codeStart][0] == '~' {
			delimRune = '~'
			runLens = matchLens(c.pieces[c.codeStart+1:], tildeRunRegexp)
		} else {
			delimRune = '`'
			runLens = matchLens(c.pieces[c.codeStart+1:], backquoteRunRegexp)
		}
		l := 3
		for x := range runLens {
			if l < x+1 {
				l = x + 1
			}
		}
		delim := strings.Repeat(string(delimRune), l)
		if l != 3 {
			c.pieces[c.codeStart] = delim
		}
		c.write(delim)
		c.write("\n")
		c.code = false
	case OpHTMLBlockStart:
		c.ensureNewStanza()
	case OpHTMLBlockEnd:
	case OpParagraphStart:
		c.ensureNewStanza()
	case OpParagraphEnd:
		c.write("\n")
	case OpBlockquoteStart:
		c.ensureNewStanza()
		c.containerStart = len(c.pieces)
		c.containers.push(&fmtContainer{typ: fmtBlockquote, marker: "> "})
	case OpBlockquoteEnd:
		if c.containerStart == len(c.pieces) {
			c.write("\n")
		}
		c.containers.pop()
	case OpListItemStart:
		c.ensureNewStanza()
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
			c.write("\n")
		}
		c.containers.peek().number++
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
	case OpCodeSpanStart:
		// TODO: Handle when content has `
		c.codeStart = c.write("`")
		c.code = true
	case OpCodeSpanEnd:
		hasRunWithLen := matchLens(c.pieces[c.codeStart+1:], backquoteRunRegexp)
		l := 1
		for hasRunWithLen[l] {
			l++
		}
		delim := strings.Repeat("`", l)
		if l != 1 {
			c.pieces[c.codeStart] = delim
		}
		first := c.pieces[c.codeStart+1][0]
		lastPiece := c.pieces[len(c.pieces)-1]
		last := lastPiece[len(lastPiece)-1]
		if first == '`' || last == '`' || (first == ' ' && last == ' ' &&
			!(c.codeStart+1 == len(c.pieces)-1 && len(lastPiece) <= 2)) {
			c.pieces[c.codeStart] += " "
			c.write(" ")
		}
		c.write(delim)
		c.code = false
	case OpEmphasisStart:
		marker := '*'
		if c.emphasisStart == len(c.pieces)-1 && len(c.emphasisMarkers) > 0 {
			marker = pickPunct('*', '_', c.emphasisMarkers.peek())
		}
		c.emphasisMarkers.push(marker)
		c.emphasisStart = c.write(string(marker))
	case OpEmphasisEnd:
		c.write(string(c.emphasisMarkers.pop()))
	case OpStrongEmphasisStart:
		c.write("**")
	case OpStrongEmphasisEnd:
		c.write("**")
	case OpLinkStart:
		c.linkDest = op.Dest
		c.linkTitle = op.Text
		c.write("[")
	case OpLinkEnd:
		c.write("]")
		c.writeLinkTail(c.linkDest, c.linkTitle)
	case OpImage:
		c.write("![" + escapeText(op.Alt) + "]")
		c.writeLinkTail(op.Dest, op.Text)
	case OpHardLineBreak:
		c.write("\\")
	}
}

func (c *FmtCodec) String() string { return strings.Join(c.pieces, "") }

func matchLens(pieces []string, pattern *regexp.Regexp) map[int]bool {
	hasRunWithLen := make(map[int]bool)
	for _, piece := range pieces {
		for _, run := range pattern.FindAllString(piece, -1) {
			hasRunWithLen[len(run)] = true
		}
	}
	return hasRunWithLen
}

func (c *FmtCodec) writeLinkTail(dest, title string) {
	c.write("(" + escapeLinkDest(dest))
	if title != "" {
		c.write(" " + wrapAndEscapeLinkTitle(title))
	}
	c.write(")")
}

const asciiControlOrSpaceOrParens = "\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f ()"

func escapeLinkDest(dest string) string {
	if strings.ContainsAny(dest, asciiControlOrSpaceOrParens) {
		return "<" + strings.ReplaceAll(escapeText(dest), ">", "&gt;") + ">"
	}
	return escapeText(dest)
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

func (c *FmtCodec) write(s string) int {
	if len(c.pieces) == 0 || c.trailingNewlines > 0 {
		for _, container := range c.containers {
			// TODO: Remove trailing spaces on empty lines
			c.appendPiece(container.useMarker())
		}
	}
	i := c.appendPiece(s)
	if s == "\n" {
		c.trailingNewlines++
	} else if strings.HasSuffix(s, "\n") {
		c.trailingNewlines = 1
	} else {
		c.trailingNewlines = 0
	}
	return i
}

func (c *FmtCodec) ensureNewStanza() {
	c.code = false
	if len(c.pieces) == 0 {
		return
	}
	for c.trailingNewlines < 2 {
		c.write("\n")
		c.trailingNewlines++
	}
}

func (c *FmtCodec) appendPiece(s string) int {
	c.pieces = append(c.pieces, s)
	return len(c.pieces) - 1
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
