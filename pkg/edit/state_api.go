package edit

import (
	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/eval/vars"
)

//elvdoc:fn insert-at-dot
//
// ```elvish
// edit:insert-at-dot $text
// ```
//
// Inserts the given text at the dot, moving the dot after the newly
// inserted text.

func insertAtDot(app cli.App, text string) {
	codeArea, ok := focusedCodeArea(app)
	if !ok {
		return
	}
	codeArea.MutateState(func(s *tk.CodeAreaState) {
		s.Buffer.InsertAtDot(text)
	})
}

//elvdoc:fn replace-input
//
// ```elvish
// edit:replace-input $text
// ```
//
// Equivalent to assigning `$text` to `$edit:current-command`.

func replaceInput(app cli.App, text string) {
	codeArea, ok := focusedCodeArea(app)
	if !ok {
		return
	}
	codeArea.MutateState(func(s *tk.CodeAreaState) {
		s.Buffer = tk.CodeBuffer{Content: text, Dot: len(text)}
	})
}

//elvdoc:var -dot
//
// Contains the current position of the cursor, as a byte position within
// `$edit:current-command`.

//elvdoc:var current-command
//
// Contains the content of the current input. Setting the variable will
// cause the cursor to move to the very end, as if `edit-dot = (count
// $edit:current-command)` has been invoked.
//
// This API is subject to change.

func initStateAPI(app cli.App, nb eval.NsBuilder) {
	// State API always operates on the root CodeArea widget
	codeArea := app.ActiveWidget().(tk.CodeArea)

	nb.AddGoFns(map[string]interface{}{
		"insert-at-dot": func(s string) { insertAtDot(app, s) },
		"replace-input": func(s string) { replaceInput(app, s) },
	})

	setDot := func(v interface{}) error {
		var dot int
		err := vals.ScanToGo(v, &dot)
		if err != nil {
			return err
		}
		codeArea.MutateState(func(s *tk.CodeAreaState) {
			s.Buffer.Dot = dot
		})
		return nil
	}
	getDot := func() interface{} {
		return vals.FromGo(codeArea.CopyState().Buffer.Dot)
	}
	nb.AddVar("-dot", vars.FromSetGet(setDot, getDot))

	setCurrentCommand := func(v interface{}) error {
		var content string
		err := vals.ScanToGo(v, &content)
		if err != nil {
			return err
		}
		replaceInput(app, content)
		return nil
	}
	getCurrentCommand := func() interface{} {
		return vals.FromGo(codeArea.CopyState().Buffer.Content)
	}
	nb.AddVar("current-command", vars.FromSetGet(setCurrentCommand, getCurrentCommand))
}
