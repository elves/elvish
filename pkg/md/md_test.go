package md_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

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
		Name: "Increasing blockquote level",
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
		Name: "Reducing blockquote level",
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
	Paragraph:      TagPair{Start: "<p>", End: "</p>"},
	Blockquote:     TagPair{Start: "<blockquote>", End: "</blockquote>"},
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
	linkRef   = regexp.MustCompile(`(^|\n)\[([^\\\[\]]|\\[\\\[\]])+\]:`)
	listItem  = regexp.MustCompile(`(^|\n)[*-] `)
	codeBlock = regexp.MustCompile("(^|\n)>*(```|~~~|    )")
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
			if listItem.MatchString(tc.Markdown) {
				t.Skipf("List item not supported")
			}
			if strings.HasPrefix(tc.Markdown, "<a ") {
				t.Skipf("HTML block not supported")
			}

			got := Render(tc.Markdown, htmlSyntax)
			if got != tc.HTML {
				t.Errorf("input:\n%sgot:\n%swant:\n%s", tc.Markdown, got, tc.HTML)
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
		"Link reference definitions",
		"Blank lines",
		"List items",
		"Lists":
		return true
	default:
		return false
	}
}

func unsupportedExample(example int) string {
	switch example {
	case 59:
		return "has setext heading"
	default:
		return ""
	}
}
