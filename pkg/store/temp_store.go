package store

import (
	"fmt"
	"os"
	"time"

	bolt "go.etcd.io/bbolt"
	"src.elv.sh/pkg/testutil"
)

// MustTempStore returns a Store backed by a temporary file for testing. The
// Store and its underlying file will be cleaned up properly after the test is
// finished.
func MustTempStore(c testutil.Cleanuper) DBStore {
	f, err := os.CreateTemp("", "elvish.test")
	if err != nil {
		panic(fmt.Sprintf("open temp file: %v", err))
	}
	db, err := bolt.Open(f.Name(), 0644, &bolt.Options{
		Timeout: time.Second, NoSync: true, NoFreelistSync: true})
	if err != nil {
		panic(fmt.Sprintf("open boltdb: %v", err))
	}

	st, err := NewStoreFromDB(db)
	if err != nil {
		panic(fmt.Sprintf("create Store instance: %v", err))
	}
	c.Cleanup(func() {
		st.Close()
		f.Close()
		err = os.Remove(f.Name())
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to remove temp file:", err)
		}
	})
	return st
}
