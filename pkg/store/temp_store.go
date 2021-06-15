package store

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	bolt "go.etcd.io/bbolt"
)

// MustGetTempStore returns a Store backed by a temporary file, and a cleanup
// function that should be called when the Store is no longer used.
func MustGetTempStore() (DBStore, func()) {
	f, err := ioutil.TempFile("", "elvish.test")
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
	return st, func() {
		st.Close()
		f.Close()
		err = os.Remove(f.Name())
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to remove temp file:", err)
		}
	}
}
