package store

import (
	"sort"
	"strconv"

	"github.com/boltdb/bolt"
	"github.com/elves/elvish/store/storedefs"
)

const (
	scoreDecay     = 0.986 // roughly 0.5^(1/50)
	scoreIncrement = 10
	scorePrecision = 6
)

const BucketDir = "dir"

func init() {
	initDB["initialize directory history table"] = func(db *bolt.DB) error {
		return db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(BucketDir))
			return err
		})
	}
}

func marshalScore(score float64) []byte {
	return []byte(strconv.FormatFloat(score, 'E', scorePrecision, 64))
}
func unmarshalScore(data []byte) float64 {
	f, _ := strconv.ParseFloat(string(data), 64)
	return f
}

// AddDir adds a directory to the directory history.
func (s *Store) AddDir(d string, incFactor float64) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketDir))

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			score := unmarshalScore(v) * scoreDecay
			b.Put(k, marshalScore(score))
		}

		k := []byte(d)
		score := float64(0)
		if v := b.Get(k); v != nil {
			score = unmarshalScore(v)
		}
		score = score + scoreIncrement*incFactor
		return b.Put(k, marshalScore(score))
	})
}

// AddDir adds a directory and its score to history.
func (s *Store) AddDirRaw(d string, score float64) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketDir))
		return b.Put([]byte(d), marshalScore(score))
	})
}

// DelDir deletes a directory record from history.
func (s *Store) DelDir(d string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketDir))
		return b.Delete([]byte(d))
	})
}

// Dirs lists all directories in the directory history whose names are not
// in the blacklist. The results are ordered by scores in descending order.
func (s *Store) Dirs(blacklist map[string]struct{}) ([]storedefs.Dir, error) {
	var dirs []storedefs.Dir

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketDir))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			d := string(k)
			if _, ok := blacklist[d]; ok {
				continue
			}
			dirs = append(dirs, storedefs.Dir{
				Path:  d,
				Score: unmarshalScore(v),
			})
		}
		sort.Sort(sort.Reverse(dirList(dirs)))
		return nil
	})
	return dirs, err
}

type dirList []storedefs.Dir

func (dl dirList) Len() int {
	return len(dl)
}

func (dl dirList) Less(i, j int) bool {
	return dl[i].Score < dl[j].Score
}

func (dl dirList) Swap(i, j int) {
	dl[i], dl[j] = dl[j], dl[i]
}
