package edit

import (
	"fmt"

	"github.com/elves/elvish/eval"
)

// Line editor builtins.

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

	// Bang mode
	{"start-bang", startBang},
	{"bang-alt-default", bangAltDefault},

	// Location mode
	{"start-location", startLocation},

	// Misc
	{"redraw", redraw},
}

var defaultBindings = map[ModeType]map[Key]string{
	modeInsert: map[Key]string{
		// Moving.
		Key{Left, 0}:     "move-dot-left",
		Key{Right, 0}:    "move-dot-right",
		Key{Up, Alt}:     "move-dot-up",
		Key{Down, Alt}:   "move-dot-down",
		Key{Left, Ctrl}:  "move-dot-left-word",
		Key{Right, Ctrl}: "move-dot-right-word",
		Key{Home, 0}:     "move-dot-sol",
		Key{End, 0}:      "move-dot-eol",
		// Killing.
		Key{'U', Ctrl}:    "kill-line-left",
		Key{'K', Ctrl}:    "kill-line-right",
		Key{'W', Ctrl}:    "kill-word-left",
		Key{Backspace, 0}: "kill-rune-left",
		// Some terminal send ^H on backspace
		// Key{'H', Ctrl}: "kill-rune-left",
		Key{Delete, 0}: "kill-rune-right",
		// Inserting.
		Key{'.', Alt}:   "insert-last-word",
		Key{Enter, Alt}: "insert-key",
		// Controls.
		Key{Enter, 0}:  "smart-enter",
		Key{'D', Ctrl}: "return-eof",
		Key{F2, 0}:     "toggle-quote-paste",
		// Key{'[', Ctrl}: "startCommand",
		Key{Tab, 0}:    "compl-prefix-or-start-compl",
		Key{Up, 0}:     "start-history",
		Key{Down, 0}:   "end-of-history",
		Key{'N', Ctrl}: "start-nav",
		Key{'R', Ctrl}: "start-histlist",
		Key{',', Alt}:  "start-bang",
		Key{'L', Ctrl}: "start-location",
		Default:        "insert-default",
	},
	modeCommand: map[Key]string{
		// Moving.
		Key{'h', 0}: "move-dot-left",
		Key{'l', 0}: "move-dot-right",
		Key{'k', 0}: "move-dot-up",
		Key{'j', 0}: "move-dot-down",
		Key{'b', 0}: "move-dot-left-word",
		Key{'w', 0}: "move-dot-right-word",
		Key{'0', 0}: "move-dot-sol",
		Key{'$', 0}: "move-dot-eol",
		// Killing.
		Key{'x', 0}: "kill-rune-right",
		Key{'D', 0}: "kill-line-right",
		// Controls.
		Key{'i', 0}: "start-insert",
		Default:     "default-command",
	},
	modeCompletion: map[Key]string{
		Key{Up, 0}:     "compl-up",
		Key{Down, 0}:   "compl-down",
		Key{Tab, 0}:    "compl-down-cycle",
		Key{Left, 0}:   "compl-left",
		Key{Right, 0}:  "compl-right",
		Key{Enter, 0}:  "compl-accept",
		Key{'F', Ctrl}: "compl-trigger-filter",
		Key{'[', Ctrl}: "start-insert",
		Default:        "compl-default",
	},
	modeNavigation: map[Key]string{
		Key{Up, 0}:       "nav-up",
		Key{Down, 0}:     "nav-down",
		Key{PageUp, 0}:   "nav-page-up",
		Key{PageDown, 0}: "nav-page-down",
		Key{Left, 0}:     "nav-left",
		Key{Right, 0}:    "nav-right",
		Key{Tab, 0}:      "nav-insert-selected",
		Key{'H', Ctrl}:   "nav-trigger-shown-hidden",
		Key{'F', Ctrl}:   "nav-trigger-filter",
		Key{'[', Ctrl}:   "start-insert",
		Default:          "navigation-default",
	},
	modeHistory: map[Key]string{
		Key{Up, 0}:     "history-up",
		Key{Down, 0}:   "history-down-or-quit",
		Key{'[', Ctrl}: "start-insert",
		Key{'R', Ctrl}: "history-switch-to-histlist",
		Default:        "history-default",
	},
	modeHistoryListing: map[Key]string{},
	modeBang: map[Key]string{
		Default: "bang-alt-default",
	},
	modeLocation: map[Key]string{},
}

var (
	builtinMap  = map[string]*BuiltinFn{}
	keyBindings = map[ModeType]map[Key]eval.FnValue{}
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
		keyBindings[mode] = map[Key]eval.FnValue{}
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
