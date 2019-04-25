package newedit

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"sync"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
	"github.com/elves/elvish/newedit/clicore"
	"github.com/elves/elvish/newedit/prompt"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/util"
)

func makePrompt(nt notifier, ev *eval.Evaler, ns eval.Ns, computeInit eval.Callable, name string) clicore.Prompt {
	compute := computeInit
	ns[name] = vars.FromPtr(&compute)
	return prompt.New(func() styled.Text {
		return callPrompt(nt, ev, compute)
	})
}

var defaultPrompt, defaultRPrompt eval.Callable

func init() {
	user, userErr := user.Current()
	isRoot := userErr == nil && user.Uid == "0"

	defaultPrompt = getDefaultPrompt(isRoot)

	username := "???"
	if userErr == nil {
		username = user.Username
	}
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "???"
	}

	defaultRPrompt = getDefaultRPrompt(username, hostname)
}

func getDefaultPrompt(isRoot bool) eval.Callable {
	p := styled.Unstyled("> ")
	if isRoot {
		p = styled.Transform(styled.Unstyled("# "), "red")
	}
	return eval.NewGoFn("default prompt", func(fm *eval.Frame) {
		out := fm.OutputChan()
		out <- string(util.Getwd())
		out <- p
	})
}

func getDefaultRPrompt(username, hostname string) eval.Callable {
	rp := styled.Transform(styled.Unstyled(username+"@"+hostname), "inverse")
	return eval.NewGoFn("default rprompt", func(fm *eval.Frame) {
		fm.OutputChan() <- rp
	})
}

// callPrompt calls a function with no arguments and closed input, and converts
// its outputs to styled objects. Used to call prompt callbacks.
func callPrompt(nt notifier, ev *eval.Evaler, fn eval.Callable) styled.Text {
	ports := []*eval.Port{
		eval.DevNullClosedChan,
		{}, // Will be replaced when capturing output
		{File: os.Stderr},
	}

	return callForStyledText(nt, ev, fn, ports)
}

func callForStyledText(nt notifier, ev *eval.Evaler, fn eval.Callable, ports []*eval.Port) styled.Text {

	var (
		result      styled.Text
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
			result = newResult.(styled.Text)
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

	// XXX There is no source to pass to NewTopEvalCtx.
	fm := eval.NewTopFrame(ev, eval.NewInternalSource("[prompt]"), ports)
	err := fm.CallWithOutputCallback(fn, nil, eval.NoOpts, valuesCb, bytesCb)

	if err != nil {
		nt.Notify(fmt.Sprintf("prompt function error: %v", err))
		return nil
	}

	return result
}
