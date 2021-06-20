package store

import (
	"fmt"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"
	"src.elv.sh/pkg/logutil"
	. "src.elv.sh/pkg/store/storedefs"
)

var logger = logutil.GetLogger("[store] ")
var initDB = map[string](func(*bolt.Tx) error){}

// DBStore is the permanent storage backend for elvish. It is not thread-safe.
// In particular, the store may be closed while another goroutine is still
// accessing the store. To prevent bad things from happening, every time the
// main goroutine spawns a new goroutine to operate on the store, it should call
// wg.Add(1) in the main goroutine before spawning another goroutine, and
// call wg.Done() in the spawned goroutine after the operation is finished.
type DBStore interface {
	Store
	Close() error
}

type dbStore struct {
	db *bolt.DB
	wg sync.WaitGroup // used for registering outstanding operations on the store
}

func dbWithDefaultOptions(dbname string) (*bolt.DB, error) {
	db, err := bolt.Open(dbname, 0644,
		&bolt.Options{
			Timeout: 1 * time.Second,
		})
	return db, err
}

// NewStore creates a new Store from the given file.
func NewStore(dbname string) (DBStore, error) {
	db, err := dbWithDefaultOptions(dbname)
	if err != nil {
		return nil, err
	}
	return NewStoreFromDB(db)
}

// NewStoreFromDB creates a new Store from a bolt DB.
func NewStoreFromDB(db *bolt.DB) (DBStore, error) {
	logger.Println("initializing store")
	defer logger.Println("initialized store")
	st := &dbStore{
		db: db,
		wg: sync.WaitGroup{},
	}

	err := db.Update(func(tx *bolt.Tx) error {
		for name, fn := range initDB {
			err := fn(tx)
			if err != nil {
				return fmt.Errorf("failed to %s: %v", name, err)
			}
		}
		return nil
	})
	return st, err
}

// Close waits for all outstanding operations to finish, and closes the
// database.
func (s *dbStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	s.wg.Wait()
	return s.db.Close()
}
