package store

import "database/sql"

// SchemaVersion is the current schema version. It should be bumped every time a
// backwards-incompatible change has been made to the schema.
const SchemaVersion = 1

func init() {
	initDB["record schema version"] = func(db *sql.DB) error {
		_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_version (version integer); INSERT INTO schema_version (version) VALUES(?)`, SchemaVersion)
		return err
	}
}

// SchemaUpToDate returns whether the database has the current or newer version
// of the schema.
func SchemaUpToDate(db *sql.DB) bool {
	var v int
	row := db.QueryRow(`SELECT version FROM schema_version`)
	return row.Scan(&v) == nil && v >= SchemaVersion
}
