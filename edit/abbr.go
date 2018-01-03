package edit

import (
	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/eval/vartypes"
	"github.com/xiaq/persistent/hashmap"
)

var _ = RegisterVariable("abbr", func() vartypes.Variable {
	return vartypes.NewValidatedPtrVariable(
		types.NewMap(hashmap.Empty), vartypes.ShouldBeMap)
})

func (ed *Editor) abbr() types.Map {
	return ed.variables["abbr"].Get().(types.Map)
}

func (ed *Editor) abbrIterate(cb func(abbr, full string) bool) {
	m := ed.abbr()
	m.IteratePair(func(abbrValue, fullValue types.Value) bool {
		abbr, ok := abbrValue.(types.String)
		if !ok {
			return true
		}
		full, ok := fullValue.(types.String)
		if !ok {
			return true
		}
		return cb(string(abbr), string(full))
	})
}
