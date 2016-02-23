package edit

import (
	"bytes"
	"errors"
	"os"
	"os/user"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/util"
)

var ErrPromptMustBeStringOrFunc = errors.New("prompt must be string or function")

// PromptVariable implements $le:prompt and $le:rprompt.
type PromptVariable struct {
	Prompt *Prompt
}

func (pv PromptVariable) Get() eval.Value {
	// XXX Should return a proper eval.Caller
	return eval.String("<prompt>")
}

func (pv PromptVariable) Set(v eval.Value) {
	if s, ok := v.(eval.String); ok {
		*pv.Prompt = BuiltinPrompt(func(*Editor) string { return string(s) })
	} else if c, ok := v.(eval.Caller); ok {
		*pv.Prompt = CallerPrompt{c}
	} else {
		throw(ErrPromptMustBeStringOrFunc)
	}
}

// Prompt is the interface of prompt functions.
type Prompt interface {
	Call(*Editor) string
}

// BuiltinPrompt is a trivial implementation of Prompt.
type BuiltinPrompt func(*Editor) string

func (bp BuiltinPrompt) Call(ed *Editor) string {
	return bp(ed)
}

// CallerPrompt adapts a eval.Caller to a Prompt.
type CallerPrompt struct {
	eval.Caller
}

func (c CallerPrompt) Call(ed *Editor) string {
	in, err := makeClosedStdin()
	if err != nil {
		return ""
	}
	ports := []*eval.Port{in, &eval.Port{File: os.Stdout}, &eval.Port{File: os.Stderr}}

	// XXX There is no source to pass to NewTopEvalCtx.
	ec := eval.NewTopEvalCtx(ed.evaler, "[editor prompt]", "", ports)
	values, err := ec.PCaptureOutput(c.Caller, nil)
	if err != nil {
		ed.notify("prompt function error: %v", err)
		return ""
	}
	var b bytes.Buffer
	for _, v := range values {
		b.WriteString(eval.ToString(v))
	}
	return b.String()
}

func defaultPrompts() (Prompt, Prompt) {
	// Make default prompts.
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
	prompt := func(*Editor) string {
		return util.Getwd() + "> "
	}
	rprompt := func(*Editor) string {
		return rpromptStr
	}
	return BuiltinPrompt(prompt), BuiltinPrompt(rprompt)
}
