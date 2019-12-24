package listing

import (
	"testing"

	. "github.com/elves/elvish/pkg/cli/apptest"
	"github.com/elves/elvish/pkg/cli/el/codearea"
	"github.com/elves/elvish/pkg/cli/term"
	"github.com/elves/elvish/pkg/ui"
)

func fooAndGreenBar(string) ([]Item, int) {
	return []Item{{"foo", ui.T("foo")}, {"bar", ui.T("bar", ui.FgGreen)}}, 0
}

func TestBasicUI(t *testing.T) {
	f := Setup()
	defer f.Stop()

	Start(f.App, Config{
		Caption:  " TEST ",
		GetItems: fooAndGreenBar,
	})
	f.TestTTY(t,
		"\n",
		" TEST  ", Styles,
		"****** ", term.DotHere, "\n",
		"foo                                               \n", Styles,
		"++++++++++++++++++++++++++++++++++++++++++++++++++",
		"bar                                               ", Styles,
		"vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv",
	)
}

func TestAccept_ClosingListing(t *testing.T) {
	f := Setup()
	defer f.Stop()

	Start(f.App, Config{
		GetItems: fooAndGreenBar,
		Accept: func(t string) bool {
			f.App.CodeArea().MutateState(func(s *codearea.CodeAreaState) {
				s.Buffer.InsertAtDot(t)
			})
			return false
		},
	})
	// foo will be selected
	f.TTY.Inject(term.K('\n'))
	f.TestTTY(t, "foo", term.DotHere)
}

func TestAccept_NotClosingListing(t *testing.T) {
	f := Setup()
	defer f.Stop()

	Start(f.App, Config{
		GetItems: fooAndGreenBar,
		Accept: func(t string) bool {
			f.App.CodeArea().MutateState(func(s *codearea.CodeAreaState) {
				s.Buffer.InsertAtDot(t)
			})
			return true
		},
	})
	// foo will be selected
	f.TTY.Inject(term.K('\n'))
	f.TestTTY(t,
		"foo\n",
		" LISTING  ", Styles,
		"********* ", term.DotHere, "\n",
		"foo                                               \n", Styles,
		"++++++++++++++++++++++++++++++++++++++++++++++++++",
		"bar                                               ", Styles,
		"vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv",
	)
}

func TestAccept_DefaultNop(t *testing.T) {
	f := Setup()
	defer f.Stop()

	Start(f.App, Config{GetItems: fooAndGreenBar})
	f.TTY.Inject(term.K('\n'))
	f.TestTTY(t /* nothing */)
}

func TestAutoAccept(t *testing.T) {
	f := Setup()
	defer f.Stop()

	Start(f.App, Config{
		GetItems: func(query string) ([]Item, int) {
			if query == "" {
				// Return two items initially.
				return []Item{
					{"foo", ui.T("foo")}, {"bar", ui.T("bar")},
				}, 0
			}
			return []Item{{"bar", ui.T("bar")}}, 0
		},
		Accept: func(t string) bool {
			f.App.CodeArea().MutateState(func(s *codearea.CodeAreaState) {
				s.Buffer.InsertAtDot(t)
			})
			return false
		},
		AutoAccept: true,
	})
	f.TTY.Inject(term.K('a'))
	f.TestTTY(t, "bar", term.DotHere)
}

func TestAbortWhenGetItemsUnspecified(t *testing.T) {
	f := Setup()
	defer f.Stop()

	Start(f.App, Config{})
	f.TestTTYNotes(t, "internal error: GetItems must be specified")
}
