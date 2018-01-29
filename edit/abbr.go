package edit

import (
	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/eval/vartypes"
	"github.com/xiaq/persistent/hashmap"
)

var _ = RegisterVariable("abbr", func() vartypes.Variable {
	return vartypes.NewValidatedPtr(types.EmptyMap, vartypes.ShouldBeMap)
})

func (ed *Editor) abbr() hashmap.Map {
	return ed.variables["abbr"].Get().(hashmap.Map)
}

func (ed *Editor) abbrIterate(cb func(abbr, full string) bool) {
	m := ed.abbr()
	for it := m.Iterator(); it.HasElem(); it.Next() {
		abbrValue, fullValue := it.Elem()
		abbr, ok := abbrValue.(string)
		if !ok {
			continue
		}
		full, ok := fullValue.(string)
		if !ok {
			continue
		}
		if !cb(abbr, full) {
			break
		}
	}
}
