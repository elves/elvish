package edit

import (
	"errors"
	"os"
	"os/user"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/util"
)

var ErrMustBeFn = errors.New("must be function")

// MustBeFn validates whether a Value is an Fn.
func MustBeFn(v eval.Value) error {
	if _, ok := v.(eval.Fn); !ok {
		return ErrMustBeFn
	}
	return nil
}

func defaultPrompts() (eval.FnValue, eval.FnValue) {
	// Make default prompts.
	prompt := func(ec *eval.EvalCtx, args []eval.Value, opts map[string]eval.Value) {
		out := ec.OutputChan()
		out <- &styled{util.Getwd() + "> ", ""}
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
		out <- &styled{rpromptStr, "7"}
	}

	return &eval.BuiltinFn{"default prompt", prompt}, &eval.BuiltinFn{"default rprompt", rprompt}
}
