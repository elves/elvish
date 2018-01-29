package edit

import (
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
)

func styled(ec *eval.Frame, args []interface{}, opts map[string]interface{}) {
	var textv, stylev string
	eval.ScanArgs(args, &textv, &stylev)
	text, style := textv, stylev
	eval.TakeNoOpt(opts)

	out := ec.OutputChan()
	out <- &ui.Styled{text, ui.StylesFromString(style)}
}
