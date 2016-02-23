package edit

import (
	"bytes"
	"os"

	"github.com/elves/elvish/eval"
)

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
	ports := []*eval.Port{in, nil, &eval.Port{File: os.Stderr}}

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
