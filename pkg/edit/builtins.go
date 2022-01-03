package edit

import (
	"errors"

	"src.elv.sh/pkg/cli"
	"src.elv.sh/pkg/cli/modes"
	"src.elv.sh/pkg/cli/term"
	"src.elv.sh/pkg/cli/tk"
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/errs"
	"src.elv.sh/pkg/eval/vals"
	"src.elv.sh/pkg/parse"
	"src.elv.sh/pkg/parse/parseutil"
	"src.elv.sh/pkg/ui"
)

//elvdoc:fn binding-table
//
// Converts a normal map into a binding map.

//elvdoc:fn -dump-buf
//
// Dumps the current UI buffer as HTML. This command is used to generate
// "ttyshots" on the [website](https://elv.sh).
//
// Example:
//
// ```elvish
// set edit:global-binding[Ctrl-X] = { print (edit:-dump-buf) > ~/a.html }
// ```

func dumpBuf(tty cli.TTY) string {
	return bufToHTML(tty.Buffer())
}

//elvdoc:fn close-mode
//
// Closes the current active mode.

func closeMode(app cli.App) {
	app.PopAddon()
}

//elvdoc:fn end-of-history
//
// Adds a notification saying "End of history".

func endOfHistory(app cli.App) {
	app.Notify(ui.T("End of history"))
}

//elvdoc:fn redraw
//
// ```elvish
// edit:redraw &full=$false
// ```
//
// Triggers a redraw.
//
// The `&full` option controls whether to do a full redraw. By default, all
// redraws performed by the line editor are incremental redraws, updating only
// the part of the screen that has changed from the last redraw. A full redraw
// updates the entire command line.

type redrawOpts struct{ Full bool }

func (redrawOpts) SetDefaultOptions() {}

func redraw(app cli.App, opts redrawOpts) {
	if opts.Full {
		app.RedrawFull()
	} else {
		app.Redraw()
	}
}

//elvdoc:fn clear
//
// ```elvish
// edit:clear
// ```
//
// Clears the screen.
//
// This command should be used in place of the external `clear` command to clear
// the screen.

func clear(app cli.App, tty cli.TTY) {
	tty.HideCursor()
	tty.ClearScreen()
	app.RedrawFull()
	tty.ShowCursor()
}

//elvdoc:fn insert-raw
//
// Requests the next terminal input to be inserted uninterpreted.

func insertRaw(app cli.App, tty cli.TTY) {
	codeArea, ok := focusedCodeArea(app)
	if !ok {
		return
	}
	tty.SetRawInput(1)
	w := modes.NewStub(modes.StubSpec{
		Bindings: tk.FuncBindings(func(w tk.Widget, event term.Event) bool {
			switch event := event.(type) {
			case term.KeyEvent:
				codeArea.MutateState(func(s *tk.CodeAreaState) {
					s.Buffer.InsertAtDot(string(event.Rune))
				})
				app.PopAddon()
				return true
			default:
				return false
			}
		}),
		Name: " RAW ",
	})
	app.PushAddon(w)
}

//elvdoc:fn key
//
// ```elvish
// edit:key $string
// ```
//
// Parses a string into a key.

var errMustBeKeyOrString = errors.New("must be key or string")

func toKey(v interface{}) (ui.Key, error) {
	switch v := v.(type) {
	case ui.Key:
		return v, nil
	case string:
		return ui.ParseKey(v)
	default:
		return ui.Key{}, errMustBeKeyOrString
	}
}

//elvdoc:fn notify
//
// ```elvish
// edit:notify $message
// ```
//
// Prints a notification message. The argument may be a string or a [styled
// text](builtin.html#styled).
//
// If called while the editor is active, this will print the message above the
// editor, and redraw the editor.
//
// If called while the editor is inactive, the message will be queued, and shown
// once the editor becomes active.

func notify(app cli.App, x interface{}) error {
	// TODO: De-duplicate with the implementation of the styled builtin.
	var t ui.Text
	switch x := x.(type) {
	case string:
		t = ui.T(x)
	case ui.Text:
		t = x.Clone()
	default:
		return errs.BadValue{What: "argument to edit:notify",
			Valid: "string, styled segment or styled text", Actual: vals.Kind(x)}
	}
	app.Notify(t)
	return nil
}

//elvdoc:fn return-line
//
// Causes the Elvish REPL to end the current read iteration and evaluate the
// code it just read. If called from a key binding, takes effect after the key
// binding returns.

//elvdoc:fn return-eof
//
// Causes the Elvish REPL to terminate. If called from a key binding, takes
// effect after the key binding returns.

//elvdoc:fn smart-enter
//
// Inserts a literal newline if the current code is not syntactically complete
// Elvish code. Accepts the current line otherwise.

func smartEnter(app cli.App) {
	codeArea, ok := focusedCodeArea(app)
	if !ok {
		return
	}
	commit := false
	codeArea.MutateState(func(s *tk.CodeAreaState) {
		buf := &s.Buffer
		if isSyntaxComplete(buf.Content) {
			commit = true
		} else {
			buf.InsertAtDot("\n")
		}
	})
	if commit {
		app.CommitCode()
	}
}

func isSyntaxComplete(code string) bool {
	_, err := parse.Parse(parse.Source{Code: code}, parse.Config{})
	if err != nil {
		for _, e := range err.(*parse.Error).Entries {
			if e.Context.From == len(code) {
				return false
			}
		}
	}
	return true
}

//elvdoc:fn wordify
//
//
// ```elvish
// edit:wordify $code
// ```
// Breaks Elvish code into words.

func wordify(fm *eval.Frame, code string) error {
	out := fm.ValueOutput()
	for _, s := range parseutil.Wordify(code) {
		err := out.Put(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func initTTYBuiltins(app cli.App, tty cli.TTY, nb eval.NsBuilder) {
	nb.AddGoFns(map[string]interface{}{
		"-dump-buf":  func() string { return dumpBuf(tty) },
		"insert-raw": func() { insertRaw(app, tty) },
		"clear":      func() { clear(app, tty) },
	})
}

func initMiscBuiltins(app cli.App, nb eval.NsBuilder) {
	nb.AddGoFns(map[string]interface{}{
		"binding-table":  makeBindingMap,
		"close-mode":     func() { closeMode(app) },
		"end-of-history": func() { endOfHistory(app) },
		"key":            toKey,
		"notify":         func(x interface{}) error { return notify(app, x) },
		"redraw":         func(opts redrawOpts) { redraw(app, opts) },
		"return-line":    app.CommitCode,
		"return-eof":     app.CommitEOF,
		"smart-enter":    func() { smartEnter(app) },
		"wordify":        wordify,
	})
}

// Like mode.FocusedCodeArea, but handles the error by writing a notification.
func focusedCodeArea(app cli.App) (tk.CodeArea, bool) {
	codeArea, err := modes.FocusedCodeArea(app)
	if err != nil {
		app.Notify(modes.ErrorText(err))
		return nil, false
	}
	return codeArea, true
}
