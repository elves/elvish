package edit

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/xiaq/persistent/hashmap"
)

func init() {
	atEditorInit(func(ed *editor, ns eval.Ns) {
		ed.abbr = vals.EmptyMap
		ns["abbr"] = eval.NewVariableFromPtr(&ed.abbr)
	})
}

func abbrIterate(abbr hashmap.Map, cb func(abbr, full string) bool) {
	for it := abbr.Iterator(); it.HasElem(); it.Next() {
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
