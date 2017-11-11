package edit

import (
	"io/ioutil"
	"os"
	"os/user"
	"sync"

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

// callPrompt calls a Fn, assuming that it is a prompt. It calls the Fn with no
// arguments and closed input, and converts its outputs to styled objects.
func callPrompt(ed *Editor, fn eval.Callable) []*ui.Styled {
	ports := []*eval.Port{
		eval.DevNullClosedChan,
		{}, // Will be replaced when capturing output
		{File: os.Stderr},
	}
	var (
		styleds      []*ui.Styled
		styledsMutex sync.Mutex
	)
	add := func(s *ui.Styled) {
		styledsMutex.Lock()
		styleds = append(styleds, s)
		styledsMutex.Unlock()
	}
	// Value output may be of type ui.Styled or any other type, in which case
	// they are converted to ui.Styled.
	valuesCb := func(ch <-chan eval.Value) {
		for v := range ch {
			if s, ok := v.(*ui.Styled); ok {
				add(s)
			} else {
				add(&ui.Styled{eval.ToString(v), ui.Styles{}})
			}
		}
	}
	// Byte output is added to the prompt as a single unstyled text.
	bytesCb := func(r *os.File) {
		allBytes, err := ioutil.ReadAll(r)
		if err != nil {
			logger.Println("error reading prompt byte output:", err)
		}
		if len(allBytes) > 0 {
			add(&ui.Styled{string(allBytes), ui.Styles{}})
		}
	}

	// XXX There is no source to pass to NewTopEvalCtx.
	ec := eval.NewTopEvalCtx(ed.evaler, "[editor prompt]", "", ports)
	err := ec.PCaptureOutputInner(fn, nil, eval.NoOpts, valuesCb, bytesCb)

	if err != nil {
		ed.Notify("prompt function error: %v", err)
		return nil
	}

	return styleds
}
