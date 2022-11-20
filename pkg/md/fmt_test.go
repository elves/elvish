package md_test

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/google/go-cmp/cmp"
	. "src.elv.sh/pkg/md"
	"src.elv.sh/pkg/testutil"
	"src.elv.sh/pkg/wcwidth"
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
		Section:  "Links",
		Name:     "Newline in link destination",
		Markdown: `[a](<&NewLine;>)`,
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

func FuzzFmtPreservesHTMLRender(f *testing.F) {
	for _, tc := range fmtTestCases {
		f.Add(tc.Markdown)
	}
	f.Fuzz(testFmtPreservesHTMLRender)
}

func testFmtPreservesHTMLRender(t *testing.T, original string) {
	testFmtPreservesHTMLRenderModulo(t, original, 0, nil)
}

func TestReflowFmtPreservesHTMLRenderModuleWhitespaces(t *testing.T) {
	testReflowFmt(t, testReflowFmtPreservesHTMLRenderModuloWhitespaces)
}

func FuzzReflowFmtPreservesHTMLRenderModuleWhitespaces(f *testing.F) {
	fuzzReflowFmt(f, testReflowFmtPreservesHTMLRenderModuloWhitespaces)
}

var (
	paragraph         = regexp.MustCompile(`(?s)<p>.*?</p>`)
	whitespaceRun     = regexp.MustCompile(`[ \t\n]+`)
	brWithWhitespaces = regexp.MustCompile(`[ \t\n]*<br />[ \t\n]*`)
)

func testReflowFmtPreservesHTMLRenderModuloWhitespaces(t *testing.T, original string, w int) {
	if strings.Contains(original, "<p>") {
		t.Skip("markdown contains <p>")
	}
	if strings.Contains(original, "</p>") {
		t.Skip("markdown contains </p>")
	}
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

func TestReflowFmtResultIsUnchangedUnderFmt(t *testing.T) {
	testReflowFmt(t, testReflowFmtResultIsUnchangedUnderFmt)
}

func FuzzReflowFmtResultIsUnchangedUnderFmt(f *testing.F) {
	fuzzReflowFmt(f, testReflowFmtResultIsUnchangedUnderFmt)
}

func testReflowFmtResultIsUnchangedUnderFmt(t *testing.T, original string, w int) {
	reflowed := formatAndSkipIfUnsupported(t, original, w)
	formatted := RenderString(reflowed, &FmtCodec{})
	if reflowed != formatted {
		t.Errorf("original:\n%s\nreflowed:\n%s\nformatted:\n%s"+
			"markdown diff (-reflowed +formatted):\n%s",
			hr+"\n"+original+hr, hr+"\n"+reflowed+hr, hr+"\n"+formatted+hr,
			cmp.Diff(reflowed, formatted))
	}
}

func TestReflowFmtResultFitsInWidth(t *testing.T) {
	testReflowFmt(t, testReflowFmtResultFitsInWidth)
}

func FuzzReflowFmtResultFitsInWidth(f *testing.F) {
	fuzzReflowFmt(f, testReflowFmtResultFitsInWidth)
}

var (
	// Match all markers that can be written by FmtCodec.
	markersRegexp  = regexp.MustCompile(`^ *(?:(?:[-*>]|[0-9]{1,9}[.)]) *)*`)
	linkRegexp     = regexp.MustCompile(`\[.*\]\(.*\)`)
	codeSpanRegexp = regexp.MustCompile("`.*`")
)

func testReflowFmtResultFitsInWidth(t *testing.T, original string, w int) {
	if w <= 0 {
		t.Skip("width <= 0")
	}

	var trace TraceCodec
	Render(original, &trace)
	for _, op := range trace.Ops() {
		switch op.Type {
		case OpHeading, OpCodeBlock, OpHTMLBlock:
			t.Skipf("input contains unsupported block type %s", op.Type)
		}
	}

	reflowed := formatAndSkipIfUnsupported(t, original, w)

	for _, line := range strings.Split(reflowed, "\n") {
		lineWidth := wcwidth.Of(line)
		if lineWidth <= w {
			continue
		}
		// Strip all markers
		content := line[len(markersRegexp.FindString(line)):]
		// Analyze whether the content is allowed to exceed width
		switch {
		case !strings.Contains(content, " "):
		case strings.Contains(content, "<"):
		case linkRegexp.MatchString(content):
		case codeSpanRegexp.MatchString(content):
		default:
			t.Errorf("line length > %d: %q\nfull reflowed:\n%s",
				w, line, hr+"\n"+reflowed+hr)
		}
	}
}

var widths = []int{20, 51, 80}

func testReflowFmt(t *testing.T, test func(*testing.T, string, int)) {
	for _, tc := range fmtTestCases {
		for _, w := range widths {
			t.Run(fmt.Sprintf("%s/Width %d", tc.testName(), w), func(t *testing.T) {
				test(t, tc.Markdown, w)
			})
		}
	}
}

func fuzzReflowFmt(f *testing.F, test func(*testing.T, string, int)) {
	for _, tc := range fmtTestCases {
		for _, w := range widths {
			f.Add(tc.Markdown, w)
		}
	}
	f.Fuzz(test)
}

func testFmtPreservesHTMLRenderModulo(t *testing.T, original string, w int, processHTML func(string) string) {
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
