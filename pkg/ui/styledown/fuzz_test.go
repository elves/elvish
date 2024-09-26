package styledown_test

import (
	"reflect"
	"testing"

	"src.elv.sh/pkg/testutil"
	"src.elv.sh/pkg/ui/styledown"
)

func FuzzDerenderIsInverseOfRender(f *testing.F) {
	// We want to test that for any (t, style), Render(Derender(t, style)) == t.
	//
	// However, we can't seed t because it is not a supported type in Go's fuzz
	// framework. Instead, we'll generate t from Styledown source using Render
	// first.
	f.Add(testutil.Dedent(`
		foo
		***
		`), "")
	f.Add(testutil.Dedent(`
		foo
		***

		bar
		###

		foobar
		***###
		`), "")
	f.Add(testutil.Dedent(`
		foo
		   

		no-eol
		`), "")
	f.Add(testutil.Dedent(`
		foo
		RRR

		R red
		`), "R red")
	f.Add(testutil.Dedent(`
		lorem ipsum
		RRRRR GGGGG

		R red
		G green
		`), "R red\nG green")

	f.Fuzz(func(t *testing.T, src string, styleDefs string) {
		text, err := styledown.Render(src)
		if err != nil {
			t.Skip("invalid src")
		}
		newSrc, err := styledown.Derender(text, styleDefs)
		if err != nil {
			t.Skip("probably invalid styleDefs")
		}
		newText, err := styledown.Render(newSrc)
		if textEqual := reflect.DeepEqual(newText, text); !textEqual || err != nil {
			if !textEqual {
				t.Errorf("unexpected newText (see below)")
			}
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			t.Logf("text is:\n%s", text)
			t.Logf("styleDefs is:\n%s", styleDefs)
			t.Logf("newSrc is:\n%s", newSrc)
			t.Logf("newText is:\n%s", newText)
		}
	})
}
