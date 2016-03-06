package store

import (
	"database/sql"
	"errors"
)

var ErrNoVar = errors.New("no such variable")

func init() {
	initDB["initialize shared variable table"] = func(db *sql.DB) error {
		_, err := db.Exec(`create table if not exists shared_var (name text unique primary key, value text)`)
		return err
	}
}

func (s *Store) GetSharedVar(n string) (string, error) {
	row := s.db.QueryRow(`select value from shared_var where name = ?`, n)
	var value string
	err := row.Scan(&value)
	if err == sql.ErrNoRows {
		err = ErrNoVar
	}
	return value, err
}

func (s *Store) SetSharedVar(n, v string) error {
	_, err := s.db.Exec(`insert or replace into shared_var (name, value) values (?, ?)`, n, v)
	return err
}

func (s *Store) DelSharedVar(n string) error {
	_, err := s.db.Exec(`delete from shared_var where name = ?`, n)
	return err
}
