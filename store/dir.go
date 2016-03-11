package store

import (
	"bytes"
	"database/sql"
)

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

// FindDirs finds directories containing a given substring. The results are
// ordered by scores in descending order.
func (s *Store) FindDirs(p string) ([]Dir, error) {
	rows, err := s.db.Query(
		"select path, score from dir where instr(path, ?) > 0 order by score desc", p)
	if err != nil {
		return nil, err
	}
	return convertDirs(rows)
}

// FindDirsLoose finds directories matching a given pattern. The results are
// ordered by scores in descending order.
//
// The pattern is first split on slashes, and have % attached to both sides of
// the parts. For instance, a/b becomes %a%/%b%, so it matches /1a1/2b2 as well
// as /home/xiaq/1a1/what/2b2.
func (s *Store) FindDirsLoose(p string) ([]Dir, error) {
	rows, err := s.db.Query(
		`select path, score from dir where path like ? escape "\" order by score desc`,
		makeLoosePattern(p))
	if err != nil {
		return nil, err
	}
	return convertDirs(rows)
}

func makeLoosePattern(pattern string) string {
	var b bytes.Buffer
	b.WriteRune('%')
	for _, p := range pattern {
		switch p {
		case '%':
			b.WriteString("\\%")
		case '\\':
			b.WriteString("\\\\")
		case '/':
			b.WriteString("%/%")
		default:
			b.WriteRune(p)
		}
	}
	b.WriteRune('%')
	return b.String()
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
