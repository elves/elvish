package cliedit

import (
	"github.com/elves/elvish/cli"
	"github.com/elves/elvish/cli/el/codearea"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/eval/vars"
)

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

func initStateAPI(app cli.App, ns eval.Ns) {
	setDot := func(v interface{}) error {
		var dot int
		err := vals.ScanToGo(v, &dot)
		if err != nil {
			return err
		}
		app.CodeArea().MutateState(func(s *codearea.State) {
			s.CodeBuffer.Dot = dot
		})
		return nil
	}
	getDot := func() interface{} {
		return vals.FromGo(app.CodeArea().CopyState().CodeBuffer.Dot)
	}
	ns.Add("-dot", vars.FromSetGet(setDot, getDot))

	setContent := func(v interface{}) error {
		var content string
		err := vals.ScanToGo(v, &content)
		if err != nil {
			return err
		}
		app.CodeArea().MutateState(func(s *codearea.State) {
			s.CodeBuffer = codearea.CodeBuffer{Content: content, Dot: len(content)}
		})
		return nil
	}
	getContent := func() interface{} {
		return vals.FromGo(app.CodeArea().CopyState().CodeBuffer.Content)
	}
	ns.Add("current-command", vars.FromSetGet(setContent, getContent))
}
