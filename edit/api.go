package edit

import (
	"errors"
	"strconv"
	"unicode/utf8"
	"unsafe"

	"github.com/elves/elvish/edit/history"
	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/xiaq/persistent/hash"
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
	errEditorInvalid      = errors.New("internal error: editor not set up")
	errEditorInactive     = errors.New("editor inactive")
)

// BuiltinFn records an editor builtin.
type BuiltinFn struct {
	name string
	impl func(ed *Editor)
}

var _ eval.CallableValue = &BuiltinFn{}

// Kind returns "fn".
func (*BuiltinFn) Kind() string {
	return "fn"
}

// Equal compares based on identity.
func (bf *BuiltinFn) Equal(a interface{}) bool {
	return bf == a
}

func (bf *BuiltinFn) Hash() uint32 {
	return hash.Pointer(unsafe.Pointer(bf))
}

// Repr returns the representation of a builtin function as a variable name.
func (bf *BuiltinFn) Repr(int) string {
	return "$" + bf.name
}

// Call calls a builtin function.
func (bf *BuiltinFn) Call(ec *eval.EvalCtx, args []eval.Value, opts map[string]eval.Value) {
	eval.TakeNoOpt(opts)
	eval.TakeNoArg(args)
	ed, ok := ec.Editor.(*Editor)
	if !ok {
		throw(errEditorInvalid)
	}
	if !ed.active {
		throw(errEditorInactive)
	}
	bf.impl(ed)
}

// installModules installs edit: and edit:* modules.
func installModules(modules map[string]eval.Namespace, ed *Editor) {
	// Construct the edit: module, starting with builtins.
	ns := makeNamespaceFromBuiltins(builtinMaps[""])

	// TODO(xiaq): Everything here should be registered to some registry instead
	// of centralized here.

	// Editor configurations.
	for name, variable := range ed.variables {
		ns[name] = variable
	}

	// Internal states.
	ns["history"] = eval.NewRoVariable(history.List{&ed.historyMutex, ed.daemon})
	ns["current-command"] = eval.MakeVariableFromCallback(
		func(v eval.Value) {
			if !ed.active {
				throw(errEditorInactive)
			}
			if s, ok := v.(eval.String); ok {
				ed.buffer = string(s)
				ed.dot = len(ed.buffer)
			} else {
				throw(errLineMustBeString)
			}
		},
		func() eval.Value { return eval.String(ed.buffer) },
	)
	ns["-dot"] = eval.MakeVariableFromCallback(
		func(v eval.Value) {
			s, ok := v.(eval.String)
			if !ok {
				throw(errDotMustBeString)
			}
			i, err := strconv.Atoi(string(s))
			if err != nil {
				if err.(*strconv.NumError).Err == strconv.ErrRange {
					throw(errDotOutOfRange)
				} else {
					throw(errDotMustBeInt)
				}
			}
			if i < 0 || i > len(ed.buffer) {
				throw(errDotOutOfRange)
			}
			if i < len(ed.buffer) {
				r, _ := utf8.DecodeRuneInString(ed.buffer[i:])
				if r == utf8.RuneError {
					throw(errDotInsideCodepoint)
				}
			}
			ed.dot = i
		},
		func() eval.Value { return eval.String(strconv.Itoa(ed.dot)) },
	)
	ns["selected-file"] = eval.MakeRoVariableFromCallback(
		func() eval.Value {
			if !ed.active {
				throw(errEditorInactive)
			}
			nav, ok := ed.mode.(*navigation)
			if !ok {
				throw(errNotNav)
			}
			return eval.String(nav.current.selectedName())
		},
	)

	// Completers.
	for _, bac := range argCompletersData {
		ns[eval.FnPrefix+bac.name] = eval.NewRoVariable(bac)
	}

	// Matchers.
	eval.AddBuiltinFns(ns, matchers...)

	// Functions.
	eval.AddBuiltinFns(ns,
		&eval.BuiltinFn{"edit:command-history", CommandHistory},
		&eval.BuiltinFn{"edit:complete-getopt", complGetopt},
		&eval.BuiltinFn{"edit:complex-candidate", outputComplexCandidate},
		&eval.BuiltinFn{"edit:insert-at-dot", InsertAtDot},
		&eval.BuiltinFn{"edit:replace-input", ReplaceInput},
		&eval.BuiltinFn{"edit:styled", styled},
		&eval.BuiltinFn{"edit:key", ui.KeyBuiltin},
		&eval.BuiltinFn{"edit:wordify", Wordify},
		&eval.BuiltinFn{"edit:-dump-buf", _dumpBuf},
		&eval.BuiltinFn{"edit:-narrow-read", NarrowRead},
	)

	modules["edit"] = ns
	// Install other modules.
	for module, builtins := range builtinMaps {
		if module != "" {
			modules["edit:"+module] = makeNamespaceFromBuiltins(builtins)
		}
	}

	// Add $edit:{mode}:binding variables.
	for mode, bindingVar := range ed.bindings {
		if modules["edit:"+mode] == nil {
			modules["edit:"+mode] = make(eval.Namespace)
		}
		modules["edit:"+mode]["binding"] = bindingVar
	}
}

// CallFn calls an Fn, displaying its outputs and possible errors as editor
// notifications. It is the preferred way to call a Fn while the editor is
// active.
func (ed *Editor) CallFn(fn eval.CallableValue, args ...eval.Value) {
	if b, ok := fn.(*BuiltinFn); ok {
		// Builtin function: quick path.
		b.impl(ed)
		return
	}

	ports := []*eval.Port{
		eval.DevNullClosedChan, ed.notifyPort, ed.notifyPort,
	}
	// XXX There is no source to pass to NewTopEvalCtx.
	ec := eval.NewTopEvalCtx(ed.evaler, "[editor]", "", ports)
	ex := ec.PCall(fn, args, eval.NoOpts)
	if ex != nil {
		ed.Notify("function error: %s", ex.Error())
	}

	ed.refresh(true, true)
}
