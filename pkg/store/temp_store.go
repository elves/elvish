package store

import (
	"fmt"
	"io/ioutil"
	"os"
)

// MustGetTempStore returns a Store backed by a temporary file, and a cleanup
// function that should be called when the Store is no longer used.
func MustGetTempStore() (DBStore, func()) {
	f, err := ioutil.TempFile("", "elvish.test")
	if err != nil {
		panic(fmt.Sprintf("Failed to open temp file: %v", err))
	}
	st, err := NewStore(f.Name())
	if err != nil {
		panic(fmt.Sprintf("Failed to create Store instance: %v", err))
	}
	return st, func() {
		st.Close()
		f.Close()
		err = os.Remove(f.Name())
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to remove temp file:", err)
		}
	}
}
