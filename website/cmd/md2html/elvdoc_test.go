package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"src.elv.sh/pkg/elvdoc"
	"src.elv.sh/pkg/testutil"
)

var (
	dedent           = testutil.Dedent
	tildeToBackquote = strings.NewReplacer("~", "`").Replace
)

var writeTests = []struct {
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

			## a

			~~~elvish
			a
			~~~

			A.

			<a name='//apple_ref/cpp/Function/b' class='dashAnchor'></a>

			## b

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

			## $b

			B.


			# Functions

			<a name='//apple_ref/cpp/Function/a' class='dashAnchor'></a>

			## a

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

func TestWriteElvdocSections(t *testing.T) {
	for _, test := range writeTests {
		t.Run(test.name, func(t *testing.T) {
			docs, _ := elvdoc.Extract(strings.NewReader(test.src), test.ns)
			w := new(strings.Builder)
			writeElvdocSections(w, test.ns, docs)
			if diff := cmp.Diff(test.wantDoc, w.String()); diff != "" {
				t.Errorf("diff (-want+got):\n%s", diff)
			}
		})
	}
}
