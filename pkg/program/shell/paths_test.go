package shell

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/elves/elvish/pkg/util"
)

var j = filepath.Join

func TestMakePaths_PopulatesUnsetSubPaths(t *testing.T) {
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

func TestMakePaths_RespectsSetSubPaths(t *testing.T) {
	sock := "sock-override"
	paths := MakePaths(os.Stderr, Paths{
		RunDir: "run", DataDir: "data", Bin: "bin",
		Sock: sock,
	})
	if paths.Sock != sock {
		t.Errorf("got paths.Sock = %q, want %q", paths.Sock, sock)
	}
}

func TestMakePaths_SetsAndCreatesDataDir(t *testing.T) {
	home, cleanupDir := util.TestDir()
	defer cleanupDir()
	cleanupEnv := util.WithTempEnv("HOME", home)
	defer cleanupEnv()

	paths := MakePaths(os.Stderr, Paths{
		RunDir: "run", Bin: "bin",
	})

	wantDataDir := home + "/.elvish"
	if paths.DataDir != wantDataDir {
		t.Errorf("paths.DataDir = %q, want %q", paths.DataDir, wantDataDir)
	}

	stat, err := os.Stat(paths.DataDir)
	if err != nil {
		t.Errorf("could not stat %q: %v", paths.DataDir, err)
	}
	if !stat.IsDir() {
		t.Errorf("data dir %q is not dir", paths.DataDir)
	}
}
