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
)

type Scraper struct {
    MaxConcurrency int
    Timeout        int
    UserAgent      string
    DB             *sql.DB
    MaxPages       int
    TotalResults   int
    CompletedAt    sql.NullTime
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
func (s *Scraper) Scrape(ctx context.Context, url string, maxPages int) {
	var wg sync.WaitGroup
    concurrencyLimiter := make(chan struct{}, s.MaxConcurrency)

    totalResults := 0
    completedAt := sql.NullTime{Valid: false}

    if maxPages == 1 {
        s.fetchPage(ctx, url, maxPages, totalResults, completedAt)
        fmt.Println("Scrape Job done!")
        return
    }

    scrapeDone := make(chan struct{})
    for i := 1; i <= maxPages; i++ {
		wg.Add(1)
        concurrencyLimiter <- struct{}{}
        go func(page int) {
            defer func() { <-concurrencyLimiter }()
			defer wg.Done()
            sep := "?"
            if strings.Contains(url, "?") {
                sep = "&"
            }
            pageURL := fmt.Sprintf("%s%spage=%d", url, sep, page)
            s.fetchPage(ctx, pageURL, maxPages, totalResults, completedAt)
            if page == maxPages {
                scrapeDone <- struct{}{}
            }
        }(i)
    }

    for i := 0; i < s.MaxConcurrency; i++ {
        concurrencyLimiter <- struct{}{}
    }
	wg.Wait()
    fmt.Println("Scrape Job done!")
}


/* http get && error handling */
func (s *Scraper) fetchPage(ctx context.Context, url string, maxPages, totalResults int, completedAt sql.NullTime) {
    if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
        fmt.Fprintf(os.Stderr, "fetch: invalid url %q\n", url)
        return
    }
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := http.DefaultClient.Do(req)
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
