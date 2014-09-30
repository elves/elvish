package store

import (
	"database/sql"
	"net/url"

	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	dm *gorp.DbMap
}

var tableAdders []func(*gorp.DbMap)

// DefaultDB returns the default database for storage.
func DefaultDB() (*sql.DB, error) {
	ddir, err := ensureDataDir()
	if err != nil {
		return nil, err
	}
	uri := "file:" + url.QueryEscape(ddir+"/db") +
		"?mode=rwc&cache=shared&vfs=unix-dotfile"
	return sql.Open("sqlite3", uri)
}

// NewStore creates a new Store with the default database.
func NewStore() (*Store, error) {
	db, err := DefaultDB()
	if err != nil {
		return nil, err
	}
	return NewStoreDB(db)
}

// NewStoreDB creates a new Store with a custom database. The database must be
// a SQLite database.
func NewStoreDB(db *sql.DB) (*Store, error) {
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
	for _, ta := range tableAdders {
		ta(dbmap)
	}
	err := dbmap.CreateTablesIfNotExists()
	if err != nil {
		return nil, err
	}
	return &Store{dbmap}, nil
}
