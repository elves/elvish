package md_test

import (
	"html"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/google/go-cmp/cmp"
	. "src.elv.sh/pkg/md"
	"src.elv.sh/pkg/testutil"
)

var supplementalFmtCases = []testCase{
	{
		Section:  "Fenced code blocks",
		Name:     "Tilde fence with info starting with tilde",
		Markdown: "~~~ ~`\n" + "~~~",
	},
	{
		Section:  "Emphasis and strong emphasis",
		Name:     "Space at start of content",
		Markdown: "*&#32;x*",
	},
	{
		Section:  "Emphasis and strong emphasis",
		Name:     "Space at end of content",
		Markdown: "*x&#32;*",
	},
	{
		Section:  "Emphasis and strong emphasis",
		Name:     "Emphasis opener after word before punctuation",
		Markdown: "&#65;*!*",
	},
	{
		Section:  "Emphasis and strong emphasis",
		Name:     "Emphasis closer after punctuation before word",
		Markdown: "*!*&#65;",
	},
	{
		Section:  "Emphasis and strong emphasis",
		Name:     "Space-only content",
		Markdown: "*&#32;*",
	},
	{
		Section:  "Links",
		Name:     "Exclamation mark before link",
		Markdown: `\![a](b)`,
	},
	{
		Section:  "Links",
		Name:     "Link title with both single and double quotes",
		Markdown: `[a](b ('"))`,
	},
	{
		Section:  "Links",
		Name:     "Link title with fewer double quotes than single quotes and parens",
		Markdown: `[a](b "\"''()")`,
	},
	{
		Section:  "Links",
		Name:     "Link title with fewer single quotes than double quotes and parens",
		Markdown: `[a](b '\'""()')`,
	},
	{
		Section:  "Links",
		Name:     "Link title with fewer parens than single and double quotes",
		Markdown: `[a](b (\(''""))`,
	},
	{
		Section:  "Soft line breaks",
		Name:     "Space at start of line",
		Markdown: "&#32;foo",
	},
	{
		Section:  "Soft line breaks",
		Name:     "Space at end of line",
		Markdown: "foo&#32;",
	},
}

var fmtTestCases = concat(htmlTestCases, supplementalFmtCases)

func TestFmtPreservesHTMLRender(t *testing.T) {
	testutil.Set(t, &UnescapeHTML, html.UnescapeString)
	for _, tc := range fmtTestCases {
		t.Run(tc.testName(), func(t *testing.T) {
			testFmtPreservesHTMLRender(t, tc.Markdown)
		})
	}
}

func TestReflowFmtPreservesHTMLRenderModuleWhitespaces(t *testing.T) {
	testutil.Set(t, &UnescapeHTML, html.UnescapeString)
	for _, tc := range fmtTestCases {
		t.Run(tc.testName(), func(t *testing.T) {
			testReflowFmtPreservesHTMLRenderModuloWhitespaces(t, tc.Markdown, 80)
		})
	}
}

func FuzzFmtPreservesHTMLRender(f *testing.F) {
	for _, tc := range fmtTestCases {
		f.Add(tc.Markdown)
	}
	f.Fuzz(testFmtPreservesHTMLRender)
}

func FuzzReflowFmtPreservesHTMLRenderModuleWhitespaces(f *testing.F) {
	for _, tc := range fmtTestCases {
		f.Add(tc.Markdown, 20)
		f.Add(tc.Markdown, 80)
	}
	f.Fuzz(testReflowFmtPreservesHTMLRenderModuloWhitespaces)
}

func testFmtPreservesHTMLRender(t *testing.T, original string) {
	t.Helper()
	testFmtPreservesHTMLRenderModulo(t, original, 0, nil)
}

var (
	paragraph         = regexp.MustCompile(`(?s)<p>.*?</p>`)
	whitespaceRun     = regexp.MustCompile(`[ \t\n]+`)
	brWithWhitespaces = regexp.MustCompile(`[ \t\n]*<br />[ \t\n]*`)
)

func testReflowFmtPreservesHTMLRenderModuloWhitespaces(t *testing.T, original string, w int) {
	t.Helper()
	testFmtPreservesHTMLRenderModulo(t, original, w, func(html string) string {
		// Coalesce whitespaces in each paragraph.
		return paragraph.ReplaceAllStringFunc(html, func(p string) string {
			body := strings.Trim(p[3:len(p)-4], " \t\n")
			// Convert each whitespace run to a single space.
			body = whitespaceRun.ReplaceAllLiteralString(body, " ")
			// Remove whitespaces around <br />.
			body = brWithWhitespaces.ReplaceAllLiteralString(body, "<br />")
			return "<p>" + body + "</p>"
		})
	})
}

func testFmtPreservesHTMLRenderModulo(t *testing.T, original string, w int, processHTML func(string) string) {
	t.Helper()
	formatted := formatAndSkipIfUnsupported(t, original, w)
	originalRender := RenderString(original, &HTMLCodec{})
	formattedRender := RenderString(formatted, &HTMLCodec{})
	if processHTML != nil {
		originalRender = processHTML(originalRender)
		formattedRender = processHTML(formattedRender)
	}
	if formattedRender != originalRender {
		t.Errorf("original:\n%s\nformatted:\n%s\n"+
			"markdown diff (-original +formatted):\n%s"+
			"HTML diff (-original +formatted):\n%s"+
			"ops diff (-original +formatted):\n%s",
			hr+"\n"+original+hr, hr+"\n"+formatted+hr,
			cmp.Diff(original, formatted),
			cmp.Diff(originalRender, formattedRender),
			cmp.Diff(RenderString(original, &TraceCodec{}), RenderString(formatted, &TraceCodec{})))
	}
}

func formatAndSkipIfUnsupported(t *testing.T, original string, w int) string {
	t.Helper()
	if !utf8.ValidString(original) {
		t.Skipf("input is not valid UTF-8")
	}
	if strings.Contains(original, "\t") {
		t.Skipf("input contains tab")
	}
	codec := &FmtCodec{Width: w}
	formatted := RenderString(original, codec)
	if u := codec.Unsupported(); u != nil {
		t.Skipf("input uses unsupported feature: %v", u)
	}
	return formatted
}
