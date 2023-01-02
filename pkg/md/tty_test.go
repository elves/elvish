package md_test

import (
	"fmt"
	"reflect"
	"testing"

	. "src.elv.sh/pkg/md"
	"src.elv.sh/pkg/ui"
)

var stylesheet = ui.RuneStylesheet{
	'/': ui.Italic, '#': ui.Bold, '^': ui.Inverse, '_': ui.Underlined,
}

var ttyTests = []struct {
	name      string
	markdown  string
	width     int
	highlight func(info, code string) ui.Text
	rellink   func(dest string) string
	ttyRender ui.Text
}{
	// Blocks
	{
		name:      "thematic break",
		markdown:  "---",
		ttyRender: ui.T("────\n"),
	},
	{
		name: "heading",
		markdown: dedent(`
			# h1

			## h2

			content
			`),
		ttyRender: markLines(
			"# h1", stylesheet,
			"####",
			"",
			"## h2", stylesheet,
			"#####",
			"",
			"content",
		),
	},
	{
		name: "code block",
		markdown: dedent(`
			Run this:

			~~~
			echo foo
			~~~
			`),
		ttyRender: ui.T(dedent(`
			Run this:

			  echo foo
			`)),
	},
	{
		name: "HTML block",
		markdown: dedent(`
			foo

			<!-- comment -->

			bar
			`),
		ttyRender: ui.T(dedent(`
			foo

			bar
			`)),
	},
	{
		name: "blockquote",
		markdown: dedent(`
			Quote:

			> foo
			>> lorem
			>
			> bar
			`),
		ttyRender: ui.T(dedent(`
			Quote:

			│ foo
			│
			│ │ lorem
			│
			│ bar
			`)),
	},
	{
		name: "bullet list",
		markdown: dedent(`
			List:

			- one
			   more

			- two
			   more
			`),
		ttyRender: ui.T(dedent(`
			List:

			• one
			  more

			• two
			  more
			`)),
	},
	{
		name: "ordered list",
		markdown: dedent(`
			List:

			1.  one
			      more

			1. two
			   more
			`),
		ttyRender: ui.T(dedent(`
			List:

			1. one
			   more

			2. two
			   more
			`)),
	},
	{
		name: "nested blocks",
		markdown: dedent(`
			> foo
			> - item
			>   1. one
			>   1. another
			> - another item
			`),
		ttyRender: ui.T(dedent(`
			│ foo
			│
			│ • item
			│
			│   1. one
			│
			│   2. another
			│
			│ • another item
			`)),
	},

	// Highlight code block
	{
		name: "highlight",
		markdown: dedent(`
			Some code:

			~~~foo bar
			code content
			~~~
			`),
		highlight: func(info, code string) ui.Text {
			return ui.T(fmt.Sprintf("(%s) %q\n", info, code))
		},
		ttyRender: ui.T(dedent(`
			Some code:

			  (foo bar) "code content\n"
			`)),
	},
	{
		name: "highlight missing trailing newline",
		markdown: dedent(`
			Some code:

			~~~foo bar
			code content
			~~~
			`),
		highlight: func(info, code string) ui.Text {
			return ui.T(fmt.Sprintf("(%s) %q", info, code))
		},
		ttyRender: ui.T(dedent(`
			Some code:

			  (foo bar) "code content\n"
			`)),
	},

	// Inline
	{
		name:      "text",
		markdown:  "foo bar",
		ttyRender: ui.T("foo bar\n"),
	},
	{
		name:     "inline kbd tag",
		markdown: "Press <kbd>Enter</kbd>.",
		ttyRender: markLines(
			"Press Enter.", stylesheet,
			"      ^^^^^ "),
	},
	{
		name:     "code span",
		markdown: "Use `put`.",
		ttyRender: markLines(
			"Use put.", stylesheet,
			"    ___ "),
	},
	{
		name:     "emphasis",
		markdown: "Try *this*.",
		ttyRender: markLines(
			"Try this.", stylesheet,
			"    //// "),
	},
	{
		name:     "strong emphasis",
		markdown: "Try **that**.",
		ttyRender: markLines(
			"Try that.", stylesheet,
			"    #### "),
	},
	{
		name:     "link with absolute destination",
		markdown: "Visit [example](https://example.com).",
		ttyRender: markLines(
			"Visit example (https://example.com).", stylesheet,
			"      _______                       "),
	},
	{
		name:     "link with relative destination",
		markdown: "See [section X](#x) and [page Y](y.html).",
		ttyRender: markLines(
			"See section X and page Y.", stylesheet,
			"    _________     ______ "),
	},
	{
		name:      "image",
		markdown:  "![Example logo](https://example.com/logo.png)",
		ttyRender: ui.T("Image: Example logo (https://example.com/logo.png)\n"),
	},
	{
		name:      "autolink",
		markdown:  "Visit <https://example.com>.",
		ttyRender: ui.T("Visit https://example.com.\n"),
	},
	{
		name: "hard line break",
		markdown: dedent(`
			foo\
			bar
			`),
		ttyRender: ui.T("foo\nbar\n"),
	},

	// ConvertRelativeLink
	{
		name:     "rellink conversion",
		markdown: "See [a](a.html).",
		rellink:  func(dest string) string { return "https://example.com/" + dest },
		ttyRender: markLines(
			"See a (https://example.com/a.html).", stylesheet,
			"    _                              "),
	},
	{
		name:     "rellink conversion return empty string",
		markdown: "See [a](a.html).",
		rellink:  func(dest string) string { return "" },
		ttyRender: markLines(
			"See a.", stylesheet,
			"    _ "),
	},

	// Reflow
	{
		name:     "reflow text",
		markdown: "foo bar lorem ipsum",
		width:    8,
		ttyRender: ui.T(dedent(`
			foo bar
			lorem
			ipsum
			`)),
	},
	{
		name:     "styled text on the same line when reflowing",
		markdown: "*foo bar* lorem ipsum",
		width:    8,
		ttyRender: markLines(
			"foo bar", stylesheet,
			"///////",
			"lorem",
			"ipsum"),
	},
	{
		name:     "styled text broken up when reflowing",
		markdown: "foo bar *lorem ipsum*",
		width:    8,
		ttyRender: markLines(
			"foo bar",
			"lorem", stylesheet,
			"/////",
			"ipsum", stylesheet,
			"/////"),
	},
	{
		name: "multiple lines merged when reflowing",
		markdown: dedent(`
			foo
			bar
			`),
		width:     8,
		ttyRender: ui.T("foo bar\n"),
	},
	{
		name: "hard line break when reflowing",
		markdown: dedent(`
			foo\
			bar
			`),
		width: 8,
		ttyRender: ui.T(dedent(`
			foo
			bar
			`)),
	},
}

func TestTTYCodec(t *testing.T) {
	for _, tc := range ttyTests {
		t.Run(tc.name, func(t *testing.T) {
			codec := TTYCodec{
				Width:               tc.width,
				HighlightCodeBlock:  tc.highlight,
				ConvertRelativeLink: tc.rellink,
			}
			Render(tc.markdown, &codec)
			got := codec.Text()
			if !reflect.DeepEqual(got, tc.ttyRender) {
				t.Errorf("markdown: %s\ngot: %s\nwant:%s",
					hrFence(tc.markdown),
					hrFence(got.VTString()), hrFence(tc.ttyRender.VTString()))
			}
		})
	}
}

func markLines(args ...any) ui.Text {
	// Add newlines to each line.
	//
	// TODO: Change ui.MarkLines to do this.
	for i := 0; i < len(args); i++ {
		switch arg := args[i].(type) {
		case string:
			args[i] = arg + "\n"
		case ui.RuneStylesheet:
			// Skip over the next argument
			i++
		}
	}
	return ui.MarkLines(args...)
}
