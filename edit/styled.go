package edit

import (
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/types"
)

func styled(ec *eval.Frame, args []types.Value, opts map[string]types.Value) {
	var textv, stylev types.String
	eval.ScanArgs(args, &textv, &stylev)
	text, style := string(textv), string(stylev)
	eval.TakeNoOpt(opts)

	out := ec.OutputChan()
	out <- &ui.Styled{text, ui.StylesFromString(style)}
}
