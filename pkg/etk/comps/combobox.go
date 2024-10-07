package comps

import (
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/etk"
)

func ComboBox(c etk.Context) (etk.View, etk.React) {
	filterView, filterReact := c.Subcomp("filter", TextArea)
	filterBufferVar := etk.BindState(c, "filter/buffer", TextBuffer{})
	listView, listReact := c.Subcomp("list", ListBox)
	listItemsVar := etk.BindState(c, "list/items", ListItems(nil))
	listSelectedVar := etk.BindState(c, "list/selected", 0)

	genListVar := etk.State(c, "gen-list", func(string) (ListItems, int) {
		return nil, -1
	})
	lastFilterContentVar := etk.State(c, "-last-filter-content", "")

	return etk.VBoxView(0, filterView, listView),
		c.WithBinding(func(ev term.Event) etk.Reaction {
			if reaction := filterReact(ev); reaction != etk.Unused {
				filterContent := filterBufferVar.Get().Content
				if filterContent != lastFilterContentVar.Get() {
					lastFilterContentVar.Set(filterContent)
					items, selected := genListVar.Get()(filterContent)
					listItemsVar.Set(items)
					listSelectedVar.Set(selected)
				}
				return reaction
			} else {
				return listReact(ev)
			}
		})
}
