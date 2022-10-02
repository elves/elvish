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
	ThematicBreak: func(_ string) string { return "<hr />" },
	Heading: func(level int) TagPair {
		tag := "h" + strconv.Itoa(level)
		return TagPair{Start: "<" + tag + ">", End: "</" + tag + ">"}
	},
	Paragraph:  TagPair{Start: "<p>", End: "</p>"},
	Blockquote: TagPair{Start: "<blockquote>", End: "</blockquote>"},
	BulletList: TagPair{Start: "<ul>", End: "</ul>"},
	BulletItem: TagPair{Start: "<li>", End: "</li>"},
	OrderedList: func(start int) TagPair {
		if start == 1 {
			return TagPair{Start: "<ol>", End: "</ol>"}
		}
		return TagPair{
			Start: `<ol start="` + strconv.Itoa(start) + `">`,
			End:   "</ol>"}
	},
	OrderedItem:    TagPair{Start: "<li>", End: "</li>"},
	CodeSpan:       TagPair{Start: "<code>", End: "</code>"},
	Emphasis:       TagPair{Start: "<em>", End: "</em>"},
	StrongEmphasis: TagPair{Start: "<strong>", End: "</strong>"},
	Link: func(dest, title string) TagPair {
		start := ""
		if title == "" {
			start = fmt.Sprintf(`<a href="%s">`, escapeDest(dest))
		} else {
			start = fmt.Sprintf(`<a href="%s" title="%s">`, escapeDest(dest), escapeHTML(title))
		}
		return TagPair{Start: start, End: "</a>"}
	},
	Image: func(dest, alt, title string) string {
		if title == "" {
			return fmt.Sprintf(`<img src="%s" alt="%s" />`, escapeDest(dest), escapeHTML(alt))
		}
		return fmt.Sprintf(`<img src="%s" alt="%s" title="%s" />`, escapeDest(dest), escapeHTML(alt), escapeHTML(title))
	},
	Escape: escapeHTML,
}

var (
	linkRef       = regexp.MustCompile(`(^|\n) {0,3}\[([^\\\[\]]|\\[\\\[\]])+\]:`)
	codeBlock     = regexp.MustCompile("(^|\n)[ >]*(```|~~~|    )")
	emptyListItem = regexp.MustCompile(`(^|\n)([-+*]|[0-9]{1,9}[.)])(\n|$)`)
	htmlBlock     = regexp.MustCompile(`(^|\n)(<a |<!--)`)
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
			if codeBlock.MatchString(tc.Markdown) {
				t.Skipf("Code block not supported")
			}
			if linkRef.MatchString(tc.Markdown) {
				t.Skipf("Link reference not supported")
			}
			if emptyListItem.MatchString(tc.Markdown) {
				t.Skipf("Empty list item not supported")
			}
			if htmlBlock.MatchString(tc.Markdown) {
				t.Skipf("HTML block not supported")
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
		"Precedence",
		"Setext headings",
		"Indented code blocks",
		"Fenced code blocks",
		"HTML blocks",
		"Link reference definitions":
		return true
	default:
		return false
	}
}

func unsupportedExample(example int) string {
	switch example {
	case 59, 300:
		return "has setext heading"
	case 304:
		return "rejection of non-1 starter not implemented yet"
	case 320, 321, 323:
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
