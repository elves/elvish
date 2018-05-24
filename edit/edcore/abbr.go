package edcore

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
	"github.com/xiaq/persistent/hashmap"
)

func init() {
	atEditorInit(func(ed *editor, ns eval.Ns) {
		ed.abbr = vals.EmptyMap
		ns["abbr"] = vars.FromPtr(&ed.abbr)
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
