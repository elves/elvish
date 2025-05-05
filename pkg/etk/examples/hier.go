package main

import (
	"fmt"
	"os"
	"path"
	"sort"

	"src.elv.sh/pkg/etk/comps"
	"src.elv.sh/pkg/ui"
)

// Hier implementations.

type dataHier struct{ data map[string]any }

var _ comps.Hier = dataHier{}

func (h dataHier) Get(path []string) (comps.ListItems, string) {
	data := h.data
	path0 := path
	for len(path) > 0 {
		if subData, ok := data[path[0]]; ok {
			path = path[1:]
			switch subData := subData.(type) {
			case map[string]any:
				data = subData
			case string:
				if len(path) == 0 {
					return nil, subData
				}
				return nil, fmt.Sprintf("not found: %v", path0)
			default:
				panic("unreachable")
			}
		} else {
			return nil, fmt.Sprintf("not found: %v", path0)
		}
	}
	return makeHierItems(data), ""
}

func (h dataHier) OnCurrentPathChange(path []string) {}

type dataHierItem struct {
	name  string
	isMap bool
}

type dataHierItems []dataHierItem

func makeHierItems(value map[string]any) dataHierItems {
	var items dataHierItems
	for k, v := range value {
		_, isMap := v.(map[string]any)
		items = append(items, dataHierItem{k, isMap})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].name < items[j].name })
	return items
}

func (hi dataHierItems) Len() int { return len(hi) }

func (hi dataHierItems) Get(i int) any { return hi[i].name }

func (hi dataHierItems) Show(i int) ui.Text {
	return ui.T(hi[i].name)
}

func (hi dataHierItems) StyleLine(i int) ui.Styling {
	if hi[i].isMap {
		return ui.Stylings(ui.FgGreen, ui.Bold)
	} else {
		return ui.Nop
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

type fsHier struct{}

var _ comps.Hier = fsHier{}

func (h fsHier) Get(pathSlice []string) (comps.ListItems, string) {
	file, err := os.Open("/" + path.Join(pathSlice...))
	if err != nil {
		return nil, fmt.Sprintf("error: %v", err)
	}
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Sprintf("error: %v", err)
	}
	if stat.IsDir() {
		entries, err := file.ReadDir(-1)
		if err != nil {
			return nil, fmt.Sprintf("error: %v", err)
		}
		items := make([]fsHierItem, len(entries))
		for i, entry := range entries {
			items[i] = fsHierItem{entry.Name(), entry.IsDir()}
		}
		sort.Slice(items, func(i, j int) bool { return items[i].name < items[j].name })
		return fsHierItems(items), ""
	} else {
		var bs [4 * 1024]byte
		n, err := file.Read(bs[:])
		if err != nil {
			return nil, fmt.Sprintf("error: %v", err)
		}
		return nil, string(bs[:n])
	}
}

func (h fsHier) OnCurrentPathChange(path []string) {}

type fsHierItem struct {
	name  string
	isDir bool
}

type fsHierItems []fsHierItem

func (hi fsHierItems) Len() int { return len(hi) }

func (hi fsHierItems) Get(i int) any { return hi[i].name }

func (hi fsHierItems) Show(i int) ui.Text {
	return ui.T(hi[i].name)
}

func (hi fsHierItems) StyleLine(i int) ui.Styling {
	if hi[i].isDir {
		return ui.Stylings(ui.FgGreen, ui.Bold)
	} else {
		return ui.Nop
	}
}
