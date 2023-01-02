package md_test

import (
	"fmt"
	"html"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	. "src.elv.sh/pkg/md"
	"src.elv.sh/pkg/testutil"
)

// The spec contains some tests where non-ASCII characters get escaped in URLs.
var escapeURLAttr = strings.NewReplacer(
	`"`, "%22", `\`, "%5C", " ", "%20", "`", "%60",
	"[", "%5B", "]", "%5D", "<", "%3C", ">", "%3E",
	"ö", "%C3%B6",
	"ä", "%C3%A4", " ", "%C2%A0").Replace

func TestHTML(t *testing.T) {
	testutil.Set(t, &UnescapeHTML, html.UnescapeString)
	testutil.Set(t, EscapeURL, escapeURLAttr)
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

func TestHTML_ConvertCodeBlock(t *testing.T) {
	f := func(info, code string) string {
		return fmt.Sprintf("%s (%q)", info, code)
	}
	c := HTMLCodec{ConvertCodeBlock: f}
	markdown := dedent(`
		~~~elvish foo bar
		echo
		echo
		~~~
		`)
	want := dedent(`
		<pre><code class="language-elvish">elvish foo bar ("echo\necho\n")</code></pre>
		`)
	got := RenderString(markdown, &c)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("diff (-want +got):\n%s", diff)
	}
}
