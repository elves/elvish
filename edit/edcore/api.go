package edcore

import (
	"errors"
	"strconv"
	"unicode/utf8"

	"github.com/elves/elvish/edit/eddefs"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vars"
)

// This file implements types and functions for interactions with the
// Elvishscript runtime.

var (
	errNotNav             = errors.New("not in navigation mode")
	errLineMustBeString   = errors.New("line must be string")
	errDotMustBeString    = errors.New("dot must be string")
	errDotMustBeInt       = errors.New("dot must be integer")
	errDotOutOfRange      = errors.New("dot out of range")
	errDotInsideCodepoint = errors.New("dot cannot be inside a codepoint")
	errEditorInactive     = errors.New("editor inactive")
)

// makeNs makes the edit: namespace.
func makeNs(ed *editor) eval.Ns {
	ns := eval.NewNs()

	// TODO(xiaq): Everything here should be registered to some registry instead
	// of centralized here.

	// Internal states.
	ns["current-command"] = vars.FromSetGet(
		func(v interface{}) error {
			if !ed.active {
				return errEditorInactive
			}
			if s, ok := v.(string); ok {
				ed.buffer = s
				ed.dot = len(ed.buffer)
			} else {
				return errLineMustBeString
			}
			return nil
		},
		func() interface{} { return ed.buffer },
	)
	ns["-dot"] = vars.FromSetGet(
		func(v interface{}) error {
			s, ok := v.(string)
			if !ok {
				return errDotMustBeString
			}
			i, err := strconv.Atoi(s)
			if err != nil {
				if err.(*strconv.NumError).Err == strconv.ErrRange {
					return errDotOutOfRange
				} else {
					return errDotMustBeInt
				}
			}
			if i < 0 || i > len(ed.buffer) {
				return errDotOutOfRange
			}
			if i < len(ed.buffer) {
				r, _ := utf8.DecodeRuneInString(ed.buffer[i:])
				if r == utf8.RuneError {
					return errDotInsideCodepoint
				}
			}
			ed.dot = i
			return nil
		},
		func() interface{} { return strconv.Itoa(ed.dot) },
	)
	ns["selected-file"] = vars.FromGet(
		func() interface{} {
			if !ed.active {
				throw(errEditorInactive)
			}
			nav, ok := ed.mode.(*navigation)
			if !ok {
				throw(errNotNav)
			}
			return nav.current.selectedName()
		},
	)

	// Functions.
	fns := map[string]interface{}{
		"binding-table": eddefs.MakeBindingMap,
		"insert-at-dot": ed.InsertAtDot,
		"replace-input": ed.replaceInput,
		"styled":        styled,
		"key":           ui.ToKey,
		"wordify":       wordifyBuiltin,
		"-dump-buf":     ed.dumpBuf,
	}
	ns.AddBuiltinFns("edit:", fns)

	return ns
}

// CallFn calls an Fn, displaying its outputs and possible errors as editor
// notifications. It is the preferred way to call a Fn while the editor is
// active.
func (ed *editor) CallFn(fn eval.Callable, args ...interface{}) {
	ports := []*eval.Port{
		eval.DevNullClosedChan, ed.notifyPort, ed.notifyPort,
	}
	// XXX There is no source to pass to NewTopEvalCtx.
	ec := eval.NewTopFrame(ed.evaler, eval.NewInternalSource("[editor]"), ports)
	ex := ec.Call(fn, args, eval.NoOpts)
	if ex != nil {
		ed.Notify("function error: %s", ex.Error())
	}

	// XXX Concurrency-dangerous!
	ed.refresh(true, true)
}
