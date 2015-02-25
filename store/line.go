package store

import (
	"database/sql"
	"errors"
)

func init() {
	createTable["line"] = `create table if not exists line (content text)`
}

func (s *Store) GetNextLineSeq() (int, error) {
	row := s.db.QueryRow(`select ifnull(max(rowid), 0) + 1 from line`)
	var seq int
	err := row.Scan(&seq)
	return seq, err
}

func (s *Store) AddLine(line string) error {
	_, err := s.db.Exec(`insert into line (content) values(?)`, line)
	return err
}

func (s *Store) GetLine(seq int) (string, error) {
	row := s.db.QueryRow(`select content from line where rowid = ?`, seq)
	var line string
	err := row.Scan(&line)
	return line, err
}

var ErrNoMatchingLine = errors.New("no matching line")

func convertLine(row *sql.Row) (int, string, error) {
	var (
		seq  int
		line string
	)
	err := row.Scan(&seq, &line)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrNoMatchingLine
		}
		return 0, "", err
	}
	return seq, line, nil
}

func (s *Store) GetLastLineWithPrefix(upto int, prefix string) (int, string, error) {
	row := s.db.QueryRow(`select rowid, content from line where rowid < ? and substr(content, 1, ?) = ? order by rowid desc limit 1`, upto, len(prefix), prefix)
	return convertLine(row)
}

func (s *Store) GetFirstLineWithPrefix(from int, prefix string) (int, string, error) {
	row := s.db.QueryRow(`select rowid, content from line where rowid >= ? and substr(content, 1, ?) = ? order by rowid asc limit 1`, from, len(prefix), prefix)
	return convertLine(row)
}
