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
	prompt := func(ec *eval.EvalCtx, args []eval.Value) {
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
	rprompt := func(ec *eval.EvalCtx, args []eval.Value) {
		out := ec.OutputChan()
		out <- &styled{rpromptStr, "7"}
	}

	return &eval.BuiltinFn{"default prompt", prompt}, &eval.BuiltinFn{"default rprompt", rprompt}
}

// callFnAsPrompt calls a Fn with closed input, captures its output and convert
// the output to a slice of *styled's.
func callFnForPrompt(ed *Editor, fn eval.Fn) []*styled {
	in, err := makeClosedStdin()
	if err != nil {
		return nil
	}
	ports := []*eval.Port{in, &eval.Port{File: os.Stdout}, &eval.Port{File: os.Stderr}}

	// XXX There is no source to pass to NewTopEvalCtx.
	ec := eval.NewTopEvalCtx(ed.evaler, "[editor prompt]", "", ports)
	values, err := ec.PCaptureOutput(fn, nil, eval.NoOpts)
	if err != nil {
		ed.Notify("prompt function error: %v", err)
		return nil
	}

	var ss []*styled
	for _, v := range values {
		if s, ok := v.(*styled); ok {
			ss = append(ss, s)
		} else {
			ss = append(ss, &styled{eval.ToString(v), ""})
		}
	}
	return ss
}
