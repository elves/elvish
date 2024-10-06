package comps

import (
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/etk"
)

func ComboBox(c etk.Context) (etk.View, etk.React) {
	queryView, queryReact := c.Subcomp("query", TextArea)
	queryBufferVar := etk.BindState(c, "query/buffer", TextBuffer{})

	listItemsVar := etk.BindState(c, "list/items", ListItems(nil))
	listSelectedVar := etk.BindState(c, "list/selected", 0)

	genListVar := etk.State(c, "gen-list", func(string) (ListItems, int) {
		return nil, -1
	})
	lastQueryContentVar := etk.State(c, "-last-query-content", "")
	regenListFromQuery := func(queryContent string) {
		lastQueryContentVar.Set(queryContent)
		items, selected := genListVar.Get()(queryContent)
		listItemsVar.Set(items)
		listSelectedVar.Set(selected)
	}
	// Note: It's important to make the initial call to gen-list before the
	// ListBox subcomponent, otherwise the result of that initial call will not
	// appear. This is also why we need to declare the ListBox state variables
	// before actually calling the subcomponent.
	initializedVar := etk.State(c, "-initialized", false)
	if !initializedVar.Get() {
		initializedVar.Set(true)
		regenListFromQuery(queryBufferVar.Get().Content)
	}

	listView, listReact := c.Subcomp("list", ListBox)

	binding := c.BindingNopDefault()

	return etk.Box(`
			[query=]
			list*`, queryView, listView),
		func(ev term.Event) etk.Reaction {
			if reaction := binding(ev); reaction != etk.Unused {
				return reaction
			} else if reaction := queryReact(ev); reaction != etk.Unused {
				queryContent := queryBufferVar.Get().Content
				if queryContent != lastQueryContentVar.Get() {
					regenListFromQuery(queryContent)
				}
				return reaction
			} else {
				return listReact(ev)
			}
		}
}
