package scraper

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Scraper struct {
	MaxConcurrency int
	Timeout int
	UserAgent string
	DB *sql.DB
}

/* config of a new scraper */
func NewScraper(maxConcurrency, timeout int, userAgent string, db *sql.DB) *Scraper {
	return &Scraper{
		MaxConcurrency: maxConcurrency,
		Timeout: timeout,
		UserAgent: userAgent,
		DB: db,
	}
}

/* setting for scrape operation */
func (s *Scraper) Scrape(url string, maxPages int) {
	sem := make(chan struct{}, s.MaxConcurrency)
	done := make(chan struct{})

	for i := 1; i <= maxPages; i++ {
		sem <- struct{}{}
		go func(page int) {
			defer func() { <-sem }()
			pageURL := fmt.Sprintf("%s?page=%d", url, page)
			s.fetchPage(pageURL)
			s.ReportProgress(i, maxPages)
			if page == maxPages {
				done <- struct{}{}
			}
		}(i)
	}

	for i := 0; i < s.MaxConcurrency; i++ {
		sem <- struct{}{}
	}
	fmt.Println("Scrape Job done!")
}


/* http get && error handling */
func (s *Scraper) fetchPage(url string) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read body %v\n", err)
		return 
	}
	s.saveRawHTMLToDB(url, body)
}

/* raw html db save */
func (s *Scraper) saveRawHTMLToDB(url string, body []byte) {
	_, err := s.DB.Exec(
		"INSERT INTO raw_html (url, html) VALUES ($1, $2)", url, body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DB insert error: %v\n", err)
	}
}

/* job progress */
func (s *Scraper)ReportProgress(current, total int) {
	fmt.Printf("Scraping progress: %d/%d pages done\n", current, total)
}