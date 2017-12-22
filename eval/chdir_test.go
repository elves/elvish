package eval

import (
	"os"
	"testing"
	"time"

	"github.com/elves/elvish/util"
)

type testAddDirer func(string, float64) error

func (t testAddDirer) AddDir(dir string, weight float64) error {
	return t(dir, weight)
}

func TestChdir(t *testing.T) {
	util.WithTempDir(func(destDir string) {
		pwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		defer func() {
			err = os.Chdir(pwd)
			if err != nil {
				panic(err)
			}
		}()

		chanAddedDir := make(chan string)
		testAddDirer := testAddDirer(func(dir string, weight float64) error {
			chanAddedDir <- dir
			return nil
		})

		err = Chdir(destDir, testAddDirer)

		if err != nil {
			t.Errorf("Chdir => error %v", err)
		}

		if env := os.Getenv("PWD"); env != destDir {
			t.Errorf("$PWD is %q after Chdir, want %q", env, destDir)
		}

		select {
		case addedDir := <-chanAddedDir:
			if addedDir != destDir {
				t.Errorf("Chdir called AddDir %q, want %q", addedDir[0], destDir)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Chdir did not call AddDir within 100ms")
		}
	})
}
