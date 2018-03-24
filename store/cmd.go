package store

import (
	"bytes"
	"encoding/binary"

	"github.com/boltdb/bolt"
	"github.com/elves/elvish/store/storedefs"
)

func init() {
	initDB["initialize command history table"] = func(db *bolt.DB) error {
		return db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(BucketCmd))
			return err
		})
	}
}

const BucketCmd = "cmd"

// NextCmdSeq returns the next sequence number of the command history.
func (s *Store) NextCmdSeq() (int, error) {
	var seq uint64
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketCmd))
		seq = b.Sequence() + 1
		return nil
	})
	return int(seq), err
}

// AddCmd adds a new command to the command history.
func (s *Store) AddCmd(cmd string) (int, error) {
	var (
		seq uint64
		err error
	)
	err = s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketCmd))
		seq, err = b.NextSequence()
		if err != nil {
			return err
		}
		return b.Put(marshalSeq(seq), []byte(cmd))
	})
	return int(seq), err
}

// DelCmd deletes a command history item with the given sequence number.
func (s *Store) DelCmd(seq int) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketCmd))
		return b.Delete(marshalSeq(uint64(seq)))
	})
}

// Cmd queries the command history item with the specified sequence number.
func (s *Store) Cmd(seq int) (string, error) {
	var cmd string
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketCmd))
		if v := b.Get(marshalSeq(uint64(seq))); v == nil {
			return storedefs.ErrNoMatchingCmd
		} else {
			cmd = string(v)
		}
		return nil
	})
	return cmd, err
}

// IterateCmds iterates all the commands in the specified range, and calls the
// callback with the content of each command sequentially.
func (s *Store) IterateCmds(from, upto int, f func(string) bool) error {
	return s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketCmd))
		c := b.Cursor()
		for k, v := c.Seek(marshalSeq(uint64(from))); k != nil && unmarshalSeq(k) < uint64(upto); k, v = c.Next() {
			if !f(string(v)) {
				break
			}
		}
		return nil
	})
}

// Cmds returns the contents of all commands within the specified range.
func (s *Store) Cmds(from, upto int) ([]string, error) {
	var cmds []string
	err := s.IterateCmds(from, upto, func(cmd string) bool {
		cmds = append(cmds, cmd)
		return true
	})
	return cmds, err
}

// NextCmd finds the first command after the given sequence number (inclusive)
// with the given prefix.
func (s *Store) NextCmd(from int, prefix string) (int, string, error) {
	var (
		seq   int
		cmd   string
		found bool
	)
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketCmd))
		c := b.Cursor()
		p := []byte(prefix)
		for k, v := c.Seek(marshalSeq(uint64(from))); k != nil; k, v = c.Next() {
			if bytes.HasPrefix(v, p) {
				seq = int(unmarshalSeq(k))
				cmd = string(v)
				found = true
				break
			}
		}
		return nil
	})

	if !found {
		return 0, "", storedefs.ErrNoMatchingCmd
	}

	return seq, cmd, err
}

// PrevCmd finds the last command before the given sequence number (exclusive)
// with the given prefix.
func (s *Store) PrevCmd(upto int, prefix string) (int, string, error) {
	var (
		seq   int
		cmd   string
		found bool
	)

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketCmd))
		c := b.Cursor()
		p := []byte(prefix)

		k, v := c.Seek(marshalSeq(uint64(upto)))
		if k == nil { // upto > LAST
			k, v = c.Last()
			if k == nil {
				return nil
			}
		} else {
			k, v = c.Prev() // upto exists, find the previous one
		}

		for ; k != nil; k, v = c.Prev() {
			if bytes.HasPrefix(v, p) {
				seq = int(unmarshalSeq(k))
				cmd = string(v)
				found = true
				break
			}
		}
		return nil
	})

	if !found {
		return 0, "", storedefs.ErrNoMatchingCmd
	}

	return seq, cmd, err
}

func marshalSeq(seq uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(seq))
	return b
}

func unmarshalSeq(key []byte) uint64 {
	return binary.BigEndian.Uint64(key)
}
