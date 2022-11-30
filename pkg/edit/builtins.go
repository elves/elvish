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

func closeMode(app cli.App) {
	app.PopAddon()
}

func endOfHistory(app cli.App) {
	app.Notify(ui.T("End of history"))
}

type redrawOpts struct{ Full bool }

func (redrawOpts) SetDefaultOptions() {}

func redraw(app cli.App, opts redrawOpts) {
	if opts.Full {
		app.RedrawFull()
	} else {
		app.Redraw()
	}
}

func clear(app cli.App, tty cli.TTY) {
	tty.HideCursor()
	tty.ClearScreen()
	app.RedrawFull()
	tty.ShowCursor()
}

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

var errMustBeKeyOrString = errors.New("must be key or string")

func toKey(v any) (ui.Key, error) {
	switch v := v.(type) {
	case ui.Key:
		return v, nil
	case string:
		return ui.ParseKey(v)
	default:
		return ui.Key{}, errMustBeKeyOrString
	}
}

func notify(app cli.App, x any) error {
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
	_, err := parse.Parse(parse.Source{Name: "[syntax check]", Code: code}, parse.Config{})
	for _, e := range parse.UnpackErrors(err) {
		if e.Context.From == len(code) {
			return false
		}
	}
	return true
}

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
	nb.AddGoFns(map[string]any{
		"insert-raw": func() { insertRaw(app, tty) },
		"clear":      func() { clear(app, tty) },
	})
}

func initMiscBuiltins(app cli.App, nb eval.NsBuilder) {
	nb.AddGoFns(map[string]any{
		"binding-table":  makeBindingMap,
		"close-mode":     func() { closeMode(app) },
		"end-of-history": func() { endOfHistory(app) },
		"key":            toKey,
		"notify":         func(x any) error { return notify(app, x) },
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
