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

// Implementation of the stat function itself is in os.go.

func statMap(fi fs.FileInfo) vals.Map {
	mode := fi.Mode()
	typeName, ok := typeNames[mode.Type()]
	if !ok {
		// This shouldn't happen, but if there is a bug this gives us a bit of
		// information.
		typeName = fmt.Sprintf("unknown %d", mode.Type())
	}
	return vals.MakeMap(
		"name", fi.Name(),
		"size", vals.Int64ToNum(fi.Size()),
		"type", typeName,
		"perm", int(mode&fs.ModePerm),
		"special-modes", specialModesToList(mode),
		"sys", statSysMap(fi.Sys()))
	// TODO: ModTime
}
