package store

import "database/sql"

func hasColumn(rows *sql.Rows, colname string) (bool, error) {
	cols, err := rows.Columns()
	if err != nil {
		return false, err
	}
	dests := make([]interface{}, len(cols))
	var name string
	for i, col := range cols {
		if col == "name" {
			dests[i] = &name
		} else {
			dests[i] = new(interface{})
		}
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(dests...)
		if name == colname {
			return true, nil
		}
	}
	return false, rows.Err()
}
