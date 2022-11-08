package md

import (
	"fmt"
	"strconv"
	"strings"
)

// There are different ways to escape HTML and URLs. The CommonMark spec does
// not specify any particular way, but the spec tests do assume a certain one.
// The schemes below are chosen to match the spec tests.
var (
	escapeHTML = strings.NewReplacer(
		"&", "&amp;", `"`, "&quot;", "<", "&lt;", ">", "&gt;",
		// No need to escape single quotes, since attributes in the output
		// always use double quotes.
	).Replace
	escapeURL = strings.NewReplacer(
		`"`, "%22", `\`, "%5C", " ", "%20", "`", "%60",
		"[", "%5B", "]", "%5D", "<", "%3C", ">", "%3E",
		"ö", "%C3%B6",
		"ä", "%C3%A4", " ", "%C2%A0").Replace
)

// HTMLCodec converts markdown to HTML.
type HTMLCodec struct {
	strings.Builder
}

var tags = []string{
	OpThematicBreak: "<hr />\n",

	OpBlockquoteStart: "<blockquote>\n", OpBlockquoteEnd: "</blockquote>\n",
	OpListItemStart: "<li>\n", OpListItemEnd: "</li>\n",
	OpBulletListStart: "<ul>\n", OpBulletListEnd: "</ul>\n",
	OpOrderedListEnd: "</ol>\n",
}

func (c *HTMLCodec) Do(op Op) {
	switch op.Type {
	case OpHeading:
		var attrs attrBuilder
		if op.Info != "" {
			// Only support #id since that's the only thing used in Elvish's
			// Markdown right now. More can be added if needed.
			if op.Info[0] == '#' {
				attrs.set("id", op.Info[1:])
			}
		}
		fmt.Fprintf(c, "<h%d%s>", op.Number, &attrs)
		RenderInlineContentToHTML(&c.Builder, op.Content)
		fmt.Fprintf(c, "</h%d>\n", op.Number)
	case OpCodeBlock:
		var attrs attrBuilder
		if op.Info != "" {
			language, _, _ := strings.Cut(op.Info, " ")
			attrs.set("class", "language-"+language)
		}
		fmt.Fprintf(c, "<pre><code%s>", &attrs)
		for _, line := range op.Lines {
			c.WriteString(escapeHTML(line))
			c.WriteByte('\n')
		}
		c.WriteString("</code></pre>\n")
	case OpHTMLBlock:
		for _, line := range op.Lines {
			c.WriteString(line)
			c.WriteByte('\n')
		}
	case OpParagraph:
		c.WriteString("<p>")
		RenderInlineContentToHTML(&c.Builder, op.Content)
		c.WriteString("</p>\n")
	case OpOrderedListStart:
		var attrs attrBuilder
		if op.Number != 1 {
			attrs.set("start", strconv.Itoa(op.Number))
		}
		fmt.Fprintf(c, "<ol%s>\n", &attrs)
	default:
		c.WriteString(tags[op.Type])
	}
}

var inlineTags = []string{
	OpNewLine:       "\n",
	OpEmphasisStart: "<em>", OpEmphasisEnd: "</em>",
	OpStrongEmphasisStart: "<strong>", OpStrongEmphasisEnd: "</strong>",
	OpLinkEnd:       "</a>",
	OpHardLineBreak: "<br />",
}

// RenderInlineContentToHTML renders inline content to HTML, writing to a
// [strings.Builder].
func RenderInlineContentToHTML(sb *strings.Builder, ops []InlineOp) {
	for _, op := range ops {
		doInline(sb, op)
	}
}

func doInline(sb *strings.Builder, op InlineOp) {
	switch op.Type {
	case OpText:
		sb.WriteString(escapeHTML(op.Text))
	case OpCodeSpan:
		sb.WriteString("<code>")
		sb.WriteString(escapeHTML(op.Text))
		sb.WriteString("</code>")
	case OpRawHTML:
		sb.WriteString(op.Text)
	case OpLinkStart:
		var attrs attrBuilder
		attrs.set("href", escapeURL(op.Dest))
		if op.Text != "" {
			attrs.set("title", op.Text)
		}
		fmt.Fprintf(sb, "<a%s>", &attrs)
	case OpImage:
		var attrs attrBuilder
		attrs.set("src", escapeURL(op.Dest))
		attrs.set("alt", op.Alt)
		if op.Text != "" {
			attrs.set("title", op.Text)
		}
		fmt.Fprintf(sb, "<img%s />", &attrs)
	case OpAutolink:
		var attrs attrBuilder
		attrs.set("href", escapeURL(op.Dest))
		fmt.Fprintf(sb, "<a%s>%s</a>", &attrs, escapeHTML(op.Text))
	default:
		sb.WriteString(inlineTags[op.Type])
	}
}

type attrBuilder struct{ strings.Builder }

func (a *attrBuilder) set(k, v string) { fmt.Fprintf(a, ` %s="%s"`, k, escapeHTML(v)) }
