package edit

import (
	"github.com/elves/elvish/edit/ui"
)

var styleForCompilerError = ui.Styles{"white", "bg-red"}

// Styles for UI.
var (
	//styleForPrompt           = ""
	//styleForRPrompt          = "inverse"
	styleForCompleted        = ui.Styles{"underlined"}
	styleForCompletedHistory = ui.Styles{"underlined"}
	styleForMode             = ui.Styles{"bold", "lightgray", "bg-magenta"}
	styleForTip              = ui.Styles{}
	styleForFilter           = ui.Styles{"underlined"}
	styleForSelected         = ui.Styles{"inverse"}
	styleForScrollBarArea    = ui.Styles{"magenta"}
	styleForScrollBarThumb   = ui.Styles{"magenta", "inverse"}

	// Use default style for completion listing
	styleForCompletion = ui.Styles{}
	// Use inverse style for selected completion entry
	styleForSelectedCompletion = ui.Styles{"inverse"}
)
