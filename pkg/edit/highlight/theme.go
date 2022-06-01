package highlight

import (
	"src.elv.sh/pkg/ui"
)

var stylingFor = map[string]ui.Styling{
	barewordRegion:     nil,
	singleQuotedRegion: ui.FgYellow,
	doubleQuotedRegion: ui.FgYellow,
	variableRegion:     ui.FgMagenta,
	wildcardRegion:     nil,
	tildeRegion:        nil,

	commentRegion: ui.FgCyan,

	">":  ui.FgGreen,
	">>": ui.FgGreen,
	"<":  ui.FgGreen,
	"?>": ui.FgGreen,
	"|":  ui.FgGreen,
	"?(": ui.Bold,
	"(":  ui.Bold,
	")":  ui.Bold,
	"[":  ui.Bold,
	"]":  ui.Bold,
	"{":  ui.Bold,
	"}":  ui.Bold,
	"&":  ui.Bold,

	commandRegion: ui.FgGreen,
	keywordRegion: ui.FgYellow,
	errorRegion:   ui.Stylings(ui.FgBrightWhite, ui.BgRed),
}

var (
	stylingForGoodCommand = ui.FgGreen
	stylingForBadCommand  = ui.FgRed
)
