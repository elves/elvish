// Package path provides functions for manipulating filesystem path names.
package path

import (
	"os"
	"path/filepath"

	"src.elv.sh/pkg/eval"
	"src.elv.sh/pkg/eval/vars"
	osmod "src.elv.sh/pkg/mods/os"
)

// Ns is the namespace for the path: module.
var Ns = eval.BuildNsNamed("path").
	AddVars(map[string]vars.Var{
		"dev-null":       vars.NewReadOnly(os.DevNull),
		"dev-tty":        vars.NewReadOnly(osmod.DevTTY),
		"list-separator": vars.NewReadOnly(string(filepath.ListSeparator)),
		"separator":      vars.NewReadOnly(string(filepath.Separator)),
	}).
	AddGoFns(map[string]any{
		"abs":    filepath.Abs,
		"base":   filepath.Base,
		"clean":  filepath.Clean,
		"dir":    filepath.Dir,
		"ext":    filepath.Ext,
		"is-abs": filepath.IsAbs,
		"join":   filepath.Join,

		// Compatibility aliases; these have moved to os: but are kept here
		// until we can properly emit deprecation messages.
		"eval-symlinks": filepath.EvalSymlinks,
		"is-dir":        osmod.IsDir,
		"is-regular":    osmod.IsRegular,
		"temp-dir":      osmod.TempDir,
		"temp-file":     osmod.TempFile,
	}).Ns()
