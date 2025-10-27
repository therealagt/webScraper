package database

import (
	"database/sql"
)

/* migrate db */
func MigrateDatabase(db *sql.DB) error {
    _, err := db.Exec(`DROP TABLE IF EXISTS raw_html`)
    if err != nil {
        return err
    }
    _, err = db.Exec(`
        CREATE TABLE raw_html (
            id SERIAL PRIMARY KEY,
            url TEXT,
            max_pages INT,
            concurrency INT,
            html BYTEA,
            totalResults INT,
            completed_at TIMESTAMP
        )
    `)
    return err
}