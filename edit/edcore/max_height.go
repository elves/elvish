package edcore

import (
	"math"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/util"
)

func init() {
	atEditorInit(func(ed *editor, ns eval.Ns) {
		ed.maxHeight = math.Inf(1)
		ns["max-height"] = vars.FromPtr(&ed.maxHeight)
	})
}

func maxHeightToInt(h float64) int {
	if math.IsInf(h, 1) {
		return util.MaxInt
	}
	return int(h)
}
