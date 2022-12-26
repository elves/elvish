package unix

import "embed"

// DElvFiles embeds all the .d.elv files for this module.
//
//go:embed *.d.elv
var DElvFiles embed.FS
