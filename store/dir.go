package store

import (
	"database/sql"

	"github.com/coopernurse/gorp"
)

type Dir struct {
	Path  string
	Score float64
}

const (
	InitScore      = 10
	ScoreIncrement = 10
)

func init() {
	tableAdders = append(tableAdders, func(dm *gorp.DbMap) {
		t := dm.AddTable(Dir{}).SetKeys(false, "Path")
		t.ColMap("Path").SetUnique(true)
	})
}

func (s *Store) AddDir(d string) error {
	tx, err := s.dm.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()

	dir := Dir{}
	err = tx.SelectOne(&dir, "select * from Dir where Path=?", d)
	if err == sql.ErrNoRows {
		dir = Dir{Path: d, Score: InitScore}
		return tx.Insert(&dir)
	} else {
		dir.Score += ScoreIncrement
		_, err = tx.Update(&dir)
		return err
	}
}

func (s *Store) FindDirs(p string) ([]Dir, error) {
	var dirs []Dir
	_, err := s.dm.Select(
		&dirs,
		"select * from Dir where Path glob ? order by score desc",
		"*"+EscapeGlob(p)+"*")
	return dirs, err
}
