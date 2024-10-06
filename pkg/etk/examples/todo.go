package main

import (
	"fmt"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/etk/comps"
	"src.elv.sh/pkg/ui"
)

type todoItem struct {
	text string
	done bool
}

type todoItems []todoItem

func (ti todoItems) Len() int { return len(ti) }
func (ti todoItems) Show(i int) ui.Text {
	done := ' '
	if ti[i].done {
		done = 'X'
	}
	return ui.T(fmt.Sprintf("[%c] %s", done, ti[i].text))
}

func Todo(c etk.Context) (etk.View, etk.React) {
	// TODO: API to combine init and bind
	listView, listReact := c.Subcomp("list", etk.WithInit(comps.ListBox, "items", todoItems{}))
	itemsVar := etk.BindState(c, "list/items", todoItems(nil))
	selectedVar := etk.BindState(c, "list/selected", 0)

	newItemView, newItemReact := c.Subcomp("new-item", etk.WithInit(comps.TextArea, "prompt", ui.T("new item: ")))
	bufferVar := etk.BindState(c, "new-item/buffer", comps.TextBuffer{})

	focusVar := etk.State(c, "focus", 1)
	focus := focusVar.Get()

	return etk.BoxFocus(`
			list=
			newItem=`, focus, listView, newItemView),
		func(e term.Event) etk.Reaction {
			if e == term.K(ui.Tab) {
				focusVar.Set(1 - focus)
				return etk.Consumed
			}
			if focus == 0 {
				reaction := listReact(e)
				if reaction == etk.Unused {
					switch e {
					case term.K(ui.Down):
						focusVar.Set(1)
						return etk.Consumed
					case term.K(' '):
						item := &itemsVar.Get()[selectedVar.Get()]
						item.done = !item.done
						return etk.Consumed
					}
				}
				return reaction
			} else {
				reaction := newItemReact(e)
				if reaction == etk.Unused {
					switch e {
					case term.K(ui.Up):
						focusVar.Set(0)
						return etk.Consumed
					case term.K(ui.Enter):
						itemsVar.Set(append(itemsVar.Get(), todoItem{text: bufferVar.Get().Content}))
						bufferVar.Set(comps.TextBuffer{})
						return etk.Consumed
					}
				}
				return reaction
			}
		}
}
