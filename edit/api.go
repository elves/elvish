package edit

import (
	"errors"
	"strconv"
	"unicode/utf8"
	"unsafe"

	. "github.com/elves/elvish/edit/edtypes"

	"github.com/elves/elvish/edit/ui"
	"github.com/elves/elvish/eval"
	"github.com/elves/elvish/eval/vartypes"
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

// BuiltinFn represents an editor builtin.
type BuiltinFn struct {
	name string
	impl func(ed *Editor)
}

var _ eval.Callable = &BuiltinFn{}

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
func (bf *BuiltinFn) Call(ec *eval.Frame, args []interface{}, opts map[string]interface{}) error {
	eval.TakeNoOpt(opts)
	eval.TakeNoArg(args)
	ed, ok := ec.Editor.(*Editor)
	if !ok {
		return errEditorInvalid
	}
	if !ed.active {
		return errEditorInactive
	}
	bf.impl(ed)
	return nil
}

// makeNs makes the edit: namespace.
func makeNs(ed *Editor) eval.Ns {
	ns := eval.NewNs()

	// TODO(xiaq): Everything here should be registered to some registry instead
	// of centralized here.

	// Internal states.
	ns["current-command"] = vartypes.NewCallback(
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
	ns["-dot"] = vartypes.NewCallback(
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
	ns["selected-file"] = vartypes.NewRoCallback(
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

	// Completers.
	for _, bac := range argCompletersData {
		ns[bac.name+eval.FnSuffix] = vartypes.NewRo(bac)
	}

	// Matchers.
	ns.AddFn("match-prefix", matchPrefix)
	ns.AddFn("match-substr", matchSubstr)
	ns.AddFn("match-subseq", matchSubseq)

	// Functions.
	fns := map[string]interface{}{
		"binding-table":     MakeBindingMap,
		"command-history":   CommandHistory,
		"complete-getopt":   complGetopt,
		"complex-candidate": makeComplexCandidate,
		"insert-at-dot":     ed.insertAtDot,
		"replace-input":     ed.replaceInput,
		"styled":            styled,
		"key":               ui.ToKey,
		"wordify":           wordifyBuiltin,
		"-dump-buf":         ed.dumpBuf,
		"-narrow-read":      NarrowRead,
	}
	ns.AddBuiltinFns("edit:", fns)

	ns.AddNs("listing", initModeAPI("listing:", listingFns, &ed.listingBinding))
	ns.AddNs("narrow", initModeAPI("narrow:", narrowFns, &ed.narrowBinding))
	ns.AddNs("histlist", initModeAPI("histlist:", histlistFns, &ed.histlistBinding))
	ns.AddNs("lastcmd", initModeAPI("lastcmd:", lastcmdFns, &ed.lastcmdBinding))
	ns.AddNs("location", initModeAPI("location:", locationFns, &ed.locationBinding))

	return ns
}

// CallFn calls an Fn, displaying its outputs and possible errors as editor
// notifications. It is the preferred way to call a Fn while the editor is
// active.
func (ed *Editor) CallFn(fn eval.Callable, args ...interface{}) {
	if b, ok := fn.(*BuiltinFn); ok {
		// Builtin function: quick path.
		b.impl(ed)
		return
	}

	ports := []*eval.Port{
		eval.DevNullClosedChan, ed.notifyPort, ed.notifyPort,
	}
	// XXX There is no source to pass to NewTopEvalCtx.
	ec := eval.NewTopFrame(ed.evaler, eval.NewInternalSource("[editor]"), ports)
	ex := ec.PCall(fn, args, eval.NoOpts)
	if ex != nil {
		ed.Notify("function error: %s", ex.Error())
	}

	// XXX Concurrency-dangerous!
	ed.refresh(true, true)
}
