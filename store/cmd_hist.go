package store

import (
	"database/sql"
	"errors"
)

func init() {
	createTable["cmd"] = `create table if not exists cmd (content text)`
}

func (s *Store) NextCmdSeq() (int, error) {
	row := s.db.QueryRow(`select ifnull(max(rowid), 0) + 1 from cmd`)
	var seq int
	err := row.Scan(&seq)
	return seq, err
}

func (s *Store) AddCmd(cmd string) error {
	_, err := s.db.Exec(`insert into cmd (content) values(?)`, cmd)
	return err
}

func (s *Store) Cmd(seq int) (string, error) {
	row := s.db.QueryRow(`select content from cmd where rowid = ?`, seq)
	var cmd string
	err := row.Scan(&cmd)
	return cmd, err
}

var ErrNoMatchingCmd = errors.New("no matching command line")

func convertCmd(row *sql.Row) (int, string, error) {
	var (
		seq int
		cmd string
	)
	err := row.Scan(&seq, &cmd)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrNoMatchingCmd
		}
		return 0, "", err
	}
	return seq, cmd, nil
}

func (s *Store) LastCmdWithPrefix(upto int, prefix string) (int, string, error) {
	row := s.db.QueryRow(`select rowid, content from cmd where rowid < ? and substr(content, 1, ?) = ? order by rowid desc limit 1`, upto, len(prefix), prefix)
	return convertCmd(row)
}

func (s *Store) FirstCmdWithPrefix(from int, prefix string) (int, string, error) {
	row := s.db.QueryRow(`select rowid, content from cmd where rowid >= ? and substr(content, 1, ?) = ? order by rowid asc limit 1`, from, len(prefix), prefix)
	return convertCmd(row)
}
