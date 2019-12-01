package cliedit

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"sync"
	"time"

	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/prompt"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/ui"
	"github.com/elves/elvish/util"
)

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
	p := ui.MakeText("> ")
	if isRoot {
		p = ui.TransformText(ui.MakeText("# "), "red")
	}
	return eval.NewGoFn("default prompt", func() ui.Text {
		return ui.MakeText(util.Getwd()).ConcatText(p)
	})
}

func getDefaultRPrompt(username, hostname string) eval.Callable {
	rp := ui.TransformText(ui.MakeText(username+"@"+hostname), "inverse")
	return eval.NewGoFn("default rprompt", func() ui.Text {
		return rp
	})
}

func defaultStaleTransform(original ui.Text) ui.Text {
	return ui.TransformText(original, "inverse")
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
	fm := eval.NewTopFrame(ev, eval.NewInternalSource("[prompt]"), ports)
	err := fm.CallWithOutputCallback(fn, args, eval.NoOpts, valuesCb, bytesCb)

	if err != nil {
		nt.Notify(fmt.Sprintf("prompt function error: %v", err))
		return nil
	}

	return result
}
