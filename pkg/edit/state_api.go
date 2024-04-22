package edit

import (
	"errors"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
)

var errDotOutOfBoundary = errors.New("dot out of command boundary")

func insertAtDot(app cli.App, text string) {
	codeArea, ok := focusedCodeArea(app)
	if !ok {
		return
	}
	codeArea.MutateState(func(s *tk.CodeAreaState) {
		s.Buffer.InsertAtDot(text)
	})
}

func replaceInput(app cli.App, text string) {
	codeArea, ok := focusedCodeArea(app)
	if !ok {
		return
	}
	codeArea.MutateState(func(s *tk.CodeAreaState) {
		s.Buffer = tk.CodeBuffer{Content: text, Dot: len(text)}
	})
}

func initStateAPI(app cli.App, nb eval.NsBuilder) {
	// State API always operates on the root CodeArea widget
	codeArea := app.ActiveWidget().(tk.CodeArea)

	nb.AddGoFns(map[string]any{
		"insert-at-dot": func(s string) { insertAtDot(app, s) },
		"replace-input": func(s string) { replaceInput(app, s) },
	})

	setDot := func(v any) error {
		var dot int
		err := vals.ScanToGo(v, &dot)
		if err != nil {
			return err
		}
		codeArea.MutateState(func(s *tk.CodeAreaState) {
			if dot < 0 || dot > len(s.Buffer.Content) {
				err = errDotOutOfBoundary
			} else {
				s.Buffer.Dot = dot
			}
		})
		return err
	}
	getDot := func() any {
		return vals.FromGo(codeArea.CopyState().Buffer.Dot)
	}
	nb.AddVar("-dot", vars.FromSetGet(setDot, getDot))

	setCurrentCommand := func(v any) error {
		var content string
		err := vals.ScanToGo(v, &content)
		if err != nil {
			return err
		}
		replaceInput(app, content)
		return nil
	}
	getCurrentCommand := func() any {
		return vals.FromGo(codeArea.CopyState().Buffer.Content)
	}
	nb.AddVar("current-command", vars.FromSetGet(setCurrentCommand, getCurrentCommand))
}
