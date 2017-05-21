package store

import "database/sql"

func hasColumn(rows *sql.Rows, colname string) (bool, error) {
	cols, err := rows.Columns()
	if err != nil {
		return false, err
	}
	for _, col := range cols {
		if col == colname {
			return true, nil
		}
	}
	return false, rows.Err()
}

// transaction creates a Tx and calls f on it. It commits or rollbacks the
// transaction depending on whether f suceeded.
func transaction(db *sql.DB, f func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	err = f(tx)
	if err != nil {
		return tx.Rollback()
	}
	return tx.Commit()
}
