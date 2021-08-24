// Package mods collects standard library modules.
package mods

import (
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/mods/epm"
	"src.elv.sh/pkg/mods/file"
	"src.elv.sh/pkg/mods/math"
	"src.elv.sh/pkg/mods/path"
	"src.elv.sh/pkg/mods/platform"
	"src.elv.sh/pkg/mods/re"
	"src.elv.sh/pkg/mods/readlinebinding"
	"src.elv.sh/pkg/mods/str"
)

// AddTo adds all standard library modules to the Evaler.
func AddTo(ev *eval.Evaler) {
	ev.AddModule("math", math.Ns)
	ev.AddModule("path", path.Ns)
	ev.AddModule("platform", platform.Ns)
	ev.AddModule("re", re.Ns)
	ev.AddModule("str", str.Ns)
	ev.AddModule("file", file.Ns)
	ev.AddBundledModule("epm", epm.Code)
	ev.AddBundledModule("readline-binding", readlinebinding.Code)
}
