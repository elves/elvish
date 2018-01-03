package edit

import (
	"math"
	"strconv"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/types"
	"github.com/elves/elvish/eval/vartypes"
	"github.com/elves/elvish/util"
)

var _ = RegisterVariable("max-height", func() vartypes.Variable {
	return vartypes.NewValidatedPtrVariable(types.String("+Inf"), eval.ShouldBeNumber)
})

func (ed *Editor) maxHeight() int {
	f, _ := strconv.ParseFloat(string(ed.variables["max-height"].Get().(types.String)), 64)
	if math.IsInf(f, 1) {
		return util.MaxInt
	}
	return int(f)
}
