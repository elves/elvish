package md_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	. "src.elv.sh/pkg/md"
	"src.elv.sh/pkg/must"
)

type testCase struct {
	Markdown string `json:"markdown"`
	HTML     string `json:"html"`
	Example  int    `json:"example"`
	Section  string `json:"section"`
	Name     string
}

//go:embed spec.json
var specJSON []byte

var testCases []testCase

var supplementalCases = []testCase{
	{
		Name: "Blockquote supplemental/Increasing level",
		Markdown: `> a
>> b
`,
		HTML: `<blockquote>
<p>a</p>
<blockquote>
<p>b</p>
</blockquote>
</blockquote>
`,
	},
	{
		Name: "Blockquote supplemental/Reducing level",
		Markdown: `>> a
>
> b
`,
		HTML: `<blockquote>
<blockquote>
<p>a</p>
</blockquote>
<p>b</p>
</blockquote>
`,
	},
	{
		Name:     "Code fence supplemental/Empty line in list item",
		Markdown: "- ```\n  a\n\n  ```\n",
		HTML: `<ul>
<li>
<pre><code>a

</code></pre>
</li>
</ul>
`,
	},
	{
		Name:     "List items supplemental/Two leading empty lines with spaces",
		Markdown: "- \n \na",
		HTML: `<ul>
<li></li>
</ul>
<p>a</p>
`,
	},
	{
		Name:     "Link supplemental/Backslash and entity in destination",
		Markdown: `[a](\&gt;)` + "\n",
		HTML:     `<p><a href="&amp;gt;">a</a></p>` + "\n",
	},
	{
		Name:     "Link supplemental/Backslash and entity in title",
		Markdown: `[a](b (\&gt;))` + "\n",
		HTML:     `<p><a href="b" title="&amp;gt;">a</a></p>` + "\n",
	},
	{
		Name:     "Autolink supplemental/Entity",
		Markdown: `<http://&gt;>` + "\n",
		HTML:     `<p><a href="http://%3E">http://&gt;</a></p>` + "\n",
	},
}

func init() {
	must.OK(json.Unmarshal(specJSON, &testCases))
	testCases = append(testCases, supplementalCases...)
}

func TestRender(t *testing.T) {
	for _, tc := range testCases {
		name := tc.Name
		if name == "" {
			name = fmt.Sprintf("%s/Example %d", tc.Section, tc.Example)
		}
		t.Run(name, func(t *testing.T) {
			switch tc.Section {
			case "Tabs",
				"Setext headings",
				"Link reference definitions":
				t.Skip("section not supported")
			}
			switch tc.Example {
			case 23, 33, 317,
				// Link
				526, 527, 528, 529, 530, 531, 532, 533, 534, 535, 536, 537, 538, 539, 540, 541, 542, 543, 544, 548, 549, 552, 553, 554, 555, 556, 557, 558, 559, 560, 561, 562, 563, 564, 565, 566, 567, 568, 569, 570, 572, 575, 576,
				// Image
				581, 582, 583, 584, 585, 586, 587, 588, 590, 591, 592:
				t.Skip("link reference definitions not supported")
			case 59, 115, 141, 300:
				t.Skip("setext heading not supported")
			case 294, 296, 307, 318, 319, 320, 321, 323:
				t.Skip("tight list not supported")
			}

			var codec htmlCodec
			Render(tc.Markdown, &codec)
			got := codec.String()
			// Try to hide the difference between tight and loose lists by
			// "loosifying" the output. This only works for tight lists whose
			// items consist of single lines, so more complex cases are still
			// skipped.
			want := loosifyLists(tc.HTML)
			hr := strings.Repeat("═", 40)
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("input:\n%s\ndiff (-want +got):\n%s",
					hr+"\n"+tc.Markdown+hr, diff)
			}
		})
	}
}

var (
	escapeHTML = strings.NewReplacer(
		"&", "&amp;", `"`, "&quot;", "<", "&lt;", ">", "&gt;").Replace
	escapeDest = strings.NewReplacer(
		`"`, "%22", `\`, "%5C", " ", "%20", "`", "%60",
		"[", "%5B", "]", "%5D", "<", "%3C", ">", "%3E",
		"ö", "%C3%B6",
		"ä", "%C3%A4", " ", "%C2%A0").Replace
)

type htmlCodec struct {
	strings.Builder
}

var tags = []string{
	OpThematicBreak:  "<hr />\n",
	OpCodeBlockEnd:   "</code></pre>\n",
	OpParagraphStart: "<p>", OpParagraphEnd: "</p>\n",

	OpBlockquoteStart: "<blockquote>\n", OpBlockquoteEnd: "</blockquote>\n",
	OpListItemStart: "<li>\n", OpListItemEnd: "</li>\n",
	OpBulletListStart: "<ul>\n", OpBulletListEnd: "</ul>\n",
	OpOrderedListEnd: "</ol>\n",

	OpCodeSpanStart: "<code>", OpCodeSpanEnd: "</code>",
	OpEmphasisStart: "<em>", OpEmphasisEnd: "</em>",
	OpStrongEmphasisStart: "<strong>", OpStrongEmphasisEnd: "</strong>",
	OpLinkEnd:       "</a>",
	OpHardLineBreak: "<br />",
}

func (c *htmlCodec) Do(op Op) {
	switch op.Type {
	case OpText:
		c.WriteString(escapeHTML(op.Text))
	case OpRawHTML:
		c.WriteString(op.Text)
	case OpHeadingStart:
		fmt.Fprintf(c, "<h%d>", op.Number)
	case OpHeadingEnd:
		fmt.Fprintf(c, "</h%d>\n", op.Number)
	case OpCodeBlockStart:
		var attrs attrBuilder
		if op.Text != "" {
			language, _, _ := strings.Cut(op.Text, " ")
			attrs.set("class", "language-"+language)
		}
		fmt.Fprintf(c, "<pre><code%s>", &attrs)
	case OpOrderedListStart:
		var attrs attrBuilder
		if op.Number != 1 {
			attrs.set("start", strconv.Itoa(op.Number))
		}
		fmt.Fprintf(c, "<ol%s>\n", &attrs)
	case OpLinkStart:
		var attrs attrBuilder
		attrs.set("href", escapeDest(op.Dest))
		if op.Text != "" {
			attrs.set("title", op.Text)
		}
		fmt.Fprintf(c, "<a%s>", &attrs)
	case OpImage:
		var attrs attrBuilder
		attrs.set("src", escapeDest(op.Dest))
		attrs.set("alt", op.Alt)
		if op.Text != "" {
			attrs.set("title", op.Text)
		}
		fmt.Fprintf(c, "<img%s />", &attrs)
	default:
		c.WriteString(tags[op.Type])
	}
}

type attrBuilder struct{ strings.Builder }

func (a *attrBuilder) set(k, v string) { fmt.Fprintf(a, ` %s="%s"`, k, escapeHTML(v)) }

var looseListItem = regexp.MustCompile(`<li>([^<]+)</li>`)

func loosifyLists(html string) string {
	return strings.ReplaceAll(
		looseListItem.ReplaceAllString(html, "<li>\n<p>$1</p>\n</li>"),
		"<li></li>", "<li>\n</li>")
}
