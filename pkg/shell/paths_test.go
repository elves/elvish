package shell

import (
	"os"
	"testing"

	"src.elv.sh/pkg/env"
	"src.elv.sh/pkg/testutil"
)

func TestMakePaths_SetsAndCreatesDataDir(t *testing.T) {
	home, cleanupDir := testutil.TestDir()
	defer cleanupDir()
	cleanupEnv := testutil.WithTempEnv(env.HOME, home)
	defer cleanupEnv()

	paths, err := DataPaths()
	if err != nil {
		t.Fatal(err)
	}

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
