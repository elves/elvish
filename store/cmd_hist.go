package store

import (
	"database/sql"
	"errors"
)

// ErrNoMatchingCmd is the error returned when a LastCmd or FirstCmd query
// completes with no result.
var ErrNoMatchingCmd = errors.New("no matching command line")

func init() {
	createTable["cmd"] = `create table if not exists cmd (content text)`
}

// NextCmdSeq returns the next sequence number of the command history.
func (s *Store) NextCmdSeq() (int, error) {
	row := s.db.QueryRow(`select ifnull(max(rowid), 0) + 1 from cmd`)
	var seq int
	err := row.Scan(&seq)
	return seq, err
}

// AddCmd adds a new command to the command history.
func (s *Store) AddCmd(cmd string) error {
	_, err := s.db.Exec(`insert into cmd (content) values(?)`, cmd)
	return err
}

// Cmd queries the command history item with the specified sequence number.
func (s *Store) Cmd(seq int) (string, error) {
	row := s.db.QueryRow(`select content from cmd where rowid = ?`, seq)
	var cmd string
	err := row.Scan(&cmd)
	return cmd, err
}

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

// LastCmd finds the last command before the given sequence number (exclusive)
// with the given prefix.
func (s *Store) LastCmd(upto int, prefix string) (int, string, error) {
	row := s.db.QueryRow(`select rowid, content from cmd where rowid < ? and substr(content, 1, ?) = ? order by rowid desc limit 1`, upto, len(prefix), prefix)
	return convertCmd(row)
}

// FirstCmd finds the first command after the given sequence number (inclusive)
// with the given prefix.
func (s *Store) FirstCmd(from int, prefix string) (int, string, error) {
	row := s.db.QueryRow(`select rowid, content from cmd where rowid >= ? and substr(content, 1, ?) = ? order by rowid asc limit 1`, from, len(prefix), prefix)
	return convertCmd(row)
}
