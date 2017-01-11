package store

import "database/sql"

const SchemaVersion = 1

func init() {
	initDB["record schema version"] = func(db *sql.DB) error {
		_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_version (version integer); INSERT INTO schema_version (version) VALUES(?)`, SchemaVersion)
		return err
	}
}

func SchemaUpToDate(db *sql.DB) bool {
	var v int
	row := db.QueryRow(`SELECT version FROM schema_version`)
	return row.Scan(&v) == nil && v >= SchemaVersion
}
