package newedit

import (
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
)

func (ed *editor) Ns() eval.Ns {
	return eval.NewNs().
		Add("max-height", vars.FromPtrWithMutex(
			&ed.core.Config.Raw.RenderConfig.MaxHeight, &ed.core.Config.Mutex))
}
