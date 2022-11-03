package md_test

import (
	"html"
	"strings"
	"testing"

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

func FuzzFmtPreservesHTMLRender(f *testing.F) {
	for _, tc := range fmtTestCases {
		f.Add(tc.Markdown)
	}
	f.Fuzz(testFmtPreservesHTMLRender)
}

func testFmtPreservesHTMLRender(t *testing.T, original string) {
	t.Helper()
	formatted := formatAndSkipIfUnsupported(t, original)
	formattedRender := RenderString(formatted, &HTMLCodec{})
	originalRender := RenderString(original, &HTMLCodec{})
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

func TestFmtIsIdempotent(t *testing.T) {
	testutil.Set(t, &UnescapeHTML, html.UnescapeString)
	for _, tc := range fmtTestCases {
		t.Run(tc.testName(), func(t *testing.T) {
			testFmtIsIdempotent(t, tc.Markdown)
		})
	}
}

func FuzzFmtIsIdempotent(f *testing.F) {
	for _, tc := range fmtTestCases {
		f.Add(tc.Markdown)
	}
	f.Fuzz(testFmtIsIdempotent)
}

func testFmtIsIdempotent(t *testing.T, original string) {
	formatted1 := formatAndSkipIfUnsupported(t, original)
	formatted2 := RenderString(formatted1, &FmtCodec{})
	if formatted1 != formatted2 {
		t.Errorf("original:\n%s\nformatted1:\n%s\nformatted2:\n%s\n"+
			"diff (-formatted1 +formatted2):\n%s",
			hr+"\n"+original+hr, hr+"\n"+formatted1+hr, hr+"\n"+formatted2+hr,
			cmp.Diff(formatted1, formatted2))
	}
}

func formatAndSkipIfUnsupported(t *testing.T, original string) string {
	t.Helper()
	if strings.Contains(original, "\t") {
		t.Skipf("input contains tab")
	}
	codec := &FmtCodec{}
	formatted := RenderString(original, codec)
	if u := codec.Unsupported(); u != nil {
		t.Skipf("input uses unsupported feature: %v", u)
	}
	return formatted
}
