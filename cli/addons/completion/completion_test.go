package completion

import (
	"testing"

	. "github.com/elves/elvish/cli/apptest"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/diag"
	"github.com/elves/elvish/ui"
)

func setupStarted(t *testing.T) *Fixture {
	f := Setup()
	Start(f.App, Config{
		Name:    "WORD",
		Replace: diag.Ranging{From: 0, To: 0},
		Items: []Item{
			{ToShow: "foo", ToInsert: "foo"},
			{ToShow: "foo bar", ToInsert: "'foo bar'",
				ShowStyle: ui.Style{Foreground: "blue"}},
		},
	})
	f.TestTTY(t,
		"foo\n", Styles,
		"___",
		" COMPLETING WORD  ", Styles,
		"***************** ", term.DotHere, "\n",
		"foo  foo bar", Styles,
		"+++  ///////",
	)
	return f
}

func TestFilter(t *testing.T) {
	f := setupStarted(t)
	defer f.Stop()

	f.TTY.Inject(term.K('b'), term.K('a'))
	f.TestTTY(t,
		"'foo bar'\n", Styles,
		"_________",
		" COMPLETING WORD  ba", Styles,
		"*****************   ", term.DotHere, "\n",
		"foo bar", Styles,
		"#######",
	)
}

func TestAccept(t *testing.T) {
	f := setupStarted(t)
	defer f.Stop()

	f.TTY.Inject(term.K(ui.Enter))
	f.TestTTY(t, "foo", term.DotHere)
}

func TestClose(t *testing.T) {
	f := setupStarted(t)
	defer f.Stop()

	Close(f.App)
	f.TestTTY(t /* nothing */)
}

func TestStart_NoItems(t *testing.T) {
	f := Setup()
	defer f.Stop()
	Start(f.App, Config{Items: []Item{}})
	f.TestTTYNotes(t, "no candidates")
}
