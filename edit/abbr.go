package edit

import (
	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/eval/vartypes"
)

var _ = RegisterVariable("abbr", func() vartypes.Variable {
	return vartypes.NewValidatedPtr(
		types.NewMap(types.EmptyMapInner), vartypes.ShouldBeMap)
})

func (ed *Editor) abbr() types.Map {
	return ed.variables["abbr"].Get().(types.Map)
}

func (ed *Editor) abbrIterate(cb func(abbr, full string) bool) {
	m := ed.abbr()
	m.IteratePair(func(abbrValue, fullValue types.Value) bool {
		abbr, ok := abbrValue.(string)
		if !ok {
			return true
		}
		full, ok := fullValue.(string)
		if !ok {
			return true
		}
		return cb(abbr, full)
	})
}
