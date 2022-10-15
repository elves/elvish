package md_test

import (
	"html"
	"testing"

	"github.com/google/go-cmp/cmp"
	"src.elv.sh/pkg/md"
	"src.elv.sh/pkg/testutil"
)

func TestFmtPreservesHTMLRender(t *testing.T) {
	testutil.Set(t, &md.UnescapeEntities, html.UnescapeString)
	for _, tc := range testCases {
		t.Run(tc.testName(), func(t *testing.T) {
			tc.skipIfNotSupported(t)
			if tc.Name != "" {
				t.Skip("TODO supplemental cases")
			}
			switch tc.Example {
			case 39, 40:
				t.Skip("TODO escape sequence")
			case 167, 280, 281, 282, 283, 284, 315:
				t.Skip("TODO newline")
			case 301, 302:
				t.Skip("TODO change of list markers")
			case 460, 462:
				t.Skip("TODO mixed double em")
			case 488, 491, 497, 498, 499, 519:
				t.Skip("TODO link and image formatting")
			}
			testFmtPreservesHTMLRender(t, tc.Markdown)
		})
	}
}

func testFmtPreservesHTMLRender(t *testing.T, original string) {
	formatted := render(original, &md.FmtCodec{})
	formattedRender := render(formatted, &htmlCodec{})
	originalRender := loosifyLists(render(original, &htmlCodec{}))
	if formattedRender != originalRender {
		t.Errorf("original:\n%s\nformatted:\n%s\nHTML diff (-original +formatted):\n%s",
			hr+"\n"+original+hr, hr+"\n"+formatted+hr,
			cmp.Diff(originalRender, formattedRender))
	}
}
