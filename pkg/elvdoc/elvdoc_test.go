package elvdoc

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"src.elv.sh/pkg/testutil"
)

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
		wantFns: []Entry{
			{
				Name: "add",
				Content: dedent(tildeToBackquote(`
						~~~elvish
						add $a $b
						~~~

						Adds numbers.
						`)),
			},
		},
	},
	{
		name: "fn with no doc comment",
		text: dedent(`
			fn add {|a b| }
			`),
		wantFns: []Entry{
			{
				Name: "add",
				Content: dedent(tildeToBackquote(`
						~~~elvish
						add $a $b
						~~~
						`)),
			},
		},
	},
	{
		name: "fn with options",
		text: dedent(`
			fn add {|a b &k=v| }
			`),
		wantFns: []Entry{
			{
				Name: "add",
				Content: dedent(tildeToBackquote(`
						~~~elvish
						add $a $b &k=v
						~~~
						`)),
			},
		},
	},
	{
		name: "option with space",
		text: dedent(`
			fn add {|a b &k=' '| }
			`),
		wantFns: []Entry{
			{
				Name: "add",
				Content: dedent(tildeToBackquote(`
						~~~elvish
						add $a $b &k=' '
						~~~
						`)),
			},
		},
	},
	{
		name: "fn with rest argument",
		text: dedent(`
			fn add {|a b @more| }
			`),
		wantFns: []Entry{
			{
				Name: "add",
				Content: dedent(tildeToBackquote(`
						~~~elvish
						add $a $b $more...
						~~~
						`)),
			},
		},
	},

	{
		name: "var with doc comment block",
		text: dedent(`
			# Foo.
			var foo
			`),
		wantVars: []Entry{
			{
				Name:    "foo",
				Content: "Foo.\n",
			},
		},
	},
	{
		name: "var with no doc comment",
		text: dedent(`
			var foo
			`),
		wantVars: []Entry{
			{
				Name:    "foo",
				Content: "",
			},
		},
	},

	{
		name: "doc:fn instruction",
		text: dedent(`
			# Special function section.
			#doc:fn special
			`),
		wantFns: []Entry{
			{
				Name:    "special",
				Content: "Special function section.\n",
			},
		},
	},

	{
		name: "doc:id instruction",
		text: dedent(`
			# Adds numbers.
			#doc:id add
			fn + {|a b| }
			`),
		wantFns: []Entry{
			{
				Name: "+",
				ID:   "add",
				Content: dedent(tildeToBackquote(`
						~~~elvish
						+ $a $b
						~~~

						Adds numbers.
						`)),
			},
		},
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
		wantFns: []Entry{
			{
				Name: "-foo",
				Content: dedent(tildeToBackquote(`
						~~~elvish
						-foo
						~~~

						Unstable.
						`)),
			},
		},
	},

	{
		name: "empty line breaks comment block",
		text: dedent(`
			# Adds numbers.

			fn add {|a b| }
			`),
		wantFns: []Entry{
			{
				Name: "add",
				Content: dedent(tildeToBackquote(`
						~~~elvish
						add $a $b
						~~~
						`)),
			},
		},
	},
	{
		name: "hash with no space breaks comment block",
		text: dedent(`
			# Adds numbers.
			#foo
			fn add {|a b| }
			`),
		wantFns: []Entry{
			{
				Name: "add",
				Content: dedent(tildeToBackquote(`
						~~~elvish
						add $a $b
						~~~
						`)),
			},
		},
	},
	{
		name: "empty comment line does not break comment block",
		text: dedent(`
			# Adds numbers.
			#
			# Supports two numbers.
			fn add {|a b| }
			`),
		wantFns: []Entry{
			{
				Name: "add",
				Content: dedent(tildeToBackquote(`
						~~~elvish
						add $a $b
						~~~

						Adds numbers.

						Supports two numbers.
						`)),
			},
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

var (
	dedent           = testutil.Dedent
	tildeToBackquote = strings.NewReplacer("~", "`").Replace
)
