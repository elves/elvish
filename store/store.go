// Package store abstracts the persistent storage used by elvish.
package store

import (
	"database/sql"
	"fmt"
	"net/url"
	"sync"

	"github.com/elves/elvish/util"

	_ "github.com/mattn/go-sqlite3" // enable the "sqlite3" SQL driver
)

var Logger = util.GetLogger("[store] ")
var initDB = map[string](func(*sql.DB) error){}

// Store is the permanent storage backend for elvish.
type Store struct {
	db    *sql.DB
	Waits sync.WaitGroup
}

// DefaultDB returns the default database for storage.
func DefaultDB(dbname string) (*sql.DB, error) {
	uri := "file:" + url.QueryEscape(dbname) +
		"?mode=rwc&cache=shared&vfs=unix-dotfile"
	return sql.Open("sqlite3", uri)
}

// NewStore creates a new Store with the default database.
func NewStore(dbname string) (*Store, error) {
	db, err := DefaultDB(dbname)
	if err != nil {
		return nil, err
	}
	return NewStoreDB(db)
}

// NewStoreDB creates a new Store with a custom database. The database must be
// a SQLite database.
func NewStoreDB(db *sql.DB) (*Store, error) {
	Logger.Println("initializing store")
	defer Logger.Println("initialized store")
	st := &Store{db, sync.WaitGroup{}}

	if SchemaUpToDate(db) {
		Logger.Println("DB schema up to date")
	} else {
		for name, fn := range initDB {
			err := fn(db)
			if err != nil {
				return nil, fmt.Errorf("failed to %s: %v", name, err)
			}
		}
	}

	return st, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	s.Waits.Wait()
	return s.db.Close()
}
