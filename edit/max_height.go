package edit

import (
	"math"
	"strconv"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vartypes"
	"github.com/elves/elvish/util"
)

var _ = RegisterVariable("max-height", func() vartypes.Variable {
	maxHeight := math.Inf(1)
	return eval.NewVariableFromPtr(&maxHeight)
})

func (ed *Editor) maxHeight() int {
	f, _ := strconv.ParseFloat(ed.variables["max-height"].Get().(string), 64)
	if math.IsInf(f, 1) {
		return util.MaxInt
	}
	return int(f)
}
