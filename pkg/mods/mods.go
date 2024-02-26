// Package mods collects standard library modules.
package mods

import (
	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/mods/doc"
	"src.elv.sh/pkg/mods/epm"
	"src.elv.sh/pkg/mods/file"
	"src.elv.sh/pkg/mods/flag"
	"src.elv.sh/pkg/mods/math"
	"src.elv.sh/pkg/mods/md"
	"src.elv.sh/pkg/mods/os"
	"src.elv.sh/pkg/mods/path"
	"src.elv.sh/pkg/mods/platform"
	"src.elv.sh/pkg/mods/re"
	readline_binding "src.elv.sh/pkg/mods/readline-binding"
	"src.elv.sh/pkg/mods/runtime"
	"src.elv.sh/pkg/mods/str"
	"src.elv.sh/pkg/mods/unix"
)

// AddTo adds all standard library modules to the Evaler.
//
// Some modules (the runtime module for now) may rely on properties set on the
// Evaler, so any mutations afterwards may not be properly reflected.
func AddTo(ev *eval.Evaler) {
	ev.AddModule("runtime", runtime.Ns(ev))
	ev.AddModule("math", math.Ns)
	ev.AddModule("path", path.Ns)
	ev.AddModule("platform", platform.Ns)
	ev.AddModule("re", re.Ns)
	ev.AddModule("str", str.Ns)
	ev.AddModule("file", file.Ns)
	ev.AddModule("flag", flag.Ns)
	ev.AddModule("doc", doc.Ns)
	ev.AddModule("os", os.Ns)
	ev.AddModule("md", md.Ns)
	if unix.ExposeUnixNs {
		ev.AddModule("unix", unix.Ns)
	}
	ev.BundledModules["epm"] = epm.Code
	ev.BundledModules["readline-binding"] = readline_binding.Code
}
