package edit

import (
	"errors"
	"os"
	"os/user"

	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/util"
)

var ErrPromptMustBeStringStyledOrFunc = errors.New("prompt must be string, styled or function")

// PromptVariable is a prompt function variable. It may be set to a String, a
// Fn, or a BuiltinPrompt. It provides $le:prompt and $le:rprompt.
type PromptVariable struct {
	Prompt *Prompt
}

func (pv PromptVariable) Get() eval.Value {
	// XXX Should return a proper eval.Fn
	return eval.String("<prompt>")
}

func (pv PromptVariable) Set(v eval.Value) {
	if s, ok := v.(eval.String); ok {
		*pv.Prompt = BuiltinPrompt(func(*Editor) []*styled {
			return []*styled{&styled{string(s), ""}}
		})
	} else if s, ok := v.(*styled); ok {
		*pv.Prompt = BuiltinPrompt(func(*Editor) []*styled {
			return []*styled{s}
		})
	} else if c, ok := v.(eval.Fn); ok {
		*pv.Prompt = FnAsPrompt{c}
	} else {
		throw(ErrPromptMustBeStringStyledOrFunc)
	}
}

// Prompt is the interface of prompt functions.
type Prompt interface {
	Call(*Editor) []*styled
}

// BuiltinPrompt is a trivial implementation of Prompt.
type BuiltinPrompt func(*Editor) []*styled

func (bp BuiltinPrompt) Call(ed *Editor) []*styled {
	return bp(ed)
}

// FnAsPrompt adapts a eval.Fn to a Prompt.
type FnAsPrompt struct {
	eval.Fn
}

func (c FnAsPrompt) Call(ed *Editor) []*styled {
	in, err := makeClosedStdin()
	if err != nil {
		return nil
	}
	ports := []*eval.Port{in, &eval.Port{File: os.Stdout}, &eval.Port{File: os.Stderr}}

	// XXX There is no source to pass to NewTopEvalCtx.
	ec := eval.NewTopEvalCtx(ed.evaler, "[editor prompt]", "", ports)
	values, err := ec.PCaptureOutput(c.Fn, nil)
	if err != nil {
		ed.notify("prompt function error: %v", err)
		return nil
	}

	var ss []*styled
	for _, v := range values {
		if s, ok := v.(*styled); ok {
			ss = append(ss, s)
		} else {
			ss = append(ss, &styled{eval.ToString(v), ""})
		}
	}
	return ss
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
	prompt := func(*Editor) []*styled {
		return []*styled{&styled{util.Getwd() + "> ", ""}}
	}
	rprompt := func(*Editor) []*styled {
		return []*styled{&styled{rpromptStr, "7"}}
	}
	return BuiltinPrompt(prompt), BuiltinPrompt(rprompt)
}
