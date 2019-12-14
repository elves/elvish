package highlight

import "github.com/elves/elvish/ui"

var stylingFor = map[string]ui.Styling{
	barewordRegion:     nil,
	singleQuotedRegion: ui.Yellow,
	doubleQuotedRegion: ui.Yellow,
	variableRegion:     ui.Magenta,
	wildcardRegion:     nil,
	tildeRegion:        nil,

	commentRegion: ui.Cyan,

	">":  ui.Green,
	">>": ui.Green,
	"<":  ui.Green,
	"?>": ui.Green,
	"|":  ui.Green,
	"?(": ui.Bold,
	"(":  ui.Bold,
	")":  ui.Bold,
	"[":  ui.Bold,
	"]":  ui.Bold,
	"{":  ui.Bold,
	"}":  ui.Bold,
	"&":  ui.Bold,

	commandRegion: ui.Green,
	keywordRegion: ui.Yellow,
	errorRegion:   ui.BgRed,
}

var (
	stylingForGoodCommand = ui.Green
	stylingForBadCommand  = ui.Red
)
