package modes

import (
	"testing"

	"src.elv.sh/pkg/cli"
	. "src.elv.sh/pkg/cli/clitest"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/diag"
	"src.elv.sh/pkg/ui"
)

func TestCompletion_Filter(t *testing.T) {
	f := setupStartedCompletion(t)
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

func TestCompletion_Accept(t *testing.T) {
	f := setupStartedCompletion(t)
	defer f.Stop()

	f.TTY.Inject(term.K(ui.Enter))
	f.TestTTY(t, "foo", term.DotHere)
}

func TestCompletion_Dismiss(t *testing.T) {
	f := setupStartedCompletion(t)
	defer f.Stop()

	f.App.PopAddon()
	f.App.Redraw()
	f.TestTTY(t /* nothing */)
}

func TestNewCompletion_NoItems(t *testing.T) {
	f := Setup()
	defer f.Stop()
	_, err := NewCompletion(f.App, CompletionSpec{Items: []CompletionItem{}})
	if err != errNoCandidates {
		t.Errorf("should return errNoCandidates")
	}
}

func TestNewCompletion_FocusedWidgetNotCodeArea(t *testing.T) {
	testFocusedWidgetNotCodeArea(t, func(app cli.App) error {
		_, err := NewCompletion(app, CompletionSpec{Items: []CompletionItem{{}}})
		return err
	})
}

func setupStartedCompletion(t *testing.T) *Fixture {
	f := Setup()
	w, _ := NewCompletion(f.App, CompletionSpec{
		Name:    "WORD",
		Replace: diag.Ranging{From: 0, To: 0},
		Items: []CompletionItem{
			{ToShow: ui.T("foo"), ToInsert: "foo"},
			{ToShow: ui.T("foo bar", ui.FgBlue), ToInsert: "'foo bar'"},
		},
	})
	f.App.PushAddon(w)
	f.App.Redraw()
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
