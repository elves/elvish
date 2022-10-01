package md_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	. "src.elv.sh/pkg/md"
	"src.elv.sh/pkg/must"
)

//go:embed spec.json
var specJSON []byte

var spec []struct {
	Markdown string `json:"markdown"`
	HTML     string `json:"html"`
	Example  int    `json:"example"`
	Section  string `json:"section"`
}

func init() {
	must.OK(json.Unmarshal(specJSON, &spec))
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
	Paragraph: TagPair{Start: "<p>", End: "</p>"},
	Code:      TagPair{Start: "<code>", End: "</code>"},
	Em:        TagPair{Start: "<em>", End: "</em>"},
	Strong:    TagPair{Start: "<strong>", End: "</strong>"},
	Link: func(dest, title string) (string, string) {
		start := ""
		if title == "" {
			start = fmt.Sprintf(`<a href="%s">`, escapeDest(dest))
		} else {
			start = fmt.Sprintf(`<a href="%s" title="%s">`, escapeDest(dest), escapeHTML(title))
		}
		return start, "</a>"
	},
	Image: func(dest, alt, title string) string {
		if title == "" {
			return fmt.Sprintf(`<img src="%s" alt="%s" />`, escapeDest(dest), escapeHTML(alt))
		}
		return fmt.Sprintf(`<img src="%s" alt="%s" title="%s" />`, escapeDest(dest), escapeHTML(alt), escapeHTML(title))
	},
	Escape: escapeHTML,
}

func TestRender(t *testing.T) {
	for _, tc := range spec {
		t.Run(fmt.Sprintf("%s/%d", tc.Section, tc.Example), func(t *testing.T) {
			if !supportedSection(tc.Section) {
				t.Skipf("Section %q not supported", tc.Section)
			}
			if strings.HasPrefix(tc.Markdown, "#") {
				t.Skipf("Header not supported")
			}
			if strings.HasPrefix(tc.Markdown, "```") || strings.HasPrefix(tc.Markdown, "~~~") || strings.HasPrefix(tc.Markdown, "    ") {
				t.Skipf("Code block not supported")
			}
			if strings.Contains(tc.Markdown, "\n\n") {
				t.Skipf("Multiple blocks not supported")
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

func supportedSection(section string) bool {
	switch section {
	case "Tabs",
		"Precedence",
		"Thematic breaks",
		"ATX headings",
		"Setext headings",
		"Indented code blocks",
		"Fenced code blocks",
		"HTML blocks",
		"Link reference definitions",
		"Paragraphs",
		"Blank lines",
		"Block quotes",
		"List items",
		"Lists":
		return false
	default:
		return true
	}
}
