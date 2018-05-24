package edcore

import (
	"github.com/elves/elvish/edit/ui"
)

// Styles for UI.
var (
	styleForReplacement    = ui.Styles{"underlined"}
	styleForTip            = ui.Styles{}
	styleForSelected       = ui.Styles{"inverse"}
	styleForScrollBarArea  = ui.Styles{"magenta"}
	styleForScrollBarThumb = ui.Styles{"magenta", "inverse"}

	styleForCompilerError = ui.Styles{"white", "bg-red"}
)
