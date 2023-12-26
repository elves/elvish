package modes

import (
	"errors"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/ui"
)

// Listing is a customizable mode for browsing through a list of items. It is
// based on the ComboBox widget.
type Listing interface {
	tk.ComboBox
}

// ListingSpec specifies the configuration for the listing mode.
type ListingSpec struct {
	// Key bindings.
	Bindings tk.Bindings
	// Caption of the listing. If empty, defaults to " LISTING ".
	Caption string
	// A function that takes the query string and returns a list of Item's and
	// the index of the Item to select. Required.
	GetItems func(query string) (items []ListingItem, selected int)
	// A function to call when the user has accepted the selected item. If the
	// return value is true, the listing will not be closed after accepting.
	// If unspecified, the Accept function default to a function that does
	// nothing other than returning false.
	Accept func(string)
	// Whether to automatically accept when there is only one item.
	AutoAccept bool
}

// ListingItem is an item to show in the listing.
type ListingItem struct {
	// Passed to the Accept callback in Config.
	ToAccept string
	// How the item is shown.
	ToShow ui.Text
}

var errGetItemsMustBeSpecified = errors.New("GetItems must be specified")

// NewListing creates a new listing mode.
func NewListing(app cli.App, spec ListingSpec) (Listing, error) {
	if spec.GetItems == nil {
		return nil, errGetItemsMustBeSpecified
	}
	if spec.Accept == nil {
		spec.Accept = func(string) {}
	}
	if spec.Caption == "" {
		spec.Caption = " LISTING "
	}
	accept := func(s string) {
		app.PopAddon()
		spec.Accept(s)
	}
	w := tk.NewComboBox(tk.ComboBoxSpec{
		CodeArea: tk.CodeAreaSpec{
			Prompt: modePrompt(spec.Caption, true),
		},
		ListBox: tk.ListBoxSpec{
			Bindings: spec.Bindings,
			OnAccept: func(it tk.Items, i int) {
				accept(it.(listingItems)[i].ToAccept)
			},
			ExtendStyle: true,
		},
		OnFilter: func(w tk.ComboBox, q string) {
			it, selected := spec.GetItems(q)
			w.ListBox().Reset(listingItems(it), selected)
			if spec.AutoAccept && len(it) == 1 {
				accept(it[0].ToAccept)
			}
		},
	})
	return w, nil
}

type listingItems []ListingItem

func (it listingItems) Len() int           { return len(it) }
func (it listingItems) Show(i int) ui.Text { return it[i].ToShow }
