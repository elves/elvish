package edit

import (
	"io"
	"os"
	"os/user"
	"sync"
	"time"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/prompt"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
	"src.elv.sh/pkg/fsutil"
	"src.elv.sh/pkg/ui"
)

func initPrompts(appSpec *cli.AppSpec, nt notifier, ev *eval.Evaler, nb eval.NsBuilder) {
	promptVal, rpromptVal := getDefaultPromptVals()
	initPrompt(&appSpec.Prompt, "prompt", promptVal, nt, ev, nb)
	initPrompt(&appSpec.RPrompt, "rprompt", rpromptVal, nt, ev, nb)

	rpromptPersistentVar := newBoolVar(false)
	appSpec.RPromptPersistent = func() bool { return rpromptPersistentVar.Get().(bool) }
	nb.AddVar("rprompt-persistent", rpromptPersistentVar)
}

func initPrompt(p *cli.Prompt, name string, val eval.Callable, nt notifier, ev *eval.Evaler, nb eval.NsBuilder) {
	computeVar := vars.FromPtr(&val)
	nb.AddVar(name, computeVar)
	eagernessVar := newIntVar(5)
	nb.AddVar("-"+name+"-eagerness", eagernessVar)
	staleThresholdVar := newFloatVar(0.2)
	nb.AddVar(name+"-stale-threshold", staleThresholdVar)
	staleTransformVar := newFnVar(
		eval.NewGoFn("<default stale transform>", defaultStaleTransform))
	nb.AddVar(name+"-stale-transform", staleTransformVar)

	*p = prompt.New(prompt.Config{
		Compute: func() ui.Text {
			return callForStyledText(nt, ev, name, computeVar.Get().(eval.Callable))
		},
		Eagerness: func() int { return eagernessVar.GetRaw().(int) },
		StaleThreshold: func() time.Duration {
			seconds := staleThresholdVar.GetRaw().(float64)
			return time.Duration(seconds * float64(time.Second))
		},
		StaleTransform: func(original ui.Text) ui.Text {
			return callForStyledText(nt, ev, name+" stale transform", staleTransformVar.Get().(eval.Callable), original)
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
		return ui.Concat(ui.T(fsutil.Getwd()), p)
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

// Calls a function with the given arguments and closed input, and concatenates
// its outputs to a styled text. Used to call prompts and stale transformers.
func callForStyledText(nt notifier, ev *eval.Evaler, ctx string, fn eval.Callable, args ...any) ui.Text {
	var (
		result      ui.Text
		resultMutex sync.Mutex
	)
	add := func(v any) {
		resultMutex.Lock()
		defer resultMutex.Unlock()
		newResult, err := result.Concat(v)
		if err != nil {
			nt.notifyf("invalid output type from prompt: %s", vals.Kind(v))
		} else {
			result = newResult.(ui.Text)
		}
	}

	// Value outputs are concatenated.
	valuesCb := func(ch <-chan any) {
		for v := range ch {
			add(v)
		}
	}
	// Byte output is added to the prompt as a single unstyled text.
	bytesCb := func(r *os.File) {
		allBytes, err := io.ReadAll(r)
		if err != nil {
			nt.notifyf("error reading prompt byte output: %v", err)
		}
		if len(allBytes) > 0 {
			add(ui.ParseSGREscapedText(string(allBytes)))
		}
	}

	port1, done1, err := eval.PipePort(valuesCb, bytesCb)
	if err != nil {
		nt.notifyf("cannot create pipe for prompt: %v", err)
		return nil
	}
	port2, done2 := makeNotifyPort(nt)

	err = ev.Call(fn,
		eval.CallCfg{Args: args, From: "[" + ctx + "]"},
		eval.EvalCfg{Ports: []*eval.Port{nil, port1, port2}})
	done1()
	done2()

	if err != nil {
		nt.notifyError(ctx, err)
	}
	return result
}
