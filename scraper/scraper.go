package scraper

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Scraper struct {
	MaxConcurrency int
	Timeout        int
	UserAgent      string
	DB             *sql.DB
	MaxPages       int
	TotalResults   int
	CompletedAt    sql.NullTime
	client         *http.Client
}

/* config of a new scraper */
func NewScraper(maxConcurrency, timeout int, userAgent string, db *sql.DB) *Scraper {
	return &Scraper{
		MaxConcurrency: maxConcurrency,
		Timeout:        timeout,
		UserAgent:      userAgent,
		DB:             db,
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
			Transport: &http.Transport{
				TLSHandshakeTimeout:   5 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
				IdleConnTimeout:       90 * time.Second,
				DisableKeepAlives:     true,
			},
		},
	}
}

/* setting for scrape operation */
func (s *Scraper) Scrape(ctx context.Context, url string, maxPages int) {
	// Context-Überwachung hinzufügen
	select {
	case <-ctx.Done():
		return
	default:
	}

	var wg sync.WaitGroup
	concurrencyLimiter := make(chan struct{}, s.MaxConcurrency)

	totalResults := 0
	completedAt := sql.NullTime{Valid: false}

	if maxPages == 1 {
		s.fetchPage(ctx, url, maxPages, totalResults, completedAt)
		return
	}

	for i := 1; i <= maxPages; i++ {
		// Context-Check vor jeder neuen Seite
		select {
		case <-ctx.Done():
			return
		default:
		}

		wg.Add(1)
		concurrencyLimiter <- struct{}{}

		pageURL := fmt.Sprintf("%s?page=%d", url, i)
		go func(url string, page int) {
			defer func() {
				<-concurrencyLimiter
				wg.Done()
			}()

			s.fetchPage(ctx, url, maxPages, totalResults, completedAt)

			// Add delay between requests to be respectful to servers
			time.Sleep(2 * time.Second)
		}(pageURL, i)
	}

	// Warten mit Context-Überwachung
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Scraping completed
	case <-ctx.Done():
		// Scraping canceled
		return
	}
}

/* http get && error handling */
func (s *Scraper) fetchPage(ctx context.Context, url string, maxPages, totalResults int, completedAt sql.NullTime) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		fmt.Fprintf(os.Stderr, "fetch: invalid url %q\n", url)
		return
	}

	// Add delay before making request to be respectful to servers
	time.Sleep(1 * time.Second)

	// Context mit zusätzlichem Timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(s.Timeout)*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctxWithTimeout, "GET", url, nil)
	// user-Agent to avoid blocking
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36")

	resp, err := s.client.Do(req)
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

	s.saveRawHTMLToDB(
		url,
		body,
		maxPages,
		s.MaxConcurrency,
		totalResults,
		completedAt,
	)
}

/* raw html db save */
func (s *Scraper) saveRawHTMLToDB(url string, body []byte, maxPages, concurrency, totalResults int, completedAt sql.NullTime) {
	_, err := s.DB.Exec(
		`INSERT INTO raw_html 
        (url, max_pages, concurrency, html, totalResults, completed_at) 
        VALUES ($1, $2, $3, $4, $5, $6)`,
		url, maxPages, concurrency, body, totalResults, completedAt,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DB insert error: %v\n", err)
	}
}
