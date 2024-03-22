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
				{
					"d1/foo.elv", nil, nil,
					[]*Node{{
						"x", nil, nil,
						[]*Node{{"", nil, []Interaction{{"~> ", "echo foo", 2, 3, "foo\n", 3, 4}}, nil, 2, 4}},
						1, 5}},
					1, 5,
				},
				{"d2/bar.elvts", nil, []Interaction{{"~> ", "echo bar", 1, 2, "bar\n", 2, 3}}, nil, 1, 3},
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
				{"a.elv", nil, nil,
					[]*Node{
						{"f", nil, nil, []*Node{
							{"", nil, []Interaction{{"~> ", "f 1", 2, 3, "1\n", 3, 4}}, nil, 2, 4},
							{"", nil, []Interaction{{"~> ", "f 2", 7, 8, "2\n", 8, 9}}, nil, 7, 9},
						}, 1, 10},
						{"$v", nil, nil, []*Node{
							{"", nil, []Interaction{{"~> ", "echo $v", 13, 14, "foo\n", 14, 15}}, nil, 13, 15},
						}, 12, 16},
					},
					1, 16},
			}, error(nil)),

		It("uses fields after elvish-transcript in session name in .elv files").
			Args(oneFile("a.elv", `
				# ~~~elvish-transcript title
				# ~> echo foo
				# foo
				# ~~~
				fn x {|| }
				`)).
			Rets([]*Node{
				{"a.elv", nil, nil, []*Node{
					{"x", nil, nil, []*Node{
						{"title", nil, []Interaction{{"~> ", "echo foo", 2, 3, "foo\n", 3, 4}}, nil, 2, 4},
					}, 1, 5},
				}, 1, 5},
			}),

		It("supports file-level and symbol-level directives").
			Args(oneFile("a.elv", `
				#//file1
				#//file2

				#//symbol1
				#//symbol2
				# ~~~elvish-transcript title
				# ~> echo foo
				# foo
				# ~~~
				fn x {|| }
				`)).
			Rets([]*Node{
				{"a.elv", []string{"file1", "file2"}, nil, []*Node{
					{"x", []string{"symbol1", "symbol2"}, nil, []*Node{
						{"title", nil, []Interaction{{"~> ", "echo foo", 7, 8, "foo\n", 8, 9}}, nil, 7, 9},
					}, 6, 10},
				}, 1, 10},
			}),

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
			Rets([]*Node{{
				"a.elv", nil, nil, []*Node{{
					"x", nil, nil, []*Node{{
						"", nil,
						[]Interaction{{"~> ", "nop top", 3, 4, "", 4, 4}},
						[]*Node{{
							"h1", nil,
							[]Interaction{{"~> ", "nop h1", 7, 8, "", 8, 8}},
							[]*Node{{
								"h2", nil,
								[]Interaction{{"~> ", "nop h2", 10, 11, "", 11, 11}}, nil,
								9, 11,
							}},
							5, 11,
						}},
						2, 11,
					}},
					1, 11,
				}},
				1, 11,
			}}, error(nil)),

		// Session splitting

		It("strips leading and trailing newlines in sessions in .elvts files").
			Args(oneFile("a.elvts", `
				# h1 #


				~> echo foo
				foo


				`)).
			Rets([]*Node{{
				"a.elvts", nil, nil,
				[]*Node{{
					"h1", nil,
					[]Interaction{{"~> ", "echo foo", 4, 5, "foo\n", 5, 6}}, nil,
					1, 8,
				}},
				1, 8,
			}}),

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
			Rets([]*Node{{
				"a.elvts", nil,
				[]Interaction{{"~> ", "nop top level", 1, 2, "", 2, 2}},
				[]*Node{
					{
						"section 1", nil,
						[]Interaction{{"~> ", "nop in section 1", 4, 5, "", 5, 5}},
						[]*Node{
							{"subsection 1.1", nil, []Interaction{{"~> ", "nop in subsection 1.1", 7, 8, "", 8, 8}}, nil, 6, 9},
							{"subsection 1.2", nil, []Interaction{{"~> ", "nop in subsection 1.2", 10, 11, "", 11, 11}}, nil, 9, 12},
						},
						3, 12,
					},
					{
						"section 2", nil,
						[]Interaction{{"~> ", "nop in section 2", 13, 14, "", 14, 14}}, nil,
						12, 14,
					},
				},
				1, 14,
			}}, error(nil)),

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
			Rets([]*Node{{
				"a.elvts", nil,
				[]Interaction{{"~> ", "echo foo; echo bar", 2, 3, "foo\nbar\n", 4, 7}}, nil,
				1, 10,
			}}),

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
			Rets([]*Node{{
				"a.elvts",
				nil,
				[]Interaction{{"~> ", "echo foo\necho bar", 1, 3, "foo\nbar\n", 3, 5}},
				nil,
				1, 5,
			}}),

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
			Rets([]*Node{{
				"a.elvts", nil,
				[]Interaction{
					{"~> ", "echo foo\necho bar", 1, 3, "foo\nbar\n", 3, 5},
					{"~> ", "echo lorem\necho ipsum", 5, 7, "lorem\nipsum\n", 7, 9},
				}, nil,
				1, 9,
			}}),

		It("supports cycles with empty output").
			Args(oneFile("a.elvts", `
				~> nop
				~> nop
				`)).
			Rets([]*Node{{
				"a.elvts", nil,
				[]Interaction{
					{"~> ", "nop", 1, 2, "", 2, 2},
					{"~> ", "nop", 2, 3, "", 3, 3},
				}, nil,
				1, 3,
			}}),

		It("supports more complex prompts").
			Args(oneFile("a.elvts", `
				~/foo> echo foo
				foo
				/opt/bar> echo bar
				bar
				`)).
			Rets([]*Node{{
				"a.elvts", nil,
				[]Interaction{
					{"~/foo> ", "echo foo", 1, 2, "foo\n", 2, 3},
					{"/opt/bar> ", "echo bar", 3, 4, "bar\n", 4, 5},
				}, nil,
				1, 5,
			}}),

		It("supports directives").
			Args(oneFile("a.elvts", `
				//directive 1

				// some comment and some empty lines

				//directive 2

				~> nop
				`)).
			Rets([]*Node{{
				"a.elvts",
				[]string{"directive 1", "directive 2"},
				[]Interaction{{"~> ", "nop", 7, 8, "", 8, 8}}, nil,
				1, 8,
			}}),

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
			Rets([]*Node(nil), errorWithMsg{"a.elv:2: first non-comment line of a session doesn't have prompt"}),

		It("errors when a session in a var elvdoc in a .elv file doesn't start with a prompt").
			Args(oneFile("a.elv", `
				# ~~~elvish-transcript
				# something
				# ~~~
				var x
				`)).
			Rets([]*Node(nil), errorWithMsg{"a.elv:2: first non-comment line of a session doesn't have prompt"}),
	)
}

func oneFile(name, content string) Dir { return Dir{name: Dedent(content)} }

type errorWithMsg struct{ msg string }

func (e errorWithMsg) Match(got tt.RetValue) bool {
	if gotErr, ok := got.(error); ok {
		return e.msg == gotErr.Error()
	}
	return false
}
