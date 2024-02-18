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
		name: "doc:fn instruction",
		text: dedent(`
			# Special function section.
			#doc:fn special
			`),
		wantFns: []Entry{{
			Name:    "special",
			Content: "Special function section.\n",
			LineNo:  1,
		}},
	},

	{
		name: "doc:id instruction",
		text: dedent(`
			# Adds numbers.
			#doc:id add
			fn + {|a b| }
			`),
		wantFns: []Entry{{
			Name:    "+",
			HTMLID:  "add",
			Content: "Adds numbers.\n",
			LineNo:  1,
			Fn:      &Fn{Signature: "a b", Usage: "+ $a $b"},
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
			# Unstable.
			#doc:show-unstable
			fn -foo { }
			`),
		wantFns: []Entry{{
			Name:    "-foo",
			Content: "Unstable.\n",
			LineNo:  1,
			Fn:      &Fn{Usage: "-foo"},
		}},
	},

	{
		name: "empty line breaks comment block",
		text: dedent(`
			# Adds numbers.

			fn add {|a b| }
			`),
		wantFns: []Entry{{
			Name: "add",
			Fn:   &Fn{Signature: "a b", Usage: "add $a $b"},
		}},
	},
	{
		name: "hash with no space breaks comment block",
		text: dedent(`
			# Adds numbers.
			#foo
			fn add {|a b| }
			`),
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
			if diff := cmp.Diff(tc.wantFns, docs.Fns); diff != "" {
				t.Errorf("unexpected Fns:\n%s", diff)
			}
			if diff := cmp.Diff(tc.wantVars, docs.Vars); diff != "" {
				t.Errorf("unexpected Vars:\n%s", diff)
			}
		})
	}
}
