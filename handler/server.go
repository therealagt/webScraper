package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"webScraper/scraper"
)

/* route handling */
func SetupRoutes(db *sql.DB, scraperInstance *scraper.Scraper) *http.ServeMux {
	mux := http.NewServeMux()
	
	// serve frontend files
	frontendDir := "./frontend"
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, filepath.Join(frontendDir, "index.html"))
			return
		}
		
		// try to serve static file
		filePath := filepath.Join(frontendDir, r.URL.Path)
		if _, err := os.Stat(filePath); err == nil {
			http.ServeFile(w, r, filePath)
			return
		}
		
		// fallback to index.html for SPA routing
		http.ServeFile(w, r, filepath.Join(frontendDir, "index.html"))
	})
	
	// API routes
	mux.HandleFunc("/query", QueryHandler(db))
	mux.HandleFunc("/health", HealthCheckHandler(db))
	mux.HandleFunc("/api/scrapes", ScrapesHandler(db))
	mux.HandleFunc("/api/scrape/bulk", BulkScrapeHandler(db, scraperInstance))
	
	return mux
}

func HealthCheckHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			http.Error(w, "DB not healthy", http.StatusServiceUnavailable)
			return
		}
		fmt.Fprintln(w, "OK")
	}
}

func ScrapesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		rows, err := db.Query(`
			SELECT id, url, max_pages, concurrency, COALESCE(totalresults, 0) as totalresults, completed_at 
			FROM raw_html 
			WHERE completed_at IS NOT NULL 
			ORDER BY completed_at DESC
		`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		
		var scrapes []map[string]interface{}
		for rows.Next() {
			var id, maxPages, concurrency, totalResults int
			var url string
			var completedAt sql.NullTime
			
			if err := rows.Scan(&id, &url, &maxPages, &concurrency, &totalResults, &completedAt); err != nil {
				continue
			}
			
			scrape := map[string]interface{}{
				"id":           id,
				"url":          url,
				"max_pages":    maxPages,
				"concurrency":  concurrency,
				"totalresults": totalResults,
			}
			
			if completedAt.Valid {
				scrape["completed_at"] = completedAt.Time.Format("2006-01-02T15:04:05Z")
			}
			
			scrapes = append(scrapes, scrape)
		}
		
		w.Header().Set("Content-Type", "application/json")
		if len(scrapes) == 0 {
			fmt.Fprint(w, "[]")
			return
		}
		
		// simple JSON marshaling
		fmt.Fprint(w, "[")
		for i, scrape := range scrapes {
			if i > 0 {
				fmt.Fprint(w, ",")
			}
			fmt.Fprintf(w, `{"id":%d,"url":"%s","max_pages":%d,"concurrency":%d,"totalresults":%d,"completed_at":"%s"}`,
				scrape["id"], scrape["url"], scrape["max_pages"], scrape["concurrency"], scrape["totalresults"], scrape["completed_at"])
		}
		fmt.Fprint(w, "]")
	}
}

