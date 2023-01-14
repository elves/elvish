package md_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"src.elv.sh/pkg/md"
)

var textTests = []struct {
	name       string
	markdown   string
	wantBlocks []md.TextBlock
}{
	{
		name: "code block",
		markdown: dedent(`
			~~~
			foo
			bar
			~~~
			`),
		wantBlocks: []md.TextBlock{
			{Text: "foo\nbar", Code: true},
		},
	},
	{
		name:       "heading",
		markdown:   "# foo",
		wantBlocks: []md.TextBlock{{Text: "foo"}},
	},
	{
		name:       "paragraph",
		markdown:   "foo",
		wantBlocks: []md.TextBlock{{Text: "foo"}},
	},

	{
		name:       "code span",
		markdown:   "foo `bar`",
		wantBlocks: []md.TextBlock{{Text: "foo bar"}},
	},
	{
		name:       "autolink",
		markdown:   "foo <https://example.com>",
		wantBlocks: []md.TextBlock{{Text: "foo https://example.com"}},
	},
	{
		name: "newline",
		markdown: dedent(`
			foo
			bar
			`),
		wantBlocks: []md.TextBlock{{Text: "foo bar"}},
	},
	{
		name: "hard line break",
		markdown: dedent(`
			foo \
			bar
			`),
		wantBlocks: []md.TextBlock{{Text: "foo \n bar"}},
	},
}

func TestText(t *testing.T) {
	for _, tc := range textTests {
		t.Run(tc.name, func(t *testing.T) {
			var codec md.TextCodec
			md.Render(tc.markdown, &codec)
			if diff := cmp.Diff(tc.wantBlocks, codec.Blocks()); diff != "" {
				t.Errorf("blocks (-want +got):\n%s", diff)
			}
		})
	}
}
