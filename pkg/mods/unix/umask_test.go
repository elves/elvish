//go:build unix

package unix

import (
	"os"
	"strconv"
	"sync"
	"testing"

	"src.elv.sh/pkg/must"
	"src.elv.sh/pkg/testutil"
)

func TestUmaskGetRace(t *testing.T) {
	testutil.Umask(t, 0o22)
	testutil.InTempDir(t)

	// An old implementation of UmaskVariable.Get had a bug where it will
	// briefly set the umask to 0, which can impact files created concurrently.
	for i := 0; i < 100; i++ {
		filename := strconv.Itoa(i)
		runParallel(
			func() {
				// Calling UmaskVariable.Get is much quicker than creating a
				// file, so do it in a loop to increase the chance of triggering
				// a race condition.
				for j := 0; j < 100; j++ {
					UmaskVariable{}.Get()
				}
			},
			func() {
				must.OK(create(filename, 0o666))
			})

		perm := must.OK1(os.Stat(filename)).Mode().Perm()
		if perm != 0o644 {
			t.Errorf("got perm %o, want 0o644 (run %d)", perm, i)
		}
	}
}

func runParallel(funcs ...func()) {
	var wg sync.WaitGroup
	wg.Add(len(funcs))
	for _, f := range funcs {
		f := f
		go func() {
			f()
			wg.Done()
		}()
	}
	wg.Wait()
}

func create(name string, perm os.FileMode) error {
	file, err := os.OpenFile(name, os.O_CREATE, perm)
	if err == nil {
		file.Close()
	}
	return err
}
