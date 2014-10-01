package store

type Dir struct {
	Path  string
	Score float64
}

const (
	ScoreIncrement = 10
)

func init() {
	createTable["dir"] = `create table if not exists dir (path text unique primary key, score real default 0)`
}

func (s *Store) AddDir(d string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()

	// Insert when the path does not already exist
	_, err = tx.Exec("insert or ignore into dir (path) values(?)", d)
	if err != nil {
		return err
	}

	// Increment score
	_, err = tx.Exec("update dir set score = score + ? where path = ?", ScoreIncrement, d)
	return err
}

func (s *Store) FindDirs(p string) ([]Dir, error) {
	rows, err := s.db.Query(
		"select path, score from dir where path glob ? order by score desc",
		"*"+EscapeGlob(p)+"*")
	if err != nil {
		return nil, err
	}
	var (
		dir  Dir
		dirs []Dir
	)

	for rows.Next() {
		rows.Scan(&dir.Path, &dir.Score)
		dirs = append(dirs, dir)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return dirs, nil
}
