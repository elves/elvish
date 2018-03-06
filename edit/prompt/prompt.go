// Package prompt implements prompt-related functionalities of the editor.
package prompt

import (
	"io/ioutil"
	"math"
	"os"
	"os/user"
	"sync"
	"time"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/util"
)

var logger = util.GetLogger("[edit/prompt] ")

// Config holds the config for the prompt
type Config struct {
	Prompt               eval.Callable
	Rprompt              eval.Callable
	StalePromptTransform eval.Callable

	RpromptPersistent bool
	PromptsMaxWait    float64
}

// Editor is the interface used by the prompt to access the editor.
type Editor interface {
	Evaler() *eval.Evaler
	Notify(string, ...interface{})
}

// maxSeconds is the maximum number of seconds time.Duration can represent.
const maxSeconds = float64(math.MaxInt64 / time.Second)

// DefaultPromptInit returns an initial value for $edit:prompt.
func DefaultPromptInit() eval.Callable {
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
	return eval.NewBuiltinFn("default prompt", prompt)
}

// DefaultRpromptInit returns an initial value for $edit:rprompt.
func DefaultRpromptInit() eval.Callable {
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

	return eval.NewBuiltinFn("default rprompt", rprompt)
}

// StalePromptTransformInit returns an initial value for $edit:-stale-prompt-transform
func StalePromptTransformInit() eval.Callable {
	stalePromptTransform := func(fm *eval.Frame) {
		out := fm.OutputChan()
		fm.IterateInputs(func(i interface{}) {
			out <- i
		})
	}

	return eval.NewBuiltinFn("default stale prompt transform", stalePromptTransform)
}

// MakeMaxWaitChan makes a channel that sends the current time after
// $edit:-prompts-max-wait seconds if the time fits in a time.Duration value, or
// nil otherwise.
func (cfg *Config) MakeMaxWaitChan() <-chan time.Time {
	f := cfg.PromptsMaxWait
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

	return callAndGetStyled(ed, fn, ports)
}

func callAndGetStyled(ed Editor, fn eval.Callable, ports []*eval.Port) []*ui.Styled {
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
	valuesCb := func(ch <-chan interface{}) {
		for v := range ch {
			if s, ok := v.(*ui.Styled); ok {
				add(s)
			} else {
				add(&ui.Styled{vals.ToString(v), ui.Styles{}})
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
	ec := eval.NewTopFrame(ed.Evaler(), eval.NewInternalSource("[prompt]"), ports)
	err := ec.CallWithOutputCallback(fn, nil, eval.NoOpts, valuesCb, bytesCb)

	if err != nil {
		ed.Notify("prompt function error: %v", err)
		return nil
	}

	return styleds
}

// callTransformer calls a Fn, assuming that it is a prompt transformer. It calls the Fn with no
// arguments and input, and converts its outputs to styled objects.
func callTransformer(ed Editor, fn eval.Callable, currentPrompt []*ui.Styled) []*ui.Styled {
	input := make(chan interface{})
	stopInputWriter := make(chan struct{})

	ports := []*eval.Port{
		{Chan: input, File: eval.DevNull},
		{}, // Will be replaced when capturing output
		{File: os.Stderr},
	}
	go func() {
		defer close(input)
		for _, char := range currentPrompt {
			select {
			case input <- char:
			case <-stopInputWriter:
				return
			}
		}
	}()
	defer close(stopInputWriter)

	return callAndGetStyled(ed, fn, ports)
}

// Updater manages the update of a prompt.
type Updater struct {
	promptFn         eval.Callable
	staleTransformFn eval.Callable
}

// NewUpdater creates a new Updater.
func NewUpdater(promptFn eval.Callable, staleTransformFn eval.Callable) *Updater {
	return &Updater{
		promptFn:         promptFn,
		staleTransformFn: staleTransformFn,
	}
}

// StalePromptTransformed returns the prompt transformed
func (pu *Updater) StalePromptTransformed(ed Editor, currentPrompt []*ui.Styled) []*ui.Styled {
	return callTransformer(ed, pu.staleTransformFn, currentPrompt)
}

// Update updates the prompt, returning a channel onto which the result will be
// written.
func (pu *Updater) Update(ed Editor) <-chan []*ui.Styled {
	ch := make(chan []*ui.Styled)
	go func() {
		result := callPrompt(ed, pu.promptFn)
		ch <- result
	}()
	return ch
}
