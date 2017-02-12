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

var _ eval.FnValue = &BuiltinFn{}

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

var builtins = []*BuiltinFn{
	// Command and insert mode
	{"start-insert", startInsert},
	{"start-command", startCommand},
	{"kill-line-left", killLineLeft},
	{"kill-line-right", killLineRight},
	{"kill-word-left", killWordLeft},
	{"kill-small-word-left", killSmallWordLeft},
	{"kill-rune-left", killRuneLeft},
	{"kill-rune-right", killRuneRight},
	{"move-dot-left", moveDotLeft},
	{"move-dot-right", moveDotRight},
	{"move-dot-left-word", moveDotLeftWord},
	{"move-dot-right-word", moveDotRightWord},
	{"move-dot-sol", moveDotSOL},
	{"move-dot-eol", moveDotEOL},
	{"move-dot-up", moveDotUp},
	{"move-dot-down", moveDotDown},
	{"insert-last-word", insertLastWord},
	{"insert-key", insertKey},
	{"return-line", returnLine},
	{"smart-enter", smartEnter},
	{"return-eof", returnEOF},
	{"toggle-quote-paste", toggleQuotePaste},
	{"end-of-history", endOfHistory},
	{"default-command", defaultCommand},
	{"insert-default", defaultInsert},

	// Completion mode
	{"compl-prefix-or-start-compl", complPrefixOrStartCompl},
	{"start-compl", startCompl},
	{"compl-up", complUp},
	{"compl-down", complDown},
	{"compl-down-cycle", complDownCycle},
	{"compl-left", complLeft},
	{"compl-right", complRight},
	{"compl-accept", complAccept},
	{"compl-trigger-filter", complTriggerFilter},
	{"compl-default", complDefault},

	// Navigation mode
	{"start-nav", startNav},
	{"nav-up", navUp},
	{"nav-down", navDown},
	{"nav-page-up", navPageUp},
	{"nav-page-down", navPageDown},
	{"nav-left", navLeft},
	{"nav-right", navRight},
	{"nav-trigger-shown-hidden", navTriggerShowHidden},
	{"nav-trigger-filter", navTriggerFilter},
	{"nav-insert-selected", navInsertSelected},
	{"nav-insert-selected-and-quit", navInsertSelectedAndQuit},
	{"navigation-default", navigationDefault},

	// History mode
	{"start-history", startHistory},
	{"history-up", historyUp},
	{"history-down", historyDown},
	{"history-down-or-quit", historyDownOrQuit},
	{"history-switch-to-histlist", historySwitchToHistlist},
	{"history-default", historyDefault},

	// History listing mode
	{"start-histlist", startHistlist},
	{"histlist-toggle-dedup", histlistToggleDedup},
	{"histlist-toggle-case-sensitivity", histlistToggleCaseSensitivity},

	// Bang mode
	{"start-bang", startBang},
	{"bang-alt-default", bangAltDefault},

	// Location mode
	{"start-location", startLocation},

	// Misc
	{"redraw", redraw},
}

