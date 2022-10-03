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
		"[", "%5B", "]", "%5D",
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
			attrs = append(attrs, "title", escapeHTML(title))
		}
		return htmlTagPair("a", attrs...)
	},
	Image: func(dest, alt, title string) string {
		attrs := []string{"src", escapeDest(dest), "alt", escapeHTML(alt)}
		if title != "" {
			attrs = append(attrs, "title", escapeHTML(title))
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
		fmt.Fprintf(&sb, ` %s="%s"`, attrPairs[i], attrPairs[i+1])
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

var (
	linkRef           = regexp.MustCompile(`(^|\n) {0,3}\[([^\\\[\]]|\\[\\\[\]])+\]:`)
	indentedCodeBlock = regexp.MustCompile("(^|\n)[ >]*(    )")
)

func TestRender(t *testing.T) {
	for _, tc := range testCases {
		name := tc.Name
		if name == "" {
			name = fmt.Sprintf("%s/%d", tc.Section, tc.Example)
		}
		t.Run(name, func(t *testing.T) {
			if unsupportedSection(tc.Section) {
				t.Skipf("Section %q not supported", tc.Section)
			}
			if reason := unsupportedExample(tc.Example); reason != "" {
				t.Skipf("Example %d not supported: %s", tc.Example, reason)
			}
			if indentedCodeBlock.MatchString(tc.Markdown) {
				t.Skipf("Indented code block not supported")
			}
			if linkRef.MatchString(tc.Markdown) {
				t.Skipf("Link reference not supported")
			}

			got := Render(tc.Markdown, htmlSyntax)
			want := loosifyLists(tc.HTML)
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("input:\n%sdiff (-want +got):\n%s", tc.Markdown, diff)
			}
		})
	}
}

func unsupportedSection(section string) bool {
	switch section {
	case "Tabs",
		"Setext headings",
		"Indented code blocks",
		"Link reference definitions":
		return true
	default:
		return false
	}
}

func unsupportedExample(example int) string {
	switch example {
	case 59, 141, 300:
		return "setext heading not supported"
	case 318, 320, 321, 323:
		return "tight list not implemented"
	default:
		return ""
	}
}

var looseListItem = regexp.MustCompile(`<li>([^<]+)</li>`)

func loosifyLists(html string) string {
	return strings.ReplaceAll(
		looseListItem.ReplaceAllString(html, "<li>\n<p>$1</p>\n</li>"),
		"<li></li>", "<li>\n</li>")
}
