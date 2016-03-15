package edit

import "fmt"

// Line editor builtins.

// Builtin records an editor builtin.
type Builtin struct {
	name string
	impl func(ed *Editor)
}

var builtins = []Builtin{
	// Command and insert mode
	{"start-insert", startInsert},
	{"start-command", startCommand},
	{"kill-line-left", killLineLeft},
	{"kill-line-right", killLineRight},
	{"kill-word-left", killWordLeft},
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
	{"default-command", defaultCommand},
	{"default-insert", defaultInsert},

	// Completion mode
	{"complete-prefix-or-start-completion", completePrefixOrStartCompletion},
	{"start-completion", startCompletion},
	{"select-cand-up", selectCandUp},
	{"select-cand-down", selectCandDown},
	{"select-cand-left", selectCandLeft},
	{"select-cand-right", selectCandRight},
	{"cycle-cand-right", cycleCandRight},
	{"accept-completion", acceptCompletion},
	{"default-completion", defaultCompletion},

	// Navigation mode
	{"start-navigation", startNavigation},
	{"select-nav-up", selectNavUp},
	{"select-nav-down", selectNavDown},
	{"ascend-nav", ascendNav},
	{"descend-nav", descendNav},
	{"trigger-nav-show-hidden", triggerNavShowHidden},
	{"nav-insert-selected", navInsertSelected},
	{"default-navigation", defaultNavigation},

	// History mode
	{"start-history", startHistory},
	{"select-history-prev", selectHistoryPrev},
	{"select-history-next", selectHistoryNext},
	{"select-history-next-or-quit", selectHistoryNextOrQuit},
	{"default-history", defaultHistory},

	// History listing mode
	{"start-history-listing", startHistoryListing},

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
		Key{Tab, 0}:    "complete-prefix-or-start-completion",
		Key{Up, 0}:     "start-history",
		Key{'N', Ctrl}: "start-navigation",
		Key{'H', Ctrl}: "start-history-listing",
		Key{',', Alt}:  "start-bang",
		Key{'L', Ctrl}: "start-location",
		Default:        "default-insert",
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
		Key{Up, 0}:     "select-cand-up",
		Key{Down, 0}:   "select-cand-down",
		Key{Left, 0}:   "select-cand-left",
		Key{Right, 0}:  "select-cand-right",
		Key{Tab, 0}:    "cycle-cand-right",
		Key{Enter, 0}:  "accept-completion",
		Key{'[', Ctrl}: "start-insert",
		Default:        "default-completion",
	},
	modeNavigation: map[Key]string{
		Key{Up, 0}:     "select-nav-up",
		Key{Down, 0}:   "select-nav-down",
		Key{Left, 0}:   "ascend-nav",
		Key{Right, 0}:  "descend-nav",
		Key{Tab, 0}:    "nav-insert-selected",
		Key{'H', Ctrl}: "trigger-nav-show-hidden",
		Key{'[', Ctrl}: "start-insert",
		Default:        "default-navigation",
	},
	modeHistory: map[Key]string{
		Key{Up, 0}:     "select-history-prev",
		Key{Down, 0}:   "select-history-next-or-quit",
		Key{'[', Ctrl}: "start-insert",
		Default:        "default-history",
	},
	modeHistoryListing: map[Key]string{},
	modeBang: map[Key]string{
		Default: "bang-alt-default",
	},
	modeLocation: map[Key]string{},
}

var (
	builtinMap  = map[string]Builtin{}
	keyBindings = map[ModeType]map[Key]BoundFunc{}
)

func init() {
	addListingBuiltins("loc-", func(ed *Editor) *listing { return &ed.location })
	addListingDefaultBindings("loc-", modeLocation)
	addListingBuiltins("histlist-", func(ed *Editor) *listing { return &ed.histlist })
	addListingDefaultBindings("histlist-", modeHistoryListing)
	addListingBuiltins("bang-", func(ed *Editor) *listing { return &ed.bang })
	addListingDefaultBindings("bang-", modeBang)

	for _, b := range builtins {
		builtinMap[b.name] = b
	}
	for mode, table := range defaultBindings {
		keyBindings[mode] = map[Key]BoundFunc{}
		for key, name := range table {
			caller, ok := builtinMap[name]
			if !ok {
				fmt.Println("bad name " + name)
			} else {
				keyBindings[mode][key] = caller
			}
		}
	}
}

type action struct {
	typ        actionType
	returnLine string
	returnErr  error
}

type actionType int

const (
	noAction actionType = iota
	reprocessKey
	exitReadLine
)

func redraw(ed *Editor) {
	ed.refresh(true, true)
}
