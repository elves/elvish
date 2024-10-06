package edit

import (
	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/eval"
)

func initCustomWidgetAPI(app cli.App, nb eval.NsBuilder) {
	nb.AddGoFns(map[string]any{
		"push-addon": app.PushAddon,
		"pop-addon":  app.PopAddon,
	})
}
