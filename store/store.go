// Package store abstracts the persistent storage used by elvish.
package store

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/elves/elvish/store/storedefs"
	"github.com/elves/elvish/util"

	"github.com/boltdb/bolt"
)

var logger = util.GetLogger("[store] ")
var initDB = map[string](func(*bolt.DB) error){}

var ErrInvalidBucket = errors.New("invalid bucket")

// Store is the permanent storage backend for elvish. It is not thread-safe. In
// particular, the store may be closed while another goroutine is still
// accessing the store. To prevent bad things from happening, every time the
// main goroutine spawns a new goroutine to operate on the store, it should call
// Waits.Add(1) in the main goroutine before spawning another goroutine, and
// call Waits.Done() in the spawned goroutine after the operation is finished.
type Store struct {
	db *bolt.DB
	// Waits is used for registering outstanding operations on the store.
	waits sync.WaitGroup
}

var _ storedefs.Store = (*Store)(nil)

// DefaultDB returns the default database for storage.
func DefaultDB(dbname string) (*bolt.DB, error) {
	db, err := bolt.Open(dbname, 0644,
		&bolt.Options{
			Timeout: 1 * time.Second,
		})
	return db, err
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
// a Bolt database.
func NewStoreDB(db *bolt.DB) (*Store, error) {
	logger.Println("initializing store")
	defer logger.Println("initialized store")
	st := &Store{
		db:    db,
		waits: sync.WaitGroup{},
	}

	if SchemaUpToDate(db) {
		logger.Println("DB schema up to date")
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

// Waits returns a WaitGroup used to register outstanding storage requests when
// making calls asynchronously.
func (s *Store) Waits() *sync.WaitGroup {
	return &s.waits
}

// Close waits for all outstanding operations to finish, and closes the
// database.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	s.waits.Wait()
	return s.db.Close()
}
