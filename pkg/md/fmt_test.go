package md_test

import (
	"html"
	"testing"

	"github.com/google/go-cmp/cmp"
	. "src.elv.sh/pkg/md"
	"src.elv.sh/pkg/testutil"
)

var supplementalFmtCases = []testCase{
	{
		Name:     "Tilde fence with info starting with tilde",
		Markdown: "~~~ ~`\n" + "~~~",
	},
	{
		Name:     "Space at start of line",
		Markdown: "&#32;foo",
	},
	{
		Name:     "Space at end of line",
		Markdown: "foo&#32;",
	},
	{
		Name:     "Exclamation mark before link",
		Markdown: `\![a](b)`,
	},
	{
		Name:     "Link title with both single and double quotes",
		Markdown: `[a](b ('"))`,
	},
	{
		Name:     "Link title with fewer double quotes than single quotes and parens",
		Markdown: `[a](b "\"''()")`,
	},
	{
		Name:     "Link title with fewer single quotes than double quotes and parens",
		Markdown: `[a](b '\'""()')`,
	},
	{
		Name:     "Link title with fewer parens than single and double quotes",
		Markdown: `[a](b (\(''""))`,
	},
}

var fmtTestCases = concat(htmlTestCases, supplementalFmtCases)

func TestFmtPreservesHTMLRender(t *testing.T) {
	testutil.Set(t, &UnescapeEntities, html.UnescapeString)
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
	codec := &FmtCodec{}
	formatted := RenderString(original, codec)
	if u := codec.Unsupported(); u != nil {
		t.Skipf("input is unsupported: %v", u)
	}
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
	testutil.Set(t, &UnescapeEntities, html.UnescapeString)
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
	t.Helper()
	codec := &FmtCodec{}
	formatted1 := RenderString(original, codec)
	if u := codec.Unsupported(); u != nil {
		t.Skipf("input is unsupported: %v", u)
	}
	formatted2 := RenderString(formatted1, &FmtCodec{})
	if formatted1 != formatted2 {
		t.Errorf("original:\n%s\nformatted1:\n%s\nformatted2:\n%s\n"+
			"diff (-formatted1 +formatted2):\n%s",
			hr+"\n"+original+hr, hr+"\n"+formatted1+hr, hr+"\n"+formatted2+hr,
			cmp.Diff(formatted1, formatted2))
	}
}
