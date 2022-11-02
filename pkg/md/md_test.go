package md_test

import (
	"html"
	"testing"

	"github.com/google/go-cmp/cmp"
	. "src.elv.sh/pkg/md"
	"src.elv.sh/pkg/testutil"
)

func TestRender(t *testing.T) {
	testutil.Set(t, &UnescapeEntities, html.UnescapeString)
	for _, tc := range htmlTestCases {
		t.Run(tc.testName(), func(t *testing.T) {
			tc.skipIfNotSupported(t)
			got := RenderString(tc.Markdown, &HTMLCodec{})
			// Try to hide the difference between tight and loose lists by
			// "loosifying" the output. This only works for tight lists whose
			// items consist of single lines, so more complex cases are still
			// skipped.
			want := loosifyLists(tc.HTML)
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("input:\n%s\ndiff (-want +got):\n%s",
					hr+"\n"+tc.Markdown+hr, diff)
			}
		})
	}
}
