package completion

import (
	"testing"

	. "github.com/elves/elvish/cli/apptest"
	"github.com/elves/elvish/cli/term"
	"github.com/elves/elvish/diag"
	"github.com/elves/elvish/ui"
)

var styles = ui.RuneStylesheet{
	'-': ui.Underlined,
	'*': ui.Stylings(ui.Bold, ui.LightGray, ui.BgMagenta),
	'#': ui.Inverse,
	'b': ui.Blue,
	'B': ui.Stylings(ui.Inverse, ui.Blue),
}

func setup(t *testing.T) *Fixture {
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
		"foo\n", styles,
		"---",
		"COMPLETING WORD ", styles,
		"*************** ", term.DotHere, "\n",
		"foo  foo bar", styles,
		"###  bbbbbbb",
	)
	return f
}

func TestFilter(t *testing.T) {
	f := setup(t)
	defer f.Stop()

	f.TTY.Inject(term.K('b'), term.K('a'))
	f.TestTTY(t,
		"'foo bar'\n", styles,
		"---------",
		"COMPLETING WORD ba", styles,
		"***************   ", term.DotHere, "\n",
		"foo bar", styles,
		"BBBBBBB",
	)
}

func TestAccept(t *testing.T) {
	f := setup(t)
	defer f.Stop()

	f.TTY.Inject(term.K(ui.Enter))
	f.TestTTY(t, "foo", term.DotHere)
}

func TestClose(t *testing.T) {
	f := setup(t)
	defer f.Stop()

	Close(f.App)
	f.TestTTY(t /* nothing */)
}
