package store

type VisitedDir struct {
	Path  string
	Score float64
}

const (
	ScoreIncrement = 10
)

func init() {
	createTable["visited_dir"] = `create table if not exists visited_dir (path text unique primary key, score real default 0)`
}

func (s *Store) AddVisistedDir(d string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()

	// Insert when the path does not already exist
	_, err = tx.Exec("insert or ignore into visited_dir (path) values(?)", d)
	if err != nil {
		return err
	}

	// Increment score
	_, err = tx.Exec("update visited_dir set score = score + ? where path = ?", ScoreIncrement, d)
	return err
}

func (s *Store) FindVisitedDirs(p string) ([]VisitedDir, error) {
	rows, err := s.db.Query(
		"select path, score from visited_dir where path glob ? order by score desc",
		"*"+EscapeGlob(p)+"*")
	if err != nil {
		return nil, err
	}
	var (
		dir  VisitedDir
		dirs []VisitedDir
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
