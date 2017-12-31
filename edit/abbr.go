package edit

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/types"
	"github.com/xiaq/persistent/hashmap"
)

var _ = RegisterVariable("abbr", func() eval.Variable {
	return eval.NewPtrVariableWithValidator(
		types.NewMap(hashmap.Empty), eval.ShouldBeMap)
})

func (ed *Editor) abbr() types.Map {
	return ed.variables["abbr"].Get().(types.Map)
}

func (ed *Editor) abbrIterate(cb func(abbr, full string) bool) {
	m := ed.abbr()
	m.IteratePair(func(abbrValue, fullValue types.Value) bool {
		abbr, ok := abbrValue.(eval.String)
		if !ok {
			return true
		}
		full, ok := fullValue.(eval.String)
		if !ok {
			return true
		}
		return cb(string(abbr), string(full))
	})
}
