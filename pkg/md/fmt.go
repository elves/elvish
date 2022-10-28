package md

import (
	"regexp"
	"strconv"
	"strings"
)

// FmtCodec is a codec that formats Markdown in a specific style.
//
// The only supported configuration option is the text width.
type FmtCodec struct {
	pieces []string
	Width  string

	code      bool
	codeStart int
	tildeCode bool

	lastNewlines   int
	containers     []*fmtContainer
	containerStart int

	linkDest  string
	linkTitle string
}

var (
	escapeText = strings.NewReplacer(
		// TODO: Don't escape in-word _
		"[", "\\[", "]", "\\]", "*", "\\*", "_", "\\_", "`", "\\`", "\\", "\\\\",
		// TODO: Don't always escape these
		"!", "\\!", ".", "\\.", "#", "\\#", "-", "\\-",
		"&", "&amp;", "<", "&lt;").Replace
)

var (
	backquoteRunRegexp = regexp.MustCompile("`+")
	tildeRunRegexp     = regexp.MustCompile("~+")
)

func (c *FmtCodec) Do(op Op) {
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
			c.tildeCode = true
		} else {
			c.codeStart = c.write("```")
			c.tildeCode = false
		}
		if c.tildeCode && strings.HasPrefix(op.Text, "~") {
			c.write(" ")
		}
		c.write(op.Text)
		c.write("\n")
		c.code = true
	case OpCodeBlockEnd:
		var delimRune rune
		var runLens map[int]bool
		if c.tildeCode {
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
		c.pushContainer(&fmtContainer{typ: fmtBlockquote, marker: "> "})
	case OpBlockquoteEnd:
		if c.containerStart == len(c.pieces) {
			c.write("\n")
		}
		c.popContainer()
	case OpListItemStart:
		c.ensureNewStanza()
		c.containerStart = len(c.pieces)
		// Set marker to start marker
		if ct := c.peekContainer(); ct.typ == fmtBulletItem {
			ct.marker = "-   "
		} else {
			ct.marker = strconv.Itoa(ct.number) + ". "
			if len(ct.marker) < 4 {
				ct.marker += strings.Repeat(" ", 4-len(ct.marker))
			}
		}
	case OpListItemEnd:
		if c.containerStart == len(c.pieces) {
			c.write("\n")
		}
		c.peekContainer().number++
	case OpBulletListStart:
		c.pushContainer(&fmtContainer{typ: fmtBulletItem})
	case OpBulletListEnd:
		c.popContainer()
	case OpOrderedListStart:
		c.pushContainer(&fmtContainer{typ: fmtOrderedItem, number: op.Number})
	case OpOrderedListEnd:
		c.popContainer()
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
		c.write("*")
	case OpEmphasisEnd:
		c.write("*")
	case OpStrongEmphasisStart:
		c.write("**")
	case OpStrongEmphasisEnd:
		c.write("**")
	case OpLinkStart:
		c.linkDest = op.Dest
		c.linkTitle = op.Text
		c.write("[")
	case OpLinkEnd:
		// TODO: Escape
		c.write("]")
		c.writeLinkTail(c.linkDest, c.linkTitle)
	case OpImage:
		// TODO: Escape
		c.write("![" + op.Alt + "]")
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
	c.write("(" + escapeText(dest))
	if title != "" {
		c.write(" (" + escapeText(title) + ")")
	}
	c.write(")")
}

func (c *FmtCodec) write(s string) int {
	if len(c.pieces) == 0 || c.lastNewlines > 0 {
		for _, container := range c.containers {
			// TODO: Remove trailing spaces on empty lines
			c.appendPiece(container.useMarker())
		}
	}
	i := c.appendPiece(s)
	if s == "\n" {
		c.lastNewlines++
	} else if strings.HasSuffix(s, "\n") {
		c.lastNewlines = 1
	} else {
		c.lastNewlines = 0
	}
	return i
}

func (c *FmtCodec) ensureNewStanza() {
	c.code = false
	if len(c.pieces) == 0 {
		return
	}
	for c.lastNewlines < 2 {
		c.write("\n")
		c.lastNewlines++
	}
}

func (c *FmtCodec) appendPiece(s string) int {
	c.pieces = append(c.pieces, s)
	return len(c.pieces) - 1
}

func (c *FmtCodec) pushContainer(ct *fmtContainer) { c.containers = append(c.containers, ct) }
func (c *FmtCodec) peekContainer() *fmtContainer   { return c.containers[len(c.containers)-1] }
func (c *FmtCodec) popContainer()                  { c.containers = c.containers[:len(c.containers)-1] }

type fmtContainer struct {
	typ    fmtContainerType
	marker string // starter or continuation marker
	number int    // only used when typ == fmtOrderedItem
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
