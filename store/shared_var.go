package store

import (
	"errors"

	"github.com/boltdb/bolt"
)

// ErrNoVar is returned by (*Store).GetSharedVar when there is no such variable.
var ErrNoVar = errors.New("no such variable")

const BucketSharedVar = "shared_var"

func init() {
	initDB["initialize shared variable table"] = func(db *bolt.DB) error {
		return db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(BucketSharedVar))
			return err
		})
	}
}

// SharedVar gets the value of a shared variable.
func (s *Store) SharedVar(n string) (string, error) {
	var value string
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketSharedVar))
		if v := b.Get([]byte(n)); v == nil {
			return ErrNoVar
		} else {
			value = string(v)
			return nil
		}
	})
	return value, err
}

// SetSharedVar sets the value of a shared variable.
func (s *Store) SetSharedVar(n, v string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketSharedVar))
		return b.Put([]byte(n), []byte(v))
	})
}

// DelSharedVar deletes a shared variable.
func (s *Store) DelSharedVar(n string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketSharedVar))
		return b.Delete([]byte(n))
	})
}
