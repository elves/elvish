package etk

import (
	"src.elv.sh/pkg/cli/term"
)

func ComboBox(c Context) (View, React) {
	filterView, filterReact := c.Subcomp("filter", TextArea)
	filterBufferVar := BindState(c, "filter/buffer", TextBuffer{})
	listView, listReact := c.Subcomp("list", ListBox)
	listItemsVar := BindState(c, "list/items", ListItems(nil))
	listSelectedVar := BindState(c, "list/selected", 0)

	genListVar := State(c, "gen-list", func(string) (ListItems, int) {
		return nil, -1
	})
	lastFilterContentVar := State(c, "-last-filter-content", "")

	return VBoxView(0, filterView, listView),
		c.WithBinding(func(ev term.Event) Reaction {
			if reaction := filterReact(ev); reaction != Unused {
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
