package store

import (
	"database/sql"
	"errors"
)

var ErrNoVar = errors.New("no such variable")

func init() {
	initDB["initialize shared variable table"] = func(db *sql.DB) error {
		_, err := db.Exec(`CREATE TABLE IF NOT EXISTS shared_var (name text UNIQUE PRIMARY KEY, value text)`)
		return err
	}
}

func (s *Store) GetSharedVar(n string) (string, error) {
	row := s.db.QueryRow(`SELECT value FROM shared_var WHERE name = ?`, n)
	var value string
	err := row.Scan(&value)
	if err == sql.ErrNoRows {
		err = ErrNoVar
	}
	return value, err
}

func (s *Store) SetSharedVar(n, v string) error {
	_, err := s.db.Exec(`INSERT OR REPLACE INTO shared_var (name, value) VALUES (?, ?)`, n, v)
	return err
}

func (s *Store) DelSharedVar(n string) error {
	_, err := s.db.Exec(`DELETE FROM shared_var WHERE name = ?`, n)
	return err
}
