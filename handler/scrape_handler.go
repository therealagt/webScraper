package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"webScraper/scraper"
)

type BulkScrapeRequest struct {
	URLs    []string `json:"urls"`
	Depth   int      `json:"depth"`
	Keyword string   `json:"keyword"`
}

func BulkScrapeHandler(db *sql.DB, scraperInstance *scraper.Scraper, appCtx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req BulkScrapeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if len(req.URLs) == 0 {
			http.Error(w, "No URLs provided", http.StatusBadRequest)
			return
		}

		if req.Depth < 1 || req.Depth > 10 {
			req.Depth = 3 // default
		}

		// Start scraping in background
		go func() {
			ctx, cancel := context.WithTimeout(appCtx, 10*time.Minute)
			defer cancel()

			var wg sync.WaitGroup
			for _, url := range req.URLs {
				wg.Add(1)
				go func(url string) {
					defer wg.Done()
					scraperInstance.Scrape(ctx, url, req.Depth)
				}(url)
			}
			wg.Wait()
		}()

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"started","url_count":%d,"depth":%d}`, len(req.URLs), req.Depth)
	}
}
