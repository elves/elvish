package prompt

import (
	"os"
	"os/user"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/util"
)

var defaultPrompt, defaultRPrompt, defaultStaleTransform eval.Callable

func init() {
	user, err := user.Current()
	isRoot := err == nil && user.Uid == "0"

	prompt := func(fm *eval.Frame) {
		out := fm.OutputChan()
		out <- string(util.Getwd())
		if isRoot {
			out <- &ui.Styled{"# ", ui.Styles{"red"}}
		} else {
			out <- &ui.Styled{"> ", ui.Styles{}}
		}
	}
	defaultPrompt = eval.NewBuiltinFn("default prompt", prompt)
}

func init() {
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
	rprompt := func(fm *eval.Frame) {
		out := fm.OutputChan()
		out <- &ui.Styled{rpromptStr, ui.Styles{"inverse"}}
	}
	defaultRPrompt = eval.NewBuiltinFn("default rprompt", rprompt)
}

func init() {
	staleTransform := func(fm *eval.Frame) {
		out := fm.OutputChan()
		fm.IterateInputs(func(i interface{}) {
			s := i.(*ui.Styled)
			out <- &ui.Styled{s.Text, ui.Styles{"inverse"}}
		})
	}
	defaultStaleTransform = eval.NewBuiltinFn("default stale transform", staleTransform)
}
