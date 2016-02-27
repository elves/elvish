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
	{"default-command", defaultCommand},
	{"default-insert", defaultInsert},

	// Completion mode
	{"complete-prefix-or-start-completion", completePrefixOrStartCompletion},
	{"start-completion", startCompletion},
	{"cancel-completion", cancelCompletion},
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
	{"quit-navigation", quitNavigation},
	{"default-navigation", defaultNavigation},

	// History mode
	{"start-history", startHistory},
	{"select-history-prev", selectHistoryPrev},
	{"select-history-next", selectHistoryNext},
	{"select-history-next-or-quit", selectHistoryNextOrQuit},
	{"default-history", defaultHistory},

	// History listing mode
	{"start-history-listing", startHistoryListing},
	{"default-history-listing", defaultHistoryListing},

	// Location mode
	{"start-location", startLocation},
	{"location-prev", locationPrev},
	{"location-next", locationNext},
	{"location-backspace", locationBackspace},
	{"accept-location", acceptLocation},
	{"cancel-location", cancelLocation},
	{"location-default", locationDefault},

	// Misc
	{"redraw", redraw},
}

var builtinMap = map[string]Builtin{}

func init() {
	for _, b := range builtins {
		builtinMap[b.name] = b
	}
	for mode, table := range defaultBindings {
		keyBindings[mode] = map[Key]Caller{}
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
