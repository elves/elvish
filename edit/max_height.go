package edit

import (
	"math"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/util"
)

func init() {
	atEditorInit(func(ed *Editor, ns eval.Ns) {
		ed.maxHeight = math.Inf(1)
		ns["max-height"] = eval.NewVariableFromPtr(&ed.maxHeight)
	})
}

func maxHeightToInt(h float64) int {
	if math.IsInf(h, 1) {
		return util.MaxInt
	}
	return int(h)
}
