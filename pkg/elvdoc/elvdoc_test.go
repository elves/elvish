package elvdoc

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"src.elv.sh/pkg/testutil"
)

var dedent = testutil.Dedent

var extractTests = []struct {
	name   string
	text   string
	prefix string

	wantFile *FileEntry
	wantFns  []Entry
	wantVars []Entry
}{
	{
		name: "fn with doc comment block",
		text: dedent(`
			# Adds numbers.
			fn add {|a b| }
			`),
		wantFns: []Entry{{
			Name:    "add",
			Content: "Adds numbers.\n",
			LineNo:  1,
			Fn:      &Fn{Signature: "a b", Usage: "add $a $b"},
		}},
	},
	{
		name: "fn with no doc comment",
		text: dedent(`
			fn add {|a b| }
			`),
		wantFns: []Entry{{
			Name: "add",
			Fn:   &Fn{Signature: "a b", Usage: "add $a $b"},
		}},
	},
	{
		name: "fn with options in signature",
		text: dedent(`
			fn add {|a b &k=v| }
			`),
		wantFns: []Entry{{
			Name: "add",
			Fn:   &Fn{Signature: "a b &k=v", Usage: "add $a $b &k=v"},
		}},
	},
	{
		name: "fn with single-quoted name",
		text: `fn 'all''s well' { }`,
		wantFns: []Entry{{
			Name: "all's well",
			Fn:   &Fn{Usage: "'all''s well'"},
		}},
	},
	{
		name: "fn with double-quoted name",
		text: `fn "\\\"" { }`,
		wantFns: []Entry{{
			Name: `\"`,
			Fn:   &Fn{Usage: `'\"'`},
		}},
	},
	{
		name: "fn with quoted string in option value",
		text: `fn add {|&a='| ' &b="\" "| }`,
		wantFns: []Entry{{
			Name: "add",
			Fn:   &Fn{Signature: `&a='| ' &b="\" "`, Usage: `add &a='| ' &b="\" "`},
		}},
	},
	{
		name: "fn with rest argument in signature",
		text: `fn add {|a b @more| }`,
		wantFns: []Entry{{
			Name: "add",
			Fn:   &Fn{Signature: "a b @more", Usage: "add $a $b $more..."},
		}},
	},

	{
		name: "var with doc comment block",
		text: dedent(`
			# Foo.
			var foo
			`),
		wantVars: []Entry{{
			Name:    "$foo",
			Content: "Foo.\n",
			LineNo:  1,
		}},
	},
	{
		name: "var with no doc comment",
		text: dedent(`
			var foo
			`),
		wantVars: []Entry{{
			Name:    "$foo",
			Content: "",
		}},
	},

	{
		name: "prefix impacts both fn and var",
		text: dedent(`
			var v
			fn f { }
			`),
		prefix:   "foo:",
		wantVars: []Entry{{Name: "$foo:v"}},
		wantFns:  []Entry{{Name: "foo:f", Fn: &Fn{Usage: "foo:f"}}},
	},

	{
		name: "directive",
		text: dedent(`
			#foo
			# Adds numbers.
			fn add {|a b| }
			`),
		wantFns: []Entry{{
			Name:       "add",
			Directives: []string{"foo"},
			Content:    "Adds numbers.\n",
			LineNo:     2,
			Fn:         &Fn{Signature: "a b", Usage: "add $a $b"},
		}},
	},

	{
		name: "file-level comment block with no other block",
		text: dedent(`
			#foo
			# This is a module.

			# Foo.
			var foo
			`),
		wantFile: &FileEntry{[]string{"foo"}, "This is a module.\n", 2},
		wantVars: []Entry{{
			Name: "$foo", Content: "Foo.\n", LineNo: 4,
		}},
	},
	{
		name: "file-level comment block with no other block",
		text: dedent(`
			#foo
			# This is a module.
			`),
		wantFile: &FileEntry{[]string{"foo"}, "This is a module.\n", 2},
	},
	{
		name: "no file-level comment",
		text: dedent(`
			use a

			# Foo
			var foo
			`),
		wantVars: []Entry{{
			Name: "$foo", Content: "Foo\n", LineNo: 3,
		}},
	},

	{
		name: "unstable symbol",
		text: dedent(`
			# Unstable.
			fn -foo { }
			`),
		wantFns: nil,
	},
	{
		name: "unstable symbol with doc:show-unstable",
		text: dedent(`
			#doc:show-unstable
			# Unstable.
			fn -foo { }
			`),
		wantFns: []Entry{{
			Name:    "-foo",
			Content: "Unstable.\n",
			LineNo:  2,
			Fn:      &Fn{Usage: "-foo"},
		}},
	},

	{
		name: "empty line breaks comment block",
		text: dedent(`
			# Adds numbers.

			fn add {|a b| }
			`),
		wantFile: &FileEntry{Content: "Adds numbers.\n", LineNo: 1},
		wantFns: []Entry{{
			Name: "add",
			Fn:   &Fn{Signature: "a b", Usage: "add $a $b"},
		}},
	},
	{
		name: "empty comment line does not break comment block",
		text: dedent(`
			# Adds numbers.
			#
			# Supports two numbers.
			fn add {|a b| }
			`),
		wantFns: []Entry{{
			Name: "add",
			Content: dedent(`
						Adds numbers.

						Supports two numbers.
						`),
			LineNo: 1,
			Fn:     &Fn{Signature: "a b", Usage: "add $a $b"},
		}},
	},

	{
		name: "line number tracking",
		text: dedent(`
			# Foo
			# function
			fn foo { }

			# Bar
			# function
			fn bar { }

			# Lorem
			# variable
			var lorem
			`),
		wantFns: []Entry{
			{Name: "foo", Content: "Foo\nfunction\n", LineNo: 1, Fn: &Fn{Usage: "foo"}},
			{Name: "bar", Content: "Bar\nfunction\n", LineNo: 5, Fn: &Fn{Usage: "bar"}},
		},
		wantVars: []Entry{
			{Name: "$lorem", Content: "Lorem\nvariable\n", LineNo: 9},
		},
	},
}

func TestExtract(t *testing.T) {
	for _, tc := range extractTests {
		t.Run(tc.name, func(t *testing.T) {
			docs, err := Extract(strings.NewReader(tc.text), tc.prefix)
			if err != nil {
				t.Errorf("error: %v", err)
			}
			if diff := cmp.Diff(tc.wantFile, docs.File); diff != "" {
				t.Errorf("unexpected File:\n%s", diff)
			}
			if diff := cmp.Diff(tc.wantFns, docs.Fns); diff != "" {
				t.Errorf("unexpected Fns:\n%s", diff)
			}
			if diff := cmp.Diff(tc.wantVars, docs.Vars); diff != "" {
				t.Errorf("unexpected Vars:\n%s", diff)
			}
		})
	}
}
