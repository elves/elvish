package shell

import (
	"os"
	"path/filepath"
	"testing"
)

var j = filepath.Join

func TestMakePaths_PopulatesSubPaths(t *testing.T) {
	paths := MakePaths(os.Stderr, Paths{RunDir: "run", DataDir: "data", Bin: "bin"})
	wantPaths := Paths{
		RunDir:          "run",
		Sock:            j("run", "sock"),
		DaemonLogPrefix: j("run", "daemon.log-"),

		DataDir: "data",
		Db:      j("data", "db"),
		Rc:      j("data", "rc.elv"),
		LibDir:  j("data", "lib"),

		Bin: "bin",
	}
	if paths != wantPaths {
		t.Errorf("got paths %v, want %v", paths, wantPaths)
	}
}
