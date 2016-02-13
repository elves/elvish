package store

import (
	"database/sql"
	"errors"
)

// ErrNoMatchingCmd is the error returned when a LastCmd or FirstCmd query
// completes with no result.
var ErrNoMatchingCmd = errors.New("no matching command line")

func init() {
	initTable["cmd"] = func(db *sql.DB) error {
		_, err := db.Exec(`create table if not exists cmd (content text, lastAmongDup bool)`)
		if err != nil {
			return err
		}
		return transaction(db, addLastAmongDup)
	}
}

func addLastAmongDup(tx *sql.Tx) error {
	// Upgrade from early version where lastAmongDup is missing.
	rows, err := tx.Query("select * from cmd limit 1")
	if err != nil {
		return err
	}
	defer rows.Close()

	hasLastAmongDup, err := hasColumn(rows, "lastAmongDup")
	if err != nil {
		return err
	}

	if !hasLastAmongDup {
		_, err := tx.Exec("alter table cmd add column lastAmongDup bool")
		if err != nil {
			return err
		}
		_, err = tx.Exec("update cmd set lastAmongDup = (rowid in (select max(rowid) from cmd group by content));")
		if err != nil {
			return err
		}
	}
	return nil
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
	_, err := s.db.Exec(`update cmd set lastAmongDup = 0 where content = ?`, cmd)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`insert into cmd (content, lastAmongDup) values(?, 1)`, cmd)
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
func (s *Store) LastCmd(upto int, prefix string, uniq bool) (int, string, error) {
	var upto64 int64 = int64(upto)
	if upto < 0 {
		upto64 = 0x7FFFFFFFFFFFFFFF
	}
	row := s.db.QueryRow(`select rowid, content from cmd where rowid < ? and substr(content, 1, ?) = ? and (? or lastAmongDup) order by rowid desc limit 1`, upto64, len(prefix), prefix, !uniq)
	return convertCmd(row)
}

// FirstCmd finds the first command after the given sequence number (inclusive)
// with the given prefix.
func (s *Store) FirstCmd(from int, prefix string, uniq bool) (int, string, error) {
	row := s.db.QueryRow(`select rowid, content from cmd where rowid >= ? and substr(content, 1, ?) = ? and (? or lastAmongDup) order by rowid asc limit 1`, from, len(prefix), prefix, !uniq)
	return convertCmd(row)
}
