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

// NoBlacklist is an empty blacklist, to be used in GetDirs.
var NoBlacklist = map[string]struct{}{}

// GetDirs lists all directories in the directory history whose names are not
// in the blacklist. The results are ordered by scores in descending order.
func (s *Store) GetDirs(blacklist map[string]struct{}) ([]Dir, error) {
	rows, err := s.db.Query(
		"select path, score from dir order by score desc")
	if err != nil {
		return nil, err
	}
	return convertDirs(rows, blacklist)
}

func convertDirs(rows *sql.Rows, blacklist map[string]struct{}) ([]Dir, error) {
	var (
		dir  Dir
		dirs []Dir
	)

	for rows.Next() {
		rows.Scan(&dir.Path, &dir.Score)
		if _, black := blacklist[dir.Path]; !black {
			dirs = append(dirs, dir)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return dirs, nil
}
