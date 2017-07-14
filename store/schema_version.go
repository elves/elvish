package store

import (
	"strconv"

	"github.com/boltdb/bolt"
)

// SchemaVersion is the current schema version. It should be bumped every time a
// backwards-incompatible change has been made to the schema.
const SchemaVersion = 1

const BucketSchema = "schema"

func init() {
	initDB["record schema version"] = func(db *bolt.DB) error {
		return db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte(BucketSchema))
			if err != nil {
				return err
			}
			return b.Put([]byte("schema"), []byte(strconv.FormatUint(uint64(SchemaVersion), 10)))
		})
	}
}

// SchemaUpToDate returns whether the database has the current or newer version
// of the schema.
func SchemaUpToDate(db *bolt.DB) bool {
	var version uint64
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BucketSchema))
		if b == nil {
			return ErrInvalidBucket
		}
		if v := b.Get([]byte("version")); v != nil {
			version, _ = strconv.ParseUint(string(v), 0, 0)
		}
		return nil
	})

	if err != nil {
		return false
	}

	return version == SchemaVersion
}
