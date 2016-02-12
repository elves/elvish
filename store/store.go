// Package store abstracts the persistent storage used by elvish.
package store

import (
	"database/sql"
	"fmt"
	"net/url"

	_ "github.com/mattn/go-sqlite3" // enable the "sqlite3" SQL driver
)

// Store is the permanent storage backend for elvish.
type Store struct {
	db *sql.DB
}

var initTable = map[string](func(*sql.DB) error){}

// DefaultDB returns the default database for storage.
func DefaultDB(dataDir string) (*sql.DB, error) {
	uri := "file:" + url.QueryEscape(dataDir+"/db") +
		"?mode=rwc&cache=shared&vfs=unix-dotfile"
	return sql.Open("sqlite3", uri)
}

// NewStore creates a new Store with the default database.
func NewStore(dataDir string) (*Store, error) {
	db, err := DefaultDB(dataDir)
	if err != nil {
		return nil, err
	}
	return NewStoreDB(db)
}

// NewStoreDB creates a new Store with a custom database. The database must be
// a SQLite database.
func NewStoreDB(db *sql.DB) (*Store, error) {
	st := &Store{db}

	for name, fn := range initTable {
		err := fn(db)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize table %s: %v", name, err)
		}
	}

	return st, nil
}
