package comps

import (
	"fmt"
	"slices"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/ui"
)

type Hier interface {
	Get(path []string) (ListItems, string)
	OnCurrentPathChange(path []string)
}

func HierNav(c etk.Context) (etk.View, etk.React) {
	hier := etk.State(c, "hier", Hier(nil)).Get()

	pathVar := etk.State(c, "path", []string{})
	path := pathVar.Get()

	var parent etk.View = etk.EmptyView{}
	if len(path) > 0 {
		np := len(path)
		parent, _ = hierNavPanel(c, hier, path[:np-1], path[np-1])
	}

	currentView, currentReact := hierNavPanel(c, hier, path, "")

	var (
		preview     etk.View = etk.EmptyView{}
		previewPath []string
	)
	// TODO: This will work not if path itself contains "/"
	selectedPath := pathToName(path) + "/list/selected"
	if c.Get(selectedPath) != nil {
		items := etk.BindState(c, pathToName(path)+"/list/items", ListItems(nil)).Get()
		selected := etk.BindState(c, pathToName(path)+"/list/selected", 0).Get()
		if 0 <= selected && selected < items.Len() {
			previewPath = slices.Concat(path, []string{items.Get(selected).(string)})
			preview, _ = hierNavPanel(c, hier, previewPath, "")
		}
	}

	return etk.Box("parent*1 1 [current*3] 1 preview*4", parent, currentView, preview),
		func(e term.Event) etk.Reaction {
			switch e {
			case term.K(ui.Left):
				if len(path) > 0 {
					pathVar.Set(path[:len(path)-1])
					return etk.Consumed
				}
				return etk.Unused
			case term.K(ui.Right):
				if previewPath != nil {
					pathVar.Set(previewPath)
					return etk.Consumed
				}
			default:
				return currentReact(e)
			}
			return etk.Unused
		}

}

func hierNavPanel(c etk.Context, h Hier, path []string, toSelect string) (etk.View, etk.React) {
	name := pathToName(path)
	if c.Get(name+"-comp") != nil {
		return c.Subcomp(name, nil)
	}

	items, s := h.Get(path)
	if items != nil {
		selected := 0
		if toSelect != "" {
			for i := 0; i < items.Len(); i++ {
				if toSelect == items.Get(i) {
					selected = i
					break
				}
			}
		}
		return c.Subcomp(name,
			etk.ModComp(
				etk.State(c, "inner-node-comp", ComboBox).Get(),
				etk.InitState("gen-list", func(query string) (ListItems, int) {
					return items, selected
				}),
				etk.InitState("list/left-padding", 1),
				etk.InitState("list/right-padding", 1)))
	} else {
		buffer := TextBuffer{Content: s}
		return c.Subcomp(name,
			etk.ModComp(
				etk.State(c, "leaf-node-comp", TextArea).Get(),
				etk.InitState("buffer", buffer)))
	}
}

func pathToName(path []string) string { return fmt.Sprint(path) }
