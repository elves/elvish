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

var additionalCases = []testCase{
	{
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
		Name: "Blockquote supplemental/Increasing level",
	},
	{
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
		Name: "Blockquote supplemental/Reducing level",
	},
	{
		Markdown: "- ```\n  a\n\n  ```\n",
		HTML: `<ul>
<li>
<pre><code>a

</code></pre>
</li>
</ul>
`,
		Name: "Code fence supplemental/Empty line in list item",
	},
	{
		Name:     "List items supplemental/Two leading empty lines with spaces",
		Markdown: "- \n \na", HTML: `<ul>
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
	testCases = append(testCases, additionalCases...)
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

var htmlSyntax = OutputSyntax{
	ThematicBreak: func(_ string) string { return htmlSelfCloseTag("hr") },
	Heading: func(level int) TagPair {
		return htmlTagPair("h" + strconv.Itoa(level))
	},
	CodeBlock: func(info string) TagPair {
		var codeAttrs []string
		if info != "" {
			language, _, _ := strings.Cut(info, " ")
			codeAttrs = []string{"class", "language-" + language}
		}
		return combineHTMLTagPairs(
			htmlTagPair("pre"), htmlTagPair("code", codeAttrs...))
	},
	Paragraph:  htmlTagPair("p"),
	Blockquote: htmlTagPair("blockquote"),
	BulletList: htmlTagPair("ul"),
	BulletItem: htmlTagPair("li"),
	OrderedList: func(start int) TagPair {
		var attrs []string
		if start != 1 {
			attrs = []string{"start", strconv.Itoa(start)}
		}
		return htmlTagPair("ol", attrs...)
	},
	OrderedItem:    htmlTagPair("li"),
	CodeSpan:       htmlTagPair("code"),
	Emphasis:       htmlTagPair("em"),
	StrongEmphasis: htmlTagPair("strong"),
	Link: func(dest, title string) TagPair {
		attrs := []string{"href", escapeDest(dest)}
		if title != "" {
			attrs = append(attrs, "title", title)
		}
		return htmlTagPair("a", attrs...)
	},
	Image: func(dest, alt, title string) string {
		attrs := []string{"src", escapeDest(dest), "alt", alt}
		if title != "" {
			attrs = append(attrs, "title", title)
		}
		return htmlSelfCloseTag("img", attrs...)
	},
	Escape: escapeHTML,
}

func htmlSelfCloseTag(name string, attrPairs ...string) string {
	return "<" + name + concatAttrPairs(attrPairs) + " />"
}

func htmlTagPair(name string, attrPairs ...string) TagPair {
	return TagPair{
		Start: "<" + name + concatAttrPairs(attrPairs) + ">",
		End:   "</" + name + ">"}
}

func concatAttrPairs(attrPairs []string) string {
	var sb strings.Builder
	for i := 0; i+1 < len(attrPairs); i += 2 {
		fmt.Fprintf(&sb, ` %s="%s"`, attrPairs[i], escapeHTML(attrPairs[i+1]))
	}
	return sb.String()
}

func combineHTMLTagPairs(p TagPair, more ...TagPair) TagPair {
	for _, q := range more {
		p.Start += q.Start
		p.End = q.End + p.End
	}
	return p
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

			got := Render(tc.Markdown, htmlSyntax)
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

var looseListItem = regexp.MustCompile(`<li>([^<]+)</li>`)

func loosifyLists(html string) string {
	return strings.ReplaceAll(
		looseListItem.ReplaceAllString(html, "<li>\n<p>$1</p>\n</li>"),
		"<li></li>", "<li>\n</li>")
}
