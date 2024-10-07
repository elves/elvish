package comps

import (
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/ui"
)

// ListItems is an interface for accessing multiple items.
type ListItems interface {
	// Len returns the number of items.
	Len() int
	// Show renders the item at the given zero-based index.
	Show(i int) ui.Text
}

type stringItems []string

// StringItems returns a [ListItems] backed up a slice of strings.
func StringItems(items ...string) ListItems { return stringItems(items) }
func (si stringItems) Show(i int) ui.Text   { return ui.T(si[i]) }
func (si stringItems) Len() int             { return len(si) }

func ListBox(c etk.Context) (etk.View, etk.React) {
	itemsVar := etk.State(c, "items", ListItems(nil))
	selectedVar := etk.State(c, "selected", 0)
	horizontalVar := etk.State(c, "horizontal", false)

	selected := selectedVar.Get()
	focus := 0
	var spans []ui.Text
	if items := itemsVar.Get(); items != nil {
		for i := 0; i < items.Len(); i++ {
			if i > 0 {
				if horizontalVar.Get() {
					spans = append(spans, ui.T("  "))
				} else {
					spans = append(spans, ui.T("\n"))
				}
			}
			if i == selected {
				focus = len(spans)
				spans = append(spans, ui.StyleText(items.Show(i), ui.Inverse))
			} else {
				spans = append(spans, items.Show(i))
			}
		}
	}

	return etk.TextView(focus, spans...),
		c.WithBinding(func(e term.Event) etk.Reaction {
			selected := selectedVar.Get()
			items := itemsVar.Get()
			if horizontalVar.Get() {
				switch e {
				case term.K(ui.Left):
					if selected > 0 {
						selectedVar.Set(selected - 1)
						return etk.Consumed
					}
				case term.K(ui.Right):
					if selected < items.Len()-1 {
						selectedVar.Set(selected + 1)
						return etk.Consumed
					}
				}
			} else {
				switch e {
				case term.K(ui.Up):
					if selected > 0 {
						selectedVar.Set(selected - 1)
						return etk.Consumed
					}
				case term.K(ui.Down):
					if selected < items.Len()-1 {
						selectedVar.Set(selected + 1)
						return etk.Consumed
					}
				}
			}
			return etk.Unused
		})
}
