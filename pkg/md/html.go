package md

import (
	"fmt"
	"strconv"
	"strings"
)

var (
	escapeHTML = strings.NewReplacer(
		"&", "&amp;", `"`, "&quot;", "<", "&lt;", ">", "&gt;",
		// No need to escape single quotes, since attributes in the output
		// always use double quotes.
	).Replace
	// Modern browsers will happily accept almost anything in a URL attribute,
	// except for the quote used by the attribute and space. But we try to be
	// conservative and escape some characters, mostly following
	// https://url.spec.whatwg.org/#url-code-points.
	//
	// We don't bother escaping control characters as they are unlikely to
	// appear in Markdown text.
	escapeURL = strings.NewReplacer(
		`"`, "%22", `\`, "%5C", " ", "%20", "`", "%60",
		"[", "%5B", "]", "%5D", "<", "%3C", ">", "%3E").Replace
)

// HTMLCodec converts markdown to HTML.
type HTMLCodec struct {
	strings.Builder
	// If non-nil, will be called for each code block. The return value is
	// inserted into the HTML output and should be properly escaped.
	ConvertCodeBlock func(info, code string) string
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
		language := ""
		if op.Info != "" {
			language, _, _ = strings.Cut(op.Info, " ")
			attrs.set("class", "language-"+language)
		}
		fmt.Fprintf(c, "<pre><code%s>", &attrs)
		if c.ConvertCodeBlock != nil {
			c.WriteString(c.ConvertCodeBlock(op.Info, strings.Join(op.Lines, "\n")+"\n"))
		} else {
			for _, line := range op.Lines {
				c.WriteString(escapeHTML(line))
				c.WriteByte('\n')
			}
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
// [strings.Builder]. This is useful for implementing an alternative
// HTML-outputting [Codec].
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
