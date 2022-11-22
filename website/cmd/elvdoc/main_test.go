package main

import (
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"src.elv.sh/pkg/testutil"
)

var extractTests = []struct {
	name    string
	src     string
	ns      string
	wantDoc string
}{
	{name: "no doc comment", src: "# not doc comment", wantDoc: ""},

	{
		name: "fn doc comments",
		src: dedent(`
			# B.
			fn b { }

			# A.
			fn a { }
			`),
		wantDoc: dedent(tildeToBackquote(`
			# Functions

			<a name='//apple_ref/cpp/Function/a' class='dashAnchor'></a>

			## a {#a}

			~~~elvish
			a
			~~~

			A.

			<a name='//apple_ref/cpp/Function/b' class='dashAnchor'></a>

			## b {#b}

			~~~elvish
			b
			~~~

			B.
			`)),
	},

	{
		name: "doc-fn and var-fn",
		src: dedent(`
			# A.
			fn a { }

			# B.
			var b
			`),
		wantDoc: dedent(tildeToBackquote(`
			# Variables

			<a name='//apple_ref/cpp/Variable/%24b' class='dashAnchor'></a>

			## $b {#b}

			B.


			# Functions

			<a name='//apple_ref/cpp/Function/a' class='dashAnchor'></a>

			## a {#a}

			~~~elvish
			a
			~~~

			A.
			`)),
	},

	{
		name: "doc:id",
		src: dedent(`
			# Add.
			#doc:id add
			fn + { }
			`),
		wantDoc: dedent(tildeToBackquote(`
			# Functions

			<a name='//apple_ref/cpp/Function/%2B' class='dashAnchor'></a>

			## + {#add}

			~~~elvish
			+
			~~~

			Add.
			`)),
	},
}

func TestExtract(t *testing.T) {
	for _, test := range extractTests {
		t.Run(test.name, func(t *testing.T) {
			r := strings.NewReader(test.src)
			w := new(strings.Builder)
			extract(r, test.ns, w)
			compare(t, w.String(), test.wantDoc)
		})
	}
}

var emptyReader = io.MultiReader()

func TestRun_MultipleFiles(t *testing.T) {
	testutil.InTempDir(t)
	testutil.ApplyDir(testutil.Dir{
		"a.elv": dedent(`
			# Function 2 from a.
			#
			#     Some indented code.
			fn f2 { }
			`),
		"b.elv": dedent(`
			# Function 1 from b.
			fn f1 { }

			# Variable 2 from b.
			var v2
			`),
		"c.elv": dedent(`
			# Variable 1 from c.
			var v1
			`),
		"not-elv.sh": dedent(`
			# This won't appear because it is not in a .go file.
			var wontappear
			`),
		// Subdirectories are ignored with -dir.
		"subpkg": testutil.Dir{
			"a.elv": dedent(`
				# Function f from subpkg/a.
				fn subpkg:f

				# Variable v from subpkg/a.
				var subpkg:v
				`),
		},
	})

	var sb strings.Builder
	run([]string{"a.elv", "b.elv"}, emptyReader, &sb)
	compare(t, sb.String(), dedent(tildeToBackquote(`
		# Variables

		<a name='//apple_ref/cpp/Variable/%24v2' class='dashAnchor'></a>

		## $v2 {#v2}

		Variable 2 from b.


		# Functions

		<a name='//apple_ref/cpp/Function/f1' class='dashAnchor'></a>

		## f1 {#f1}

		~~~elvish
		f1
		~~~

		Function 1 from b.

		<a name='//apple_ref/cpp/Function/f2' class='dashAnchor'></a>

		## f2 {#f2}

		~~~elvish
		f2
		~~~

		Function 2 from a.

		    Some indented code.
		`)))
}

func compare(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("diff (-want+got):\n%s", cmp.Diff(want, got))
	}
}

var (
	dedent           = testutil.Dedent
	tildeToBackquote = strings.NewReplacer("~", "`").Replace
)
