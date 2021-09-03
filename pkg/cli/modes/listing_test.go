package modes

import (
	"testing"

	"src.elv.sh/pkg/cli"
	. "src.elv.sh/pkg/cli/clitest"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/ui"
)

func fooAndGreenBar(string) ([]ListingItem, int) {
	return []ListingItem{{"foo", ui.T("foo")}, {"bar", ui.T("bar", ui.FgGreen)}}, 0
}

func TestListing_BasicUI(t *testing.T) {
	f := Setup()
	defer f.Stop()

	startListing(f.App, ListingSpec{
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

func TestListing_Accept_ClosingListing(t *testing.T) {
	f := Setup()
	defer f.Stop()

	startListing(f.App, ListingSpec{
		GetItems: fooAndGreenBar,
		Accept: func(t string) {
			f.App.ActiveWidget().(tk.CodeArea).MutateState(func(s *tk.CodeAreaState) {
				s.Buffer.InsertAtDot(t)
			})
		},
	})
	// foo will be selected
	f.TTY.Inject(term.K('\n'))
	f.TestTTY(t, "foo", term.DotHere)
}

func TestListing_Accept_DefaultNop(t *testing.T) {
	f := Setup()
	defer f.Stop()

	startListing(f.App, ListingSpec{GetItems: fooAndGreenBar})
	f.TTY.Inject(term.K('\n'))
	f.TestTTY(t /* nothing */)
}

func TestListing_AutoAccept(t *testing.T) {
	f := Setup()
	defer f.Stop()

	startListing(f.App, ListingSpec{
		GetItems: func(query string) ([]ListingItem, int) {
			if query == "" {
				// Return two items initially.
				return []ListingItem{
					{"foo", ui.T("foo")}, {"bar", ui.T("bar")},
				}, 0
			}
			return []ListingItem{{"bar", ui.T("bar")}}, 0
		},
		Accept: func(t string) {
			f.App.ActiveWidget().(tk.CodeArea).MutateState(func(s *tk.CodeAreaState) {
				s.Buffer.InsertAtDot(t)
			})
		},
		AutoAccept: true,
	})
	f.TTY.Inject(term.K('a'))
	f.TestTTY(t, "bar", term.DotHere)
}

func TestNewListing_NoGetItems(t *testing.T) {
	f := Setup()
	defer f.Stop()

	_, err := NewListing(f.App, ListingSpec{})
	if err != errGetItemsMustBeSpecified {
		t.Error("expect errGetItemsMustBeSpecified")
	}
}

func startListing(app cli.App, spec ListingSpec) {
	w, err := NewListing(app, spec)
	startMode(app, w, err)
}
