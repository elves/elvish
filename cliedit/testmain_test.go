package cliedit

import (
	"os"
	"testing"

	"github.com/elves/elvish/store"
	"github.com/elves/elvish/store/storedefs"
)

var (
	devNull   *os.File
	testStore storedefs.Store
)

// TestMain sets up the test fixture.
func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	// Set up devNull. This can be used as I/O for the editor when we do not
	// wish to test the IO.
	f, err := os.Open(os.DevNull)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	devNull = f

	// Set up testStore.
	st, cleanup := store.MustGetTempStore()
	defer cleanup()
	testStore = st

	// Add some data to testStore.
	_, err = testStore.AddCmd("echo hello world")
	if err != nil {
		panic(err)
	}

	return m.Run()
}
