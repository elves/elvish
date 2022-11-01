package md_test

import (
	"html"
	"testing"

	"github.com/google/go-cmp/cmp"
	"src.elv.sh/pkg/md"
	"src.elv.sh/pkg/testutil"
)

var fmtCases = []struct {
	Name     string
	Markdown string
}{
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

func TestFmtPreservesHTMLRender(t *testing.T) {
	testutil.Set(t, &md.UnescapeEntities, html.UnescapeString)
	for _, tc := range testCases {
		t.Run(tc.testName(), func(t *testing.T) {
			if tc.Name == "HTML blocks supplemental/Closed by insufficient list item indentation" {
				t.Skip("TODO HTML output has superfluous newline")
			}
			testFmtPreservesHTMLRender(t, tc.Markdown)
		})
	}
	for _, tc := range fmtCases {
		t.Run(tc.Name, func(t *testing.T) {
			testFmtPreservesHTMLRender(t, tc.Markdown)
		})
	}
}

func testFmtPreservesHTMLRender(t *testing.T, original string) {
	formatted := render(original, &md.FmtCodec{})
	formattedRender := render(formatted, &htmlCodec{})
	originalRender := render(original, &htmlCodec{})
	if formattedRender != originalRender {
		t.Errorf("original:\n%s\nformatted:\n%s\n"+
			"HTML diff (-original +formatted):\n%sops diff (-original +formatted):\n%s",
			hr+"\n"+original+hr, hr+"\n"+formatted+hr,
			cmp.Diff(originalRender, formattedRender),
			cmp.Diff(render(original, &md.OpTraceCodec{}), render(formatted, &md.OpTraceCodec{})))
	}
}
