package os

import (
	"fmt"
	"io/fs"

	"src.elv.sh/pkg/eval/vals"
)

var typeNames = map[fs.FileMode]string{
	0:                                 "regular",
	fs.ModeDir:                        "dir",
	fs.ModeSymlink:                    "symlink",
	fs.ModeNamedPipe:                  "named-pipe",
	fs.ModeSocket:                     "socket",
	fs.ModeDevice:                     "device",
	fs.ModeDevice | fs.ModeCharDevice: "char-device",
	fs.ModeIrregular:                  "irregular",
}

var specialModeNames = [...]struct {
	bit  fs.FileMode
	name string
}{
	// fs.ModeAppend, fs.ModeExclusive and fs.ModeTemporary are only used on
	// Plan 9, which Elvish doesn't support (yet).
	{fs.ModeSetuid, "setuid"},
	{fs.ModeSetgid, "setgid"},
	{fs.ModeSticky, "sticky"},
}

// Implementation of the stat function itself is in os.go.

func statMap(fi fs.FileInfo) vals.Map {
	mode := fi.Mode()
	typeName, ok := typeNames[mode.Type()]
	if !ok {
		// This shouldn't happen, but if there is a bug this gives us a bit of
		// information.
		typeName = fmt.Sprintf("unknown %d", mode.Type())
	}
	// TODO: Make this a set when Elvish has a set type.
	specialModes := vals.EmptyList
	for _, special := range specialModeNames {
		if mode&special.bit != 0 {
			specialModes = specialModes.Conj(special.name)
		}
	}
	return vals.MakeMap(
		"name", fi.Name(),
		"size", vals.Int64ToNum(fi.Size()),
		"type", typeName,
		"perm", int(fi.Mode()&fs.ModePerm),
		"special-modes", specialModes,
		"sys", statSysMap(fi.Sys()))
	// TODO: ModTime
}
