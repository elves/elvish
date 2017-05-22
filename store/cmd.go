package store

import (
	"database/sql"
	"errors"
)

// ErrNoMatchingCmd is the error returned when a LastCmd or FirstCmd query
// completes with no result.
var ErrNoMatchingCmd = errors.New("no matching command line")

// Cmd is an entry in the command history.
type Cmd struct {
	Seq  int
	Text string
}

func init() {
	initDB["initialize command history table"] = func(db *sql.DB) error {
		_, err := db.Exec(`CREATE TABLE IF NOT EXISTS cmd (content text)`)
		return err
	}
}

// NextCmdSeq returns the next sequence number of the command history.
func (s *Store) NextCmdSeq() (int, error) {
	row := s.db.QueryRow(`SELECT ifnull(max(rowid), 0) + 1 FROM cmd`)
	var seq int
	err := row.Scan(&seq)
	return seq, err
}

// AddCmd adds a new command to the command history.
func (s *Store) AddCmd(cmd string) error {
	_, err := s.db.Exec(`INSERT INTO cmd (content) VALUES(?)`, cmd)
	return err
}

// GetCmd queries the command history item with the specified sequence number.
func (s *Store) GetCmd(seq int) (string, error) {
	row := s.db.QueryRow(`SELECT content FROM cmd WHERE rowid = ?`, seq)
	var cmd string
	err := row.Scan(&cmd)
	return cmd, err
}

// IterateCmds iterates all the commands in the specified range, and calls the
// callback with the content of each command sequentially.
func (s *Store) IterateCmds(from, upto int, f func(string) bool) error {
	rows, err := s.db.Query(`SELECT content FROM cmd WHERE rowid >= ? AND rowid < ?`, from, upto)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var cmd string
		err = rows.Scan(&cmd)
		if err != nil {
			break
		}
		if !f(cmd) {
			break
		}
	}
	return err
}

// GetCmds returns the contents of all commands within the specified range.
func (s *Store) GetCmds(from, upto int) ([]string, error) {
	var cmds []string
	err := s.IterateCmds(from, upto, func(cmd string) bool {
		cmds = append(cmds, cmd)
		return true
	})
	return cmds, err
}

// GetFirstCmd finds the first command after the given sequence number (inclusive)
// with the given prefix.
func (s *Store) GetFirstCmd(from int, prefix string) (Cmd, error) {
	row := s.db.QueryRow(`SELECT rowid, content FROM cmd WHERE rowid >= ? AND substr(content, 1, ?) = ? ORDER BY rowid asc LIMIT 1`, from, len(prefix), prefix)
	return convertCmd(row)
}

// GetLastCmd finds the last command before the given sequence number (exclusive)
// with the given prefix.
func (s *Store) GetLastCmd(upto int, prefix string) (Cmd, error) {
	var upto64 = int64(upto)
	if upto < 0 {
		upto64 = 0x7FFFFFFFFFFFFFFF
	}
	row := s.db.QueryRow(`SELECT rowid, content FROM cmd WHERE rowid < ? AND substr(content, 1, ?) = ? ORDER BY rowid DESC LIMIT 1`, upto64, len(prefix), prefix)
	return convertCmd(row)
}

func convertCmd(row *sql.Row) (Cmd, error) {
	var cmd Cmd
	err := row.Scan(&cmd.Seq, &cmd.Text)
	if err == sql.ErrNoRows {
		err = ErrNoMatchingCmd
	}
	return cmd, err
}
