package database

import (
	"database/sql"
)

/* migrate db - ORIGINAL MIGRATION WIEDERHERSTELLEN */
func MigrateDatabase(db *sql.DB) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		url TEXT NOT NULL,
		html TEXT,
		scraped_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		max_pages INTEGER,
		concurrency INTEGER,
		total_results INTEGER,
		completed_at TIMESTAMP
	);`

	_, err := db.Exec(createTableSQL)
	return err
}

/* NEUE separate Migration für raw_html */
func MigrateRawHTML(db *sql.DB) error {
	_, err := db.Exec(`DROP TABLE IF EXISTS raw_html`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS raw_html (
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
