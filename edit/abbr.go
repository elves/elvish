package edit

import (
	"github.com/elves/elvish/eval"
	"github.com/xiaq/persistent/hashmap"
)

var _ = registerVariable("abbr", func() eval.Variable {
	return eval.NewPtrVariableWithValidator(
		eval.NewMap(hashmap.Empty), eval.ShouldBeMap)
})

func (ed *Editor) abbr() eval.Map {
	return ed.variables["abbr"].Get().(eval.Map)
}

func (ed *Editor) abbrIterate(cb func(abbr, full string) bool) {
	m := ed.abbr()
	m.IteratePair(func(abbrValue, fullValue eval.Value) bool {
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
