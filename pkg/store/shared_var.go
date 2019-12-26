package store

import (
	"errors"

	"github.com/boltdb/bolt"
)

// ErrNoVar is returned by (*Store).GetSharedVar when there is no such variable.
var ErrNoVar = errors.New("no such variable")

func init() {
	initDB["initialize shared variable table"] = func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketSharedVar))
		return err
	}
}

// SharedVar gets the value of a shared variable.
func (s *store) SharedVar(n string) (string, error) {
	var value string
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketSharedVar))
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
func (s *store) SetSharedVar(n, v string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketSharedVar))
		return b.Put([]byte(n), []byte(v))
	})
}

// DelSharedVar deletes a shared variable.
func (s *store) DelSharedVar(n string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketSharedVar))
		return b.Delete([]byte(n))
	})
}
