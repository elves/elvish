package comps

import (
	"fmt"
	"slices"
	"strings"

	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/ui"
)

// var _ Hier = MapHier{}

type MapHier struct {
	Map vals.Map
}

func (mh MapHier) Get(path []string) (any, string) {
	m := mh.Map
	path0 := path
	for len(path) > 0 {
		if subData, ok := m.Index(path[0]); ok {
			path = path[1:]
			switch subData := subData.(type) {
			case vals.Map:
				m = subData
			case string:
				if len(path) == 0 {
					return nil, subData
				}
				return nil, fmt.Sprintf("not found: %v", path0)
			default:
				return nil, fmt.Sprintf("not found: %v", path0)
			}
		} else {
			return nil, fmt.Sprintf("not found: %v", path0)
		}
	}
	return makeMapListItems(m), ""
}

func (mh MapHier) OnCurrentPathChange(path []string) {}

type mapListItem struct {
	name  string
	isMap bool
}

type mapListItems []mapListItem

func makeMapListItems(m vals.Map) mapListItems {
	var items mapListItems
	for it := m.Iterator(); it.HasElem(); it.Next() {
		k, v := it.Elem()
		_, isMap := v.(vals.Map)
		items = append(items, mapListItem{vals.ToString(k), isMap})
	}
	slices.SortFunc(items, func(a, b mapListItem) int {
		return strings.Compare(a.name, b.name)
	})
	return items
}

func (mli mapListItems) Len() int      { return len(mli) }
func (mli mapListItems) Get(i int) any { return mli[i].name }

func (mli mapListItems) Show(i int) (ui.Text, ui.Styling) {
	areaStyling := ui.Nop
	if mli[i].isMap {
		areaStyling = ui.FgMagenta
	}
	return ui.T(mli[i].name), areaStyling
}
