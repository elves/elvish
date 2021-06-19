package store

import (
	"bytes"
	"encoding/binary"

	bolt "go.etcd.io/bbolt"
	. "src.elv.sh/pkg/store/storedefs"
)

func init() {
	initDB["initialize command history table"] = func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketCmd))
		return err
	}
}

// NextCmdSeq returns the next sequence number of the command history.
func (s *dbStore) NextCmdSeq() (int, error) {
	var seq uint64
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketCmd))
		seq = b.Sequence() + 1
		return nil
	})
	return int(seq), err
}

// AddCmd adds a new command to the command history.
func (s *dbStore) AddCmd(cmd string) (int, error) {
	var (
		seq uint64
		err error
	)
	err = s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketCmd))
		seq, err = b.NextSequence()
		if err != nil {
			return err
		}
		return b.Put(marshalSeq(seq), []byte(cmd))
	})
	return int(seq), err
}

// DelCmd deletes a command history item with the given sequence number.
func (s *dbStore) DelCmd(seq int) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketCmd))
		return b.Delete(marshalSeq(uint64(seq)))
	})
}

// Cmd queries the command history item with the specified sequence number.
func (s *dbStore) Cmd(seq int) (string, error) {
	var cmd string
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketCmd))
		v := b.Get(marshalSeq(uint64(seq)))
		if v == nil {
			return ErrNoMatchingCmd
		}
		cmd = string(v)
		return nil
	})
	return cmd, err
}

// IterateCmds iterates all the commands in the specified range, and calls the
// callback with the content of each command sequentially.
func (s *dbStore) IterateCmds(from, upto int, f func(Cmd)) error {
	return s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketCmd))
		c := b.Cursor()
		for k, v := c.Seek(marshalSeq(uint64(from))); k != nil && unmarshalSeq(k) < uint64(upto); k, v = c.Next() {
			f(Cmd{Text: string(v), Seq: int(unmarshalSeq(k))})
		}
		return nil
	})
}

// CmdsWithSeq returns all commands within the specified range.
func (s *dbStore) CmdsWithSeq(from, upto int) ([]Cmd, error) {
	var cmds []Cmd
	err := s.IterateCmds(from, upto, func(cmd Cmd) {
		cmds = append(cmds, cmd)
	})
	return cmds, err
}

// NextCmd finds the first command after the given sequence number (inclusive)
// with the given prefix.
func (s *dbStore) NextCmd(from int, prefix string) (Cmd, error) {
	var cmd Cmd
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketCmd))
		c := b.Cursor()
		p := []byte(prefix)
		for k, v := c.Seek(marshalSeq(uint64(from))); k != nil; k, v = c.Next() {
			if bytes.HasPrefix(v, p) {
				cmd = Cmd{Text: string(v), Seq: int(unmarshalSeq(k))}
				return nil
			}
		}
		return ErrNoMatchingCmd
	})
	return cmd, err
}

// PrevCmd finds the last command before the given sequence number (exclusive)
// with the given prefix.
func (s *dbStore) PrevCmd(upto int, prefix string) (Cmd, error) {
	var cmd Cmd
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketCmd))
		c := b.Cursor()
		p := []byte(prefix)

		var v []byte
		k, _ := c.Seek(marshalSeq(uint64(upto)))
		if k == nil { // upto > LAST
			k, v = c.Last()
			if k == nil {
				return ErrNoMatchingCmd
			}
		} else {
			k, v = c.Prev() // upto exists, find the previous one
		}

		for ; k != nil; k, v = c.Prev() {
			if bytes.HasPrefix(v, p) {
				cmd = Cmd{Text: string(v), Seq: int(unmarshalSeq(k))}
				return nil
			}
		}
		return ErrNoMatchingCmd
	})
	return cmd, err
}

func marshalSeq(seq uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, seq)
	return b
}

func unmarshalSeq(key []byte) uint64 {
	return binary.BigEndian.Uint64(key)
}
