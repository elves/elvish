package store

import (
	"sort"
	"strconv"

	bolt "go.etcd.io/bbolt"
	. "src.elv.sh/pkg/store/storedefs"
)

// Parameters for directory history scores.
const (
	DirScoreDecay     = 0.986 // roughly 0.5^(1/50)
	DirScoreIncrement = 10
	DirScorePrecision = 6
)

func init() {
	initDB["initialize directory history table"] = func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketDir))
		return err
	}
}

func marshalScore(score float64) []byte {
	return []byte(strconv.FormatFloat(score, 'E', DirScorePrecision, 64))
}

func unmarshalScore(data []byte) float64 {
	f, _ := strconv.ParseFloat(string(data), 64)
	return f
}

// AddDir adds a directory to the directory history.
func (s *dbStore) AddDir(d string, incFactor float64) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketDir))

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			score := unmarshalScore(v) * DirScoreDecay
			b.Put(k, marshalScore(score))
		}

		k := []byte(d)
		score := float64(0)
		if v := b.Get(k); v != nil {
			score = unmarshalScore(v)
		}
		score += DirScoreIncrement * incFactor
		return b.Put(k, marshalScore(score))
	})
}

// AddDir adds a directory and its score to history.
func (s *dbStore) AddDirRaw(d string, score float64) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketDir))
		return b.Put([]byte(d), marshalScore(score))
	})
}

// DelDir deletes a directory record from history.
func (s *dbStore) DelDir(d string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketDir))
		return b.Delete([]byte(d))
	})
}

// Dirs lists all directories in the directory history whose names are not
// in the blacklist. The results are ordered by scores in descending order.
func (s *dbStore) Dirs(blacklist map[string]struct{}) ([]Dir, error) {
	var dirs []Dir

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketDir))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			d := string(k)
			if _, ok := blacklist[d]; ok {
				continue
			}
			dirs = append(dirs, Dir{
				Path:  d,
				Score: unmarshalScore(v),
			})
		}
		sort.Sort(sort.Reverse(dirList(dirs)))
		return nil
	})
	return dirs, err
}

type dirList []Dir

func (dl dirList) Len() int {
	return len(dl)
}

func (dl dirList) Less(i, j int) bool {
	return dl[i].Score < dl[j].Score
}

func (dl dirList) Swap(i, j int) {
	dl[i], dl[j] = dl[j], dl[i]
}
