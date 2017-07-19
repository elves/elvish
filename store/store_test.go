package store

// This file also sets up the test fixture.

import (
	"fmt"
	"io/ioutil"
	"os"
)

var tStore *Store

func init() {
	f, err := ioutil.TempFile("", "elvish.test")
	if err != nil {
		panic(fmt.Sprintf("Failed to open temp file: %v", err))
	}
	db, err := DefaultDB(f.Name())
	if err != nil {
		panic(fmt.Sprintf("Failed to create Store instance: %v", err))
	}
	os.Remove(f.Name())

	tStore, err = NewStoreDB(db)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Store instance: %v", err))
	}
}
