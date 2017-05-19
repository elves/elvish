package edit

import (
	"errors"
	"fmt"

	"github.com/elves/elvish/edit/uitypes"
	"github.com/elves/elvish/eval"
)

// Line editor builtins.

var ErrEditorInactive = errors.New("editor inactive")

// BuiltinFn records an editor builtin.
type BuiltinFn struct {
	name string
	impl func(ed *Editor)
}

var _ eval.CallableValue = &BuiltinFn{}

func (*BuiltinFn) Kind() string {
	return "fn"
}

func (bf *BuiltinFn) Repr(int) string {
	return "$le:&" + bf.name
}

func (bf *BuiltinFn) Call(ec *eval.EvalCtx, args []eval.Value, opts map[string]eval.Value) {
	eval.TakeNoOpt(opts)
	eval.TakeNoArg(args)
	ed, ok := ec.Editor.(*Editor)
	if !ok || !ed.active {
		throw(ErrEditorInactive)
	}
	bf.impl(ed)
}

var builtinImpls = map[string]func(*Editor){
	// Command and insert mode
	"kill-line-left":       killLineLeft,
	"kill-line-right":      killLineRight,
	"kill-word-left":       killWordLeft,
	"kill-small-word-left": killSmallWordLeft,
	"kill-rune-left":       killRuneLeft,
	"kill-rune-right":      killRuneRight,
	"move-dot-left":        moveDotLeft,
	"move-dot-right":       moveDotRight,
	"move-dot-left-word":   moveDotLeftWord,
	"move-dot-right-word":  moveDotRightWord,
	"move-dot-sol":         moveDotSOL,
	"move-dot-eol":         moveDotEOL,
	"move-dot-up":          moveDotUp,
	"move-dot-down":        moveDotDown,
	"insert-last-word":     insertLastWord,
	"insert-key":           insertKey,
	"return-line":          returnLine,
	"smart-enter":          smartEnter,
	"return-eof":           returnEOF,
	"toggle-quote-paste":   toggleQuotePaste,
	"end-of-history":       endOfHistory,
	"insert-raw":           startInsertRaw,

	// Misc
	"redraw": redraw,
}

var defaultBindings = map[ModeType]map[uitypes.Key]string{
	modeInsert:         insertKeyBindings,
	modeCommand:        commandKeyBindings,
	modeCompletion:     complKeyBindings,
	modeNavigation:     navKeyBindings,
	modeHistory:        historyKeyBindings,
	modeHistoryListing: histlistKeyBindings,
	modeBang:           bangKeyBindings,
	modeLocation:       locKeyBindings,
}

var (
	builtinMap  = map[string]*BuiltinFn{}
	keyBindings = map[ModeType]map[uitypes.Key]eval.CallableValue{}
)

func addBuiltinFns(prefix string, implMap map[string]func(*Editor)) {
	for name, impl := range implMap {
		fullName := prefix + name
		builtinMap[fullName] = &BuiltinFn{fullName, impl}
	}
}

func init() {
	addBuiltinFns("", builtinImpls)
	addBuiltinFns("insert-", insertBuiltinImpls)
	addBuiltinFns("command-", commandBuiltinImpls)
	addBuiltinFns("nav-", navBuiltinImpls)
	addBuiltinFns("loc-", locBuiltinImpls)
	addBuiltinFns("bang-", bangBuiltinImpls)
	addBuiltinFns("compl-", complBuiltinImpls)
	addBuiltinFns("history-", historyBuiltinImpls)
	addBuiltinFns("histlist-", histlistBuiltinImpls)

	addListingBuiltins("loc-", func(ed *Editor) *listing { return &ed.location.listing })
	addListingDefaultBindings("loc-", modeLocation)
	addListingBuiltins("histlist-", func(ed *Editor) *listing { return &ed.histlist.listing })
	addListingDefaultBindings("histlist-", modeHistoryListing)
	addListingBuiltins("bang-", func(ed *Editor) *listing { return &ed.bang.listing })
	addListingDefaultBindings("bang-", modeBang)

	for mode, table := range defaultBindings {
		keyBindings[mode] = map[uitypes.Key]eval.CallableValue{}
		for key, name := range table {
			fn, ok := builtinMap[name]
			if !ok {
				fmt.Println("bad name " + name)
			} else {
				keyBindings[mode][key] = fn
			}
		}
	}
}

func redraw(ed *Editor) {
	ed.refresh(true, true)
}

func endOfHistory(ed *Editor) {
	ed.Notify("End of history")
}
