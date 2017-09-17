package edit

import (
	"os"
	"os/user"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/util"
)

var _ = registerVariable("prompt", promptVariable)

func promptVariable() eval.Variable {
	user, err := user.Current()
	isRoot := err == nil && user.Uid == "0"

	prompt := func(ec *eval.EvalCtx,
		args []eval.Value, opts map[string]eval.Value) {

		out := ec.OutputChan()
		out <- eval.String(util.Getwd())
		if isRoot {
			out <- &ui.Styled{"# ", ui.Styles{"red"}}
		} else {
			out <- &ui.Styled{"> ", ui.Styles{}}
		}
	}
	return eval.NewPtrVariableWithValidator(
		&eval.BuiltinFn{"default prompt", prompt}, eval.ShouldBeFn)
}

func (ed *Editor) prompt() eval.Callable {
	return ed.variables["prompt"].Get().(eval.Callable)
}

var _ = registerVariable("rprompt", rpromptVariable)

func rpromptVariable() eval.Variable {
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
	rprompt := func(ec *eval.EvalCtx,
		args []eval.Value, opts map[string]eval.Value) {

		out := ec.OutputChan()
		out <- &ui.Styled{rpromptStr, ui.Styles{"inverse"}}
	}

	return eval.NewPtrVariableWithValidator(
		&eval.BuiltinFn{"default rprompt", rprompt}, eval.ShouldBeFn)
}

func (ed *Editor) rprompt() eval.Callable {
	return ed.variables["rprompt"].Get().(eval.Callable)
}

var _ = registerVariable("rprompt-persistent", func() eval.Variable {
	return eval.NewPtrVariableWithValidator(eval.Bool(false), eval.ShouldBeBool)
})

func (ed *Editor) rpromptPersistent() bool {
	return bool(ed.variables["rprompt-persistent"].Get().(eval.Bool).Bool())
}
