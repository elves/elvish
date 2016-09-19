package store

import "database/sql"

// Dir is an entry in the directory history.
type Dir struct {
	Path  string
	Score float64
}

const (
	scoreDecay     = 0.986 // roughly 0.5^(1/50)
	scoreIncrement = 10
)

func init() {
	initDB["initialize directory history table"] = func(db *sql.DB) error {
		_, err := db.Exec(`create table if not exists dir (path text unique primary key, score real default 0)`)
		return err
	}
}

// AddDir adds a directory to the directory history.
func (s *Store) AddDir(d string, incFactor float64) error {
	return transaction(s.db, func(tx *sql.Tx) error {
		// Insert when the path does not already exist
		_, err := tx.Exec("insert or ignore into dir (path) values(?)", d)
		if err != nil {
			return err
		}

		// Decay scores
		_, err = tx.Exec("update dir set score = score * ?", scoreDecay)
		if err != nil {
			return err
		}

		// Increment score
		_, err = tx.Exec("update dir set score = score + ? where path = ?", scoreIncrement*incFactor, d)
		return err
	})
}

// ListDirs lists all directories in the directory history. The results are
// ordered by scores in descending order.
func (s *Store) ListDirs() ([]Dir, error) {
	rows, err := s.db.Query(
		"select path, score from dir order by score desc")
	if err != nil {
		return nil, err
	}
	return convertDirs(rows)
}

func convertDirs(rows *sql.Rows) ([]Dir, error) {
	var (
		dir  Dir
		dirs []Dir
	)

	for rows.Next() {
		rows.Scan(&dir.Path, &dir.Score)
		dirs = append(dirs, dir)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return dirs, nil
}
