package md

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

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

var htmlSyntax = outSyntax{
	codeStart:   "<code>",
	codeEnd:     "</code>",
	emStart:     "<em>",
	emEnd:       "</em>",
	strongStart: "<strong>",
	strongEnd:   "</strong>",
	link: func(dest, title string) (string, string) {
		start := ""
		if title == "" {
			start = fmt.Sprintf(`<a href="%s">`, escapeDest(dest))
		} else {
			start = fmt.Sprintf(`<a href="%s" title="%s">`, escapeDest(dest), escapeHTML(title))
		}
		return start, "</a>"
	},
	image: func(dest, alt, title string) string {
		if title == "" {
			return fmt.Sprintf(`<img src="%s" alt="%s" />`, escapeDest(dest), escapeHTML(alt))
		}
		return fmt.Sprintf(`<img src="%s" alt="%s" title="%s" />`, escapeDest(dest), escapeHTML(alt), escapeHTML(title))
	},
	escape: escapeHTML,
}

func TestConvertInline(t *testing.T) {
	for _, tc := range spec {
		t.Run(fmt.Sprintf("%s/%d", tc.Section, tc.Example), func(t *testing.T) {
			if !supportedSection(tc.Section) {
				t.Skipf("Section %q not supported", tc.Section)
			}
			if strings.HasPrefix(tc.Markdown, "#") {
				t.Skipf("Header not supported")
			}
			if strings.HasPrefix(tc.Markdown, "```") || strings.HasPrefix(tc.Markdown, "    ") {
				t.Skipf("Code block not supported")
			}
			if strings.Contains(tc.Markdown, "\n\n") {
				t.Skipf("Multiple blocks not supported")
			}

			want := strings.TrimSuffix(strings.TrimPrefix(
				strings.TrimRight(tc.HTML, "\n"), "<p>"), "</p>")
			got := strings.TrimRight(renderInline(tc.Markdown, htmlSyntax), "\n")
			if want != got {
				t.Errorf("input:\n%swant:\n%s\ngot:\n%s", tc.Markdown, want, got)
			}
		})
	}
}

func supportedSection(section string) bool {
	switch section {
	case "Entity and numeric character references",
		"Inlines",
		"Code spans",
		"Emphasis and strong emphasis",
		"Links",
		"Autolinks",
		"Images",
		"Raw HTML",
		"Hard line breaks",
		"Soft line breaks",
		"Textual content":
		return true
	default:
		return false
	}
}