var defaultBindings = map[ModeType]map[uitypes.Key]string{
	modeInsert: map[uitypes.Key]string{
		// Moving.
		uitypes.Key{uitypes.Left, 0}:             "move-dot-left",
		uitypes.Key{uitypes.Right, 0}:            "move-dot-right",
		uitypes.Key{uitypes.Up, uitypes.Alt}:     "move-dot-up",
		uitypes.Key{uitypes.Down, uitypes.Alt}:   "move-dot-down",
		uitypes.Key{uitypes.Left, uitypes.Ctrl}:  "move-dot-left-word",
		uitypes.Key{uitypes.Right, uitypes.Ctrl}: "move-dot-right-word",
		uitypes.Key{uitypes.Home, 0}:             "move-dot-sol",
		uitypes.Key{uitypes.End, 0}:              "move-dot-eol",
		// Killing.
		uitypes.Key{'U', uitypes.Ctrl}:    "kill-line-left",
		uitypes.Key{'K', uitypes.Ctrl}:    "kill-line-right",
		uitypes.Key{'W', uitypes.Ctrl}:    "kill-word-left",
		uitypes.Key{uitypes.Backspace, 0}: "kill-rune-left",
		// Some terminal send ^H on backspace
		// uitypes.Key{'H', uitypes.Ctrl}: "kill-rune-left",
		uitypes.Key{uitypes.Delete, 0}: "kill-rune-right",
		// Inserting.
		uitypes.Key{'.', uitypes.Alt}:           "insert-last-word",
		uitypes.Key{uitypes.Enter, uitypes.Alt}: "insert-key",
		// Controls.
		uitypes.Key{uitypes.Enter, 0}:  "smart-enter",
		uitypes.Key{'D', uitypes.Ctrl}: "return-eof",
		uitypes.Key{uitypes.F2, 0}:     "toggle-quote-paste",
		// uitypes.Key{'[', uitypes.Ctrl}: "startCommand",
		uitypes.Key{uitypes.Tab, 0}:    "compl-prefix-or-start-compl",
		uitypes.Key{uitypes.Up, 0}:     "start-history",
		uitypes.Key{uitypes.Down, 0}:   "end-of-history",
		uitypes.Key{'N', uitypes.Ctrl}: "start-nav",
		uitypes.Key{'R', uitypes.Ctrl}: "start-histlist",
		uitypes.Key{',', uitypes.Alt}:  "start-bang",
		uitypes.Key{'L', uitypes.Ctrl}: "start-location",
		uitypes.Default:                "insert-default",
	},
	modeCommand: map[uitypes.Key]string{
		// Moving.
		uitypes.Key{'h', 0}: "move-dot-left",
		uitypes.Key{'l', 0}: "move-dot-right",
		uitypes.Key{'k', 0}: "move-dot-up",
		uitypes.Key{'j', 0}: "move-dot-down",
		uitypes.Key{'b', 0}: "move-dot-left-word",
		uitypes.Key{'w', 0}: "move-dot-right-word",
		uitypes.Key{'0', 0}: "move-dot-sol",
		uitypes.Key{'$', 0}: "move-dot-eol",
		// Killing.
		uitypes.Key{'x', 0}: "kill-rune-right",
		uitypes.Key{'D', 0}: "kill-line-right",
		// Controls.
		uitypes.Key{'i', 0}: "start-insert",
		uitypes.Default:     "default-command",
	},
	modeCompletion: map[uitypes.Key]string{
		uitypes.Key{uitypes.Up, 0}:     "compl-up",
		uitypes.Key{uitypes.Down, 0}:   "compl-down",
		uitypes.Key{uitypes.Tab, 0}:    "compl-down-cycle",
		uitypes.Key{uitypes.Left, 0}:   "compl-left",
		uitypes.Key{uitypes.Right, 0}:  "compl-right",
		uitypes.Key{uitypes.Enter, 0}:  "compl-accept",
		uitypes.Key{'F', uitypes.Ctrl}: "compl-trigger-filter",
		uitypes.Key{'[', uitypes.Ctrl}: "start-insert",
		uitypes.Default:                "compl-default",
	},
	modeNavigation: map[uitypes.Key]string{
		uitypes.Key{uitypes.Up, 0}:              "nav-up",
		uitypes.Key{uitypes.Down, 0}:            "nav-down",
		uitypes.Key{uitypes.PageUp, 0}:          "nav-page-up",
		uitypes.Key{uitypes.PageDown, 0}:        "nav-page-down",
		uitypes.Key{uitypes.Left, 0}:            "nav-left",
		uitypes.Key{uitypes.Right, 0}:           "nav-right",
		uitypes.Key{uitypes.Enter, uitypes.Alt}: "nav-insert-selected",
		uitypes.Key{uitypes.Enter, 0}:           "nav-insert-selected-and-quit",
		uitypes.Key{'H', uitypes.Ctrl}:          "nav-trigger-shown-hidden",
		uitypes.Key{'F', uitypes.Ctrl}:          "nav-trigger-filter",
		uitypes.Key{'[', uitypes.Ctrl}:          "start-insert",
		uitypes.Default:                         "navigation-default",
	},
	modeHistory: map[uitypes.Key]string{
		uitypes.Key{uitypes.Up, 0}:     "history-up",
		uitypes.Key{uitypes.Down, 0}:   "history-down-or-quit",
		uitypes.Key{'[', uitypes.Ctrl}: "start-insert",
		uitypes.Key{'R', uitypes.Ctrl}: "history-switch-to-histlist",
		uitypes.Default:                "history-default",
	},
	modeHistoryListing: map[uitypes.Key]string{
		uitypes.Key{'G', uitypes.Ctrl}: "histlist-toggle-case-sensitivity",
		uitypes.Key{'D', uitypes.Ctrl}: "histlist-toggle-dedup",
	},
	modeBang: map[uitypes.Key]string{
		uitypes.Default: "bang-alt-default",
	},
	modeLocation: map[uitypes.Key]string{},
}

var (
	builtinMap  = map[string]*BuiltinFn{}
	keyBindings = map[ModeType]map[uitypes.Key]eval.FnValue{}
)

func init() {
	addListingBuiltins("loc-", func(ed *Editor) *listing { return &ed.location.listing })
	addListingDefaultBindings("loc-", modeLocation)
	addListingBuiltins("histlist-", func(ed *Editor) *listing { return &ed.histlist.listing })
	addListingDefaultBindings("histlist-", modeHistoryListing)
	addListingBuiltins("bang-", func(ed *Editor) *listing { return &ed.bang.listing })
	addListingDefaultBindings("bang-", modeBang)

	for _, b := range builtins {
		builtinMap[b.name] = b
	}
	for mode, table := range defaultBindings {
		keyBindings[mode] = map[uitypes.Key]eval.FnValue{}
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
