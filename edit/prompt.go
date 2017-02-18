package edit

import (
	"os"
	"os/user"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/util"
)

func defaultPrompts() (eval.CallableValue, eval.CallableValue) {
	// Make default prompts.
	prompt := func(ec *eval.EvalCtx, args []eval.Value, opts map[string]eval.Value) {
		out := ec.OutputChan()
		out <- &styled{util.Getwd() + "> ", styles{}}
	}

	username := "???"
	user, err := user.Current()
	if err == nil {
		username = user.Username
	}
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "???"
	}
	rpromptStr := username + "@" + hostname
	rprompt := func(ec *eval.EvalCtx, args []eval.Value, opts map[string]eval.Value) {
		out := ec.OutputChan()
		out <- &styled{rpromptStr, styles{"7"}}
	}

	return &eval.BuiltinFn{"default prompt", prompt}, &eval.BuiltinFn{"default rprompt", rprompt}
}
