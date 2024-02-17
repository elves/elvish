package transcript_test

import (
	"testing"

	"src.elv.sh/pkg/testutil"
	. "src.elv.sh/pkg/transcript"
	"src.elv.sh/pkg/tt"
)

type Dir = testutil.Dir

var (
	It     = tt.It
	Dedent = testutil.Dedent
)

func TestParseSessionsInFS(t *testing.T) {
	tt.Test(t, ParseFromFS,
		// How sessions are discovered, in both .elv and .elvts files.

		It("scans .elv and .elvts files recursively, ignoring other files").
			Args(Dir{
				"d1": Dir{
					"foo.elv": Dedent(`
						# ~~~elvish-transcript
						# ~> echo foo
						# foo
						# ~~~
						fn x {|| }
						`),
					"ignored.txt": "",
				},
				"d2": Dir{
					"bar.elvts": Dedent(`
						~> echo bar
						bar
						`),
				},
				"ignored.go": "package a",
			}).
			Rets([]*Node{
				{Name: "d1/foo.elv/x", Interactions: []Interaction{{Prompt: "~> ", Code: "echo foo", Output: "foo\n"}}},
				{Name: "d2/bar.elvts", Interactions: []Interaction{{Prompt: "~> ", Code: "echo bar", Output: "bar\n"}}},
			}),

		// .elv file-specific handling

		It("extracts all elvish-transcript code blocks from .elv files").
			Args(oneFile("a.elv", `
				# ~~~elvish-transcript
				# ~> f 1
				# 1
				# ~~~
				#
				# ~~~elvish-transcript
				# ~> f 2
				# 2
				# ~~~
				fn f {|| }

				# ~~~elvish-transcript
				# ~> echo $v
				# foo
				# ~~~
				var v
				`)).
			Rets([]*Node{
				{Name: "a.elv/f", Interactions: []Interaction{{"~> ", "f 1", "1\n"}}},
				{Name: "a.elv/f", Interactions: []Interaction{{"~> ", "f 2", "2\n"}}},
				{Name: "a.elv/$v", Interactions: []Interaction{{"~> ", "echo $v", "foo\n"}}},
			}, error(nil)),

		It("uses fields after elvish-transcript in session name in .elv files").
			Args(oneFile("a.elv", `
				# ~~~elvish-transcript title
				# ~> echo foo
				# foo
				# ~~~
				fn x {|| }
				`)).
			Rets(oneNode("a.elv/x/title",
				Interaction{Prompt: "~> ", Code: "echo foo", Output: "foo\n"})),

		It("processes each code block in .elv files like a .elvts file").
			Args(oneFile("a.elv", `
				# ~~~elvish-transcript
				#
				# ~> nop top
				#
				# # h1 #
				#
				# ~> nop h1
				#
				# ## h2 ##
				# ~> nop h2
				fn x { }
				`)).
			Rets([]*Node{
				{
					Name:         "a.elv/x",
					Interactions: []Interaction{{Prompt: "~> ", Code: "nop top"}},
					Children: []*Node{
						{
							Name:         "h1",
							Interactions: []Interaction{{Prompt: "~> ", Code: "nop h1"}},
							Children: []*Node{
								{Name: "h2", Interactions: []Interaction{{Prompt: "~> ", Code: "nop h2"}}},
							},
						},
					},
				},
			}, error(nil)),

		// Session splitting

		It("strips leading and trailing newlines in sessions in .elvts files").
			Args(oneFile("a.elvts", `
				# h1 #


				~> echo foo
				foo


				`)).
			Rets([]*Node{
				{
					Name: "a.elvts",
					Children: []*Node{{
						Name:         "h1",
						Interactions: []Interaction{{Prompt: "~> ", Code: "echo foo", Output: "foo\n"}},
					}},
				},
			}),

		It("organizes nodes into a tree").
			Args(oneFile("a.elvts", `
				~> nop top level
				
				# section 1 #
				~> nop in section 1

				## subsection 1.1 ##
				~> nop in subsection 1.1

				## subsection 1.2 ##
				~> nop in subsection 1.2

				# section 2 #
				~> nop in section 2
				`)).
			Rets([]*Node{
				{
					Name:         "a.elvts",
					Interactions: []Interaction{{"~> ", "nop top level", ""}},
					Children: []*Node{
						{
							Name:         "section 1",
							Interactions: []Interaction{{"~> ", "nop in section 1", ""}},
							Children: []*Node{
								{Name: "subsection 1.1", Interactions: []Interaction{{"~> ", "nop in subsection 1.1", ""}}},
								{Name: "subsection 1.2", Interactions: []Interaction{{"~> ", "nop in subsection 1.2", ""}}},
							},
						},
						{
							Name:         "section 2",
							Interactions: []Interaction{{"~> ", "nop in section 2", ""}},
						},
					},
				},
			}, error(nil)),

		It("ignores comment lines in .elvts files").
			Args(oneFile("a.elvts", `
				// some comment before code
				~> echo foo; echo bar
				// some comments before output
				foo
				// some comments inside output
				bar

				// some comments after output; note that the preceding empty
				// lines is also stripped
				`)).
			Rets(oneNode("a.elvts",
				Interaction{Prompt: "~> ", Code: "echo foo; echo bar", Output: "foo\nbar\n"})),

		It("errors if h2 appears before any h1 in .elvts files").
			Args(oneFile("a.elvts", `
				## h2 ##
				`)).
			Rets([]*Node(nil), errorWithMsg{"a.elvts:1: h2 before h1"}),

		// How a single session is parsed into (REPL) cycles. Most of the code
		// path is shared between .elv and .elvts files, so most of the cases
		// below only test .elvts files.

		It("supports cycles with multi-line code and output").
			Args(oneFile("a.elvts", `
				~> echo foo
				   echo bar
				foo
				bar
				`)).
			Rets(oneNode("a.elvts",
				Interaction{Prompt: "~> ", Code: "echo foo\necho bar", Output: "foo\nbar\n"})),

		It("supports multiple cycles").
			Args(oneFile("a.elvts", `
				~> echo foo
				   echo bar
				foo
				bar
				~> echo lorem
				   echo ipsum
				lorem
				ipsum
				`)).
			Rets(oneNode("a.elvts",
				Interaction{Prompt: "~> ", Code: "echo foo\necho bar", Output: "foo\nbar\n"},
				Interaction{Prompt: "~> ", Code: "echo lorem\necho ipsum", Output: "lorem\nipsum\n"})),

		It("supports cycles with empty output").
			Args(oneFile("a.elvts", `
				~> nop
				~> nop
				`)).
			Rets(oneNode("a.elvts",
				Interaction{Prompt: "~> ", Code: "nop", Output: ""},
				Interaction{Prompt: "~> ", Code: "nop", Output: ""})),

		It("supports more complex prompts").
			Args(oneFile("a.elvts", `
				~/foo> echo foo
				foo
				/opt/bar> echo bar
				bar
				`)).
			Rets(oneNode("a.elvts",
				Interaction{Prompt: "~/foo> ", Code: "echo foo", Output: "foo\n"},
				Interaction{Prompt: "/opt/bar> ", Code: "echo bar", Output: "bar\n"})),

		It("supports directives").
			Args(oneFile("a.elvts", `
				//directive 1

				// some comment and some empty lines

				//directive 2

				~> nop
				`)).
			Rets([]*Node{{
				Name:         "a.elvts",
				Directives:   []string{"directive 1", "directive 2"},
				Interactions: []Interaction{{Prompt: "~> ", Code: "nop"}}}}),

		It("errors when a session in a .elvts file doesn't start with a prompt").
			Args(oneFile("a.elvts", `

				something
				~> echo foo
				foo
				`)).
			Rets([]*Node(nil), errorWithMsg{"a.elvts:2: first non-comment line of a session doesn't have prompt"}),

		It("errors when a session in a fn elvdoc in a .elv file doesn't start with a prompt").
			Args(oneFile("a.elv", `
				# ~~~elvish-transcript
				# something
				# ~~~
				fn x { }
				`)).
			Rets([]*Node(nil), errorWithMsg{"a.elv/x:1: first non-comment line of a session doesn't have prompt"}),

		It("errors when a session in a var elvdoc in a .elv file doesn't start with a prompt").
			Args(oneFile("a.elv", `
				# ~~~elvish-transcript
				# something
				# ~~~
				var x
				`)).
			Rets([]*Node(nil), errorWithMsg{"a.elv/$x:1: first non-comment line of a session doesn't have prompt"}),
	)
}

func oneFile(name, content string) Dir { return Dir{name: Dedent(content)} }

func oneNode(name string, interactions ...Interaction) ([]*Node, error) {
	return []*Node{{Name: name, Interactions: interactions}}, nil
}

type errorWithMsg struct{ msg string }

func (e errorWithMsg) Match(got tt.RetValue) bool {
	if gotErr, ok := got.(error); ok {
		return e.msg == gotErr.Error()
	}
	return false
}
