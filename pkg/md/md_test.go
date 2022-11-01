package md_test

import (
	"fmt"
	"html"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	. "src.elv.sh/pkg/md"
	"src.elv.sh/pkg/testutil"
)

func TestRender(t *testing.T) {
	testutil.Set(t, &UnescapeEntities, html.UnescapeString)
	for _, tc := range testCases {
		t.Run(tc.testName(), func(t *testing.T) {
			tc.skipIfNotSupported(t)
			got := render(tc.Markdown, &htmlCodec{})
			// Try to hide the difference between tight and loose lists by
			// "loosifying" the output. This only works for tight lists whose
			// items consist of single lines, so more complex cases are still
			// skipped.
			want := loosifyLists(tc.HTML)
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("input:\n%s\ndiff (-want +got):\n%s",
					hr+"\n"+tc.Markdown+hr, diff)
			}
		})
	}
}

// There are different ways to escape HTML and URLs. The CommonMark spec does
// not specify any particular way, but the spec tests do assume a certain one.
// The schemes below are chosen to match the spec tests.
var (
	escapeHTML = strings.NewReplacer(
		"&", "&amp;", `"`, "&quot;", "<", "&lt;", ">", "&gt;").Replace
	escapeURL = strings.NewReplacer(
		`"`, "%22", `\`, "%5C", " ", "%20", "`", "%60",
		"[", "%5B", "]", "%5D", "<", "%3C", ">", "%3E",
		"ö", "%C3%B6",
		"ä", "%C3%A4", " ", "%C2%A0").Replace
)

type htmlCodec struct {
	strings.Builder
}

var tags = []string{
	OpThematicBreak: "<hr />\n",

	OpBlockquoteStart: "<blockquote>\n", OpBlockquoteEnd: "</blockquote>\n",
	OpListItemStart: "<li>\n", OpListItemEnd: "</li>\n",
	OpBulletListStart: "<ul>\n", OpBulletListEnd: "</ul>\n",
	OpOrderedListEnd: "</ol>\n",
}

func (c *htmlCodec) Do(op Op) {
	switch op.Type {
	case OpHeading:
		fmt.Fprintf(c, "<h%d>", op.Number)
		for _, inlineOp := range op.Content {
			c.doInline(inlineOp)
		}
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
		for _, inlineOp := range op.Content {
			c.doInline(inlineOp)
		}
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

func (c *htmlCodec) doInline(op InlineOp) {
	switch op.Type {
	case OpText:
		c.WriteString(escapeHTML(op.Text))
	case OpCodeSpan:
		c.WriteString("<code>")
		c.WriteString(escapeHTML(op.Text))
		c.WriteString("</code>")
	case OpRawHTML:
		c.WriteString(op.Text)
	case OpLinkStart:
		var attrs attrBuilder
		attrs.set("href", escapeURL(op.Dest))
		if op.Text != "" {
			attrs.set("title", op.Text)
		}
		fmt.Fprintf(c, "<a%s>", &attrs)
	case OpImage:
		var attrs attrBuilder
		attrs.set("src", escapeURL(op.Dest))
		attrs.set("alt", op.Alt)
		if op.Text != "" {
			attrs.set("title", op.Text)
		}
		fmt.Fprintf(c, "<img%s />", &attrs)
	default:
		c.WriteString(inlineTags[op.Type])
	}
}

type attrBuilder struct{ strings.Builder }

func (a *attrBuilder) set(k, v string) { fmt.Fprintf(a, ` %s="%s"`, k, escapeHTML(v)) }
