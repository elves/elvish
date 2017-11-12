// Package prompt implements prompt-related functionalities of the editor.
package prompt

import (
	"io/ioutil"
	"math"
	"os"
	"os/user"
	"strconv"
	"sync"
	"time"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/util"
)

var logger = util.GetLogger("[edit/prompt] ")

// Editor is the interface used by the prompt to access the editor.
type Editor interface {
	Evaler() *eval.Evaler
	Variable(string) eval.Variable
	Notify(string, ...interface{})
}

// maxSeconds is the maximum number of seconds time.Duration can represent.
const maxSeconds = float64(math.MaxInt64 / time.Second)

// PromptVariable returns a variable for $edit:prompt.
func PromptVariable() eval.Variable {
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

// Prompt extracts $edit:prompt.
func Prompt(ed Editor) eval.Callable {
	return ed.Variable("prompt").Get().(eval.Callable)
}

// RpromptVariable returns a variable for $edit:rprompt.
func RpromptVariable() eval.Variable {
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

// Rprompt extracts $edit:rprompt.
func Rprompt(ed Editor) eval.Callable {
	return ed.Variable("rprompt").Get().(eval.Callable)
}

// Implementation for $rprompt-persistent.

// RpromptPersistentVariable returns a variable for $edit:rprompt-persistent.
func RpromptPersistentVariable() eval.Variable {
	return eval.NewPtrVariableWithValidator(eval.Bool(false), eval.ShouldBeBool)
}

// RpromptPersistent extracts $edit:rprompt-persistent.
func RpromptPersistent(ed Editor) bool {
	return bool(ed.Variable("rprompt-persistent").Get().(eval.Bool).Bool())
}

// MaxWaitVariable returns a variable for $edit:-prompts-max-wait.
func MaxWaitVariable() eval.Variable {
	return eval.NewPtrVariableWithValidator(eval.String("+Inf"), eval.ShouldBeNumber)
}

// MaxWait extracts $edit:-prompts-max-wait.
func MaxWait(ed Editor) float64 {
	f, _ := strconv.ParseFloat(string(ed.Variable("-prompts-max-wait").Get().(eval.String)), 64)
	return f
}

// MakeMaxWait makes a channel that sends the current time after
// $edit:-prompts-max-wait seconds if the time fits in a time.Duration value, or
// nil otherwise.
func MakeMaxWaitChan(ed Editor) <-chan time.Time {
	f := MaxWait(ed)
	if f > maxSeconds {
		return nil
	}
	return time.After(time.Duration(f * float64(time.Second)))
}

// callPrompt calls a Fn, assuming that it is a prompt. It calls the Fn with no
// arguments and closed input, and converts its outputs to styled objects.
func callPrompt(ed Editor, fn eval.Callable) []*ui.Styled {
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
	ec := eval.NewTopEvalCtx(ed.Evaler(), "[editor prompt]", "", ports)
	err := ec.PCaptureOutputInner(fn, nil, eval.NoOpts, valuesCb, bytesCb)

	if err != nil {
		ed.Notify("prompt function error: %v", err)
		return nil
	}

	return styleds
}

// Updater manages the update of a prompt.
type Updater struct {
	promptFn func(Editor) eval.Callable
	Staled   []*ui.Styled
}

var staledPrompt = &ui.Styled{"?", ui.Styles{"inverse"}}

// NewUpdater creates a new Updater.
func NewUpdater(promptFn func(Editor) eval.Callable) *Updater {
	return &Updater{promptFn, []*ui.Styled{staledPrompt}}
}

// Update updates the prompt, returning a channel onto which the result will be
// written.
func (pu *Updater) Update(ed Editor) <-chan []*ui.Styled {
	ch := make(chan []*ui.Styled)
	go func() {
		result := callPrompt(ed, pu.promptFn(ed))
		pu.Staled = make([]*ui.Styled, len(result)+1)
		pu.Staled[0] = staledPrompt
		copy(pu.Staled[1:], result)
		ch <- result
	}()
	return ch
}
