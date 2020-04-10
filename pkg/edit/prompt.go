package edit

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"sync"
	"time"

	"github.com/elves/elvish/pkg/cli"
	"github.com/elves/elvish/pkg/cli/prompt"
	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/eval/vars"
	"github.com/elves/elvish/pkg/ui"
	"github.com/elves/elvish/pkg/util"
)

//elvdoc:var prompt
//
// See [Prompts](#prompts).

//elvdoc:var -prompt-eagerness
//
// See [Prompt Eagerness](#prompt-eagerness).

//elvdoc:var prompt-stale-threshold
//
// See [Stale Prompt](#stale-prompt).

//elvdoc:var prompt-stale-transformer.
//
// See [Stale Prompt](#stale-prompt).

//elvdoc:var rprompt
//
// See [Prompts](#prompts).

//elvdoc:var -rprompt-eagerness
//
// See [Prompt Eagerness](#prompt-eagerness).

//elvdoc:var rprompt-stale-threshold
//
// See [Stale Prompt](#stale-prompt).

//elvdoc:var rprompt-stale-transformer.
//
// See [Stale Prompt](#stale-prompt).

//elvdoc:var rprompt-persistent
//
// See [RPrompt Persistency](#rprompt-persistency).

func initPrompts(appSpec *cli.AppSpec, nt notifier, ev *eval.Evaler, ns eval.Ns) {
	promptVal, rpromptVal := getDefaultPromptVals()
	initPrompt(&appSpec.Prompt, "prompt", promptVal, nt, ev, ns)
	initPrompt(&appSpec.RPrompt, "rprompt", rpromptVal, nt, ev, ns)

	rpromptPersistentVar := newBoolVar(false)
	appSpec.RPromptPersistent = func() bool { return rpromptPersistentVar.Get().(bool) }
	ns["rprompt-persistent"] = rpromptPersistentVar
}

func initPrompt(p *cli.Prompt, name string, val eval.Callable, nt notifier, ev *eval.Evaler, ns eval.Ns) {
	computeVar := vars.FromPtr(&val)
	ns[name] = computeVar
	eagernessVar := newIntVar(5)
	ns["-"+name+"-eagerness"] = eagernessVar
	staleThresholdVar := newFloatVar(0.2)
	ns[name+"-stale-threshold"] = staleThresholdVar
	staleTransformVar := newFnVar(
		eval.NewGoFn("<default stale transform>", defaultStaleTransform))
	ns[name+"-stale-transform"] = staleTransformVar

	*p = prompt.New(prompt.Config{
		Compute: func() ui.Text {
			return callForStyledText(nt, ev, computeVar.Get().(eval.Callable))
		},
		Eagerness: func() int { return eagernessVar.GetRaw().(int) },
		StaleThreshold: func() time.Duration {
			seconds := staleThresholdVar.GetRaw().(float64)
			return time.Duration(seconds * float64(time.Second))
		},
		StaleTransform: func(original ui.Text) ui.Text {
			return callForStyledText(nt, ev, staleTransformVar.Get().(eval.Callable), original)
		},
	})
}

func getDefaultPromptVals() (prompt, rprompt eval.Callable) {
	user, userErr := user.Current()
	isRoot := userErr == nil && user.Uid == "0"

	username := "???"
	if userErr == nil {
		username = user.Username
	}
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "???"
	}

	return getDefaultPrompt(isRoot), getDefaultRPrompt(username, hostname)
}

func getDefaultPrompt(isRoot bool) eval.Callable {
	p := ui.T("> ")
	if isRoot {
		p = ui.T("# ", ui.FgRed)
	}
	return eval.NewGoFn("default prompt", func() ui.Text {
		return ui.T(util.Getwd()).ConcatText(p)
	})
}

func getDefaultRPrompt(username, hostname string) eval.Callable {
	rp := ui.T(username+"@"+hostname, ui.Inverse)
	return eval.NewGoFn("default rprompt", func() ui.Text {
		return rp
	})
}

func defaultStaleTransform(original ui.Text) ui.Text {
	return ui.StyleText(original, ui.Inverse)
}

// callPrompt calls a function with the given arguments and closed input, and
// concatenates its outputs to a styled text. Used to call prompts and stale
// transformers.
func callForStyledText(nt notifier, ev *eval.Evaler, fn eval.Callable, args ...interface{}) ui.Text {
	var (
		result      ui.Text
		resultMutex sync.Mutex
	)
	add := func(v interface{}) {
		resultMutex.Lock()
		defer resultMutex.Unlock()
		newResult, err := result.Concat(v)
		if err != nil {
			nt.Notify(fmt.Sprintf(
				"invalid output type from prompt: %s", vals.Kind(v)))
		} else {
			result = newResult.(ui.Text)
		}
	}

	// Value outputs are concatenated.
	valuesCb := func(ch <-chan interface{}) {
		for v := range ch {
			add(v)
		}
	}
	// Byte output is added to the prompt as a single unstyled text.
	bytesCb := func(r *os.File) {
		allBytes, err := ioutil.ReadAll(r)
		if err != nil {
			nt.Notify(fmt.Sprintf("error reading prompt byte output: %v", err))
		}
		if len(allBytes) > 0 {
			add(string(allBytes))
		}
	}

	ports := []*eval.Port{
		eval.DevNullClosedChan,
		{}, // Will be replaced when capturing output
		{File: os.Stderr},
	}
	// XXX There is no source to pass to NewTopEvalCtx.
	fm := eval.NewTopFrame(ev, eval.NewInternalGoSource("[prompt]"), ports)
	f := func(fm *eval.Frame) error { return fn.Call(fm, args, eval.NoOpts) }
	err := fm.PipeOutput(f, valuesCb, bytesCb)

	if err != nil {
		nt.Notify(fmt.Sprintf("prompt function error: %v", err))
		return nil
	}

	return result
}
