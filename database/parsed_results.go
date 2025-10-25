package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"
)

type ParsedResults struct {
	URL 		string
	Price		string
	Title 		string 
	ScrapedAt	time.Time
}

/* migrate parsed results */
func (p *Postgres) MigrateParsedResults(db *sql.DB) error {
    _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS parsed_results (
            url TEXT,
            price TEXT,
            title TEXT,
            scraped_at TIMESTAMP
        )
    `)
    return err
}

/* input of parsed results */
func (p *Postgres) InputParsedResults(url, price, title string, scraped_at time.Time) error {
	_, err := p.DB.Exec(
		"INSERT INTO parsed_results (url, price, title, scraped_at) VALUES ($1, $2, $3, $4)",
		url, price, title, scraped_at,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DB insert error: %v\n", err)
	}
	return err
}

/* read parsed results */
func (p *Postgres) ReadParsedResults() ([]ParsedResults, error) {
    rows, err := p.DB.Query("SELECT url, price, title, scraped_at FROM parsed_results")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var results []ParsedResults
    for rows.Next() {
        var r ParsedResults
        if err := rows.Scan(&r.URL, &r.Price, &r.Title, &r.ScrapedAt); err != nil {
            return nil, err
        }
        results = append(results, r)
    }
    return results, nil
}

/* rawhtml to parsedResults */
func (p *Postgres) ProcessRawHTML(parseFunc func(url string, html []byte) (price, title string, scrapedAt time.Time)) error {
    rows, err := p.DB.Query("SELECT url, html, scraped_at FROM raw_html")
    if err != nil {
        return err
    }
    defer rows.Close()

    for rows.Next() {
        var url string
        var html []byte
        var scrapedAt time.Time
        if err := rows.Scan(&url, &html, &scrapedAt); err != nil {
            return err
        }
        price, title, parsedAt := parseFunc(url, html)
        if err := p.InputParsedResults(url, price, title, parsedAt); err != nil {
            return err
        }
    }
    return nil
}