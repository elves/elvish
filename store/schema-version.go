package store

import "database/sql"

const SchemaVersion = 1

func init() {
	initDB["record schema version"] = func(db *sql.DB) error {
		_, err := db.Exec(`create table if not exists schema_version (version integer); insert into schema_version (version) values(?)`, SchemaVersion)
		return err
	}
}

func SchemaUpToDate(db *sql.DB) bool {
	var v int
	row := db.QueryRow(`select version from schema_version`)
	return row.Scan(&v) == nil && v >= SchemaVersion
}
