package main

import (
	"fmt"
	"slices"
	"sort"

	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/etk"
	"src.elv.sh/pkg/etk/comps"
	"src.elv.sh/pkg/ui"
)

func HierNav(c etk.Context) (etk.View, etk.React) {
	dataVar := etk.State(c, "data", map[string]any{})
	data := dataVar.Get()

	pathVar := etk.State(c, "path", []string{})
	path := pathVar.Get()

	var parent etk.View
	if len(path) > 0 {
		parent = hierNavPanel(c, data, path[:len(path)-1])
	} else {
		parent = etk.EmptyView()
	}

	var (
		currentView  etk.View
		currentReact etk.React
		preview      etk.View
		react        func(term.Event) etk.Reaction
	)
	switch value := access(data, path).(type) {
	case map[string]any:
		// TODO: Don't recalculate?
		items := makeHierItems(value)
		currentView, currentReact = c.Subcomp(pathToName(path), etk.WithInit(comps.ListBox, "items", items))
		selectedVar := etk.BindState(c, pathToName(path)+"/selected", 0)
		previewPath := slices.Concat(path, []string{items[selectedVar.Get()].key})
		preview = hierNavPanel(c, data, previewPath)
		react = func(e term.Event) etk.Reaction {
			switch e {
			case term.K(ui.Left):
				if len(path) > 0 {
					pathVar.Set(path[:len(path)-1])
					return etk.Consumed
				}
				return etk.Unused
			case term.K(ui.Right):
				pathVar.Set(previewPath)
				return etk.Consumed
			default:
				return currentReact(e)
			}
		}
	case string:
		currentView = etk.TextView(0, ui.T(value))
		currentReact = func(term.Event) etk.Reaction { return etk.Unused }
		preview = etk.EmptyView()
		react = func(term.Event) etk.Reaction { return etk.Unused }
	}

	return etk.VBoxView(
		0,
		etk.TextView(0, ui.T(fmt.Sprintf("path = %s", path))),
		etk.HBoxView(1, parent, currentView, preview),
	), react
}

func hierNavPanel(b etk.Context, data map[string]any, path []string) etk.View {
	switch value := access(data, path).(type) {
	case map[string]any:
		items := makeHierItems(value)
		view, _ := b.Subcomp(pathToName(path), etk.WithInit(comps.ListBox, "items", items))
		return view
	case string:
		return etk.TextView(0, ui.T(value))
	default:
		panic("unreachable")
	}
}

func access(data map[string]any, path []string) any {
	for len(path) > 0 {
		if subData, ok := data[path[0]]; ok {
			path = path[1:]
			switch subData := subData.(type) {
			case map[string]any:
				data = subData
			case string:
				if len(path) == 0 {
					return subData
				}
				return "not found"
			default:
				panic("unreachable")
			}
		} else {
			return "not found"
		}
	}
	return data
}

func pathToName(path []string) string { return fmt.Sprint(path) }

type hierItem struct {
	key   string
	value any
}

type hierItems []hierItem

func makeHierItems(value map[string]any) hierItems {
	var items hierItems
	for k, v := range value {
		items = append(items, hierItem{k, v})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].key < items[j].key })
	return items
}

func (hi hierItems) Len() int { return len(hi) }

func (hi hierItems) Show(i int) ui.Text {
	switch hi[i].value.(type) {
	case map[string]any:
		return ui.T(hi[i].key, ui.FgGreen, ui.Bold)
	default:
		return ui.T(hi[i].key)
	}
}

var hierData = map[string]any{
	"bin": map[string]any{
		"cat":    "Concatenate files",
		"elvish": "Elvish shell",
		"zsh":    "The Z shell",
	},
	"home": map[string]any{
		"elf": map[string]any{
			"bin": map[string]any{
				"elvish": "Local Elvish build",
				"foo":    "bar",
			},
			"README": "this is the elf user's home directory.",
		},
		"root": map[string]any{
			"README": "this is the root user's home directory.",
		},
	},
	"README": "this is the root.",
}
