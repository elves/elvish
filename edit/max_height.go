package edit

import (
	"math"
	"strconv"

	"github.com/elves/elvish/eval/vartypes"
	"github.com/elves/elvish/util"
)

var _ = RegisterVariable("max-height", func() vartypes.Variable {
	return vartypes.NewValidatedPtr("+Inf", vartypes.ShouldBeNumber)
})

func (ed *Editor) maxHeight() int {
	f, _ := strconv.ParseFloat(ed.variables["max-height"].Get().(string), 64)
	if math.IsInf(f, 1) {
		return util.MaxInt
	}
	return int(f)
}
