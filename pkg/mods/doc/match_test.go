package doc_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"src.elv.sh/pkg/mods/doc"
	"src.elv.sh/pkg/testutil"
	"src.elv.sh/pkg/ui"
)

var Dedent = testutil.Dedent

var matchAndShowTests = []struct {
	name     string
	markdown string
	qs       []string
	want     []string
}{
	{
		name:     "no match",
		markdown: "Here is some doc",
		qs:       []string{"foo"},
		want:     nil,
	},
	{
		name:     "not all queries match",
		markdown: "Here is some doc",
		qs:       []string{"some", "foo"},
		want:     nil,
	},

	{
		name:     "one query matches one block",
		markdown: "Here is some doc",
		qs:       []string{"some"},
		want:     []string{"Here is {some} doc"},
	},
	{
		name:     "multiple queries match one block",
		markdown: "Here is some doc",
		qs:       []string{"some", "doc"},
		want:     []string{"Here is {some} {doc}"},
	},
	{
		name:     "multiple queries match one block, reverse order",
		markdown: "Here is some doc",
		qs:       []string{"doc", "some"},
		want:     []string{"Here is {some} {doc}"},
	},
	{
		name: "one query matches multiple blocks, only first matched block shown",
		markdown: Dedent(`
			Here is some doc

			Here is some more doc
			`),
		qs:   []string{"some"},
		want: []string{"Here is {some} doc"},
	},
	{
		name: "multiple queries match multiple blocks respectively",
		markdown: Dedent(`
			Here is some doc

			Here is some more doc
			`),
		qs: []string{"some", "more"},
		want: []string{
			"Here is {some} doc",
			"Here is some {more} doc",
		},
	},
	{
		name:     "overlapping matches",
		markdown: "Here is some doc",
		qs:       []string{"is some", "some doc"},
		want:     []string{"Here {is some doc}"},
	},

	{
		name:     "match in first sentence",
		markdown: "Here is some doc. Here is more. Here is even more.",
		qs:       []string{"some"},
		want:     []string{"Here is {some} doc. …"},
	},
	{
		name:     "match in last sentence",
		markdown: "Here is some doc. Here is more. Here is even more.",
		qs:       []string{"even"},
		want:     []string{"… Here is {even} more."},
	},
	{
		name:     "match in middle sentence",
		markdown: "Here is some doc. Here is more. Here is even more.",
		qs:       []string{"is more"},
		want:     []string{"… Here {is more}. …"},
	},
	{
		name:     "matches in adjacent sentences",
		markdown: "Here is some doc. Here is more. Here is even more.",
		qs:       []string{"some", "more"},
		want:     []string{"Here is {some} doc. Here is {more}. …"},
	},
	{
		name:     "matches in non-adjacent sentences",
		markdown: "Here is some doc. Here is more. Here is even more.",
		qs:       []string{"some", "even"},
		want:     []string{"Here is {some} doc. … Here is {even} more."},
	},
	{
		name:     "multiple matches in one sentence",
		markdown: "Here is some doc. Here is more.",
		qs:       []string{"some", "doc"},
		want:     []string{"Here is {some} {doc}. …"},
	},
	{
		name:     "one match spanning multiple sentences",
		markdown: "Here is some doc. Here is more. Here is even more.",
		qs:       []string{"some doc. Here", "even"},
		want:     []string{"Here is {some doc. Here} is more. Here is {even} more."},
	},

	{
		name:     "match in first line of code block",
		markdown: codeBlockWithLines("echo foo", "echo bar", "echo lorem"),
		qs:       []string{"foo"},
		want:     []string{"echo {foo} …"},
	},
	{
		name:     "match in last line of code block",
		markdown: codeBlockWithLines("echo foo", "echo bar", "echo lorem"),
		qs:       []string{"lorem"},
		want:     []string{"… echo {lorem}"},
	},
	{
		name:     "match in middle line of code block",
		markdown: codeBlockWithLines("echo foo", "echo bar", "echo lorem"),
		qs:       []string{"bar"},
		want:     []string{"… echo {bar} …"},
	},
	{
		name:     "matches in adjacent lines",
		markdown: codeBlockWithLines("echo foo", "echo bar", "echo lorem"),
		qs:       []string{"foo", "bar"},
		want:     []string{"echo {foo} … echo {bar} …"},
	},
	{
		name:     "matches in non-adjacent lines",
		markdown: codeBlockWithLines("echo foo", "echo bar", "echo lorem"),
		qs:       []string{"foo", "lorem"},
		want:     []string{"echo {foo} … echo {lorem}"},
	},
	{
		name:     "multiple matches in one line",
		markdown: codeBlockWithLines("echo foo", "echo bar", "echo lorem"),
		qs:       []string{"ch", "fo"},
		want:     []string{"e{ch}o {fo}o …"},
	},
	{
		name:     "one match spanning multiple lines",
		markdown: codeBlockWithLines("echo foo", "echo bar", "echo lorem"),
		qs:       []string{"foo\necho"},
		want:     []string{"echo {foo … echo} bar …"},
	},
}

func codeBlockWithLines(lines ...string) string {
	return "```\n" + strings.Join(lines, "\n") + "\n```"
}

// Test match and matchedBlock.Show together. They are used together in actual
// production code and easier to test together.
func TestMatchAndShow(t *testing.T) {
	for _, tc := range matchAndShowTests {
		t.Run(tc.name, func(t *testing.T) {
			want := tc.want
			for i := range want {
				want[i] = highlightBraced(want[i])
			}
			got := matchAndShow(tc.markdown, tc.qs)
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}

func matchAndShow(markdown string, qs []string) []string {
	bs, ok := doc.Match(markdown, qs)
	if !ok {
		return nil
	}
	shown := make([]string, len(bs))
	for i, b := range bs {
		shown[i] = b.Show()
	}
	return shown
}

var braced = regexp.MustCompile(`\{.*?\}`)

func highlightBraced(s string) string {
	return braced.ReplaceAllStringFunc(s, func(p string) string {
		return ui.T(p[1:len(p)-1], ui.Bold, ui.FgRed).VTString()
	})
}
