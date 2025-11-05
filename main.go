package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"webScraper/database"
	"webScraper/handler"
	"webScraper/scraper"

	_ "github.com/lib/pq"
)

/* init */
var db *sql.DB

func main() {
    var err error

    /* db init and migrate stuff */
    db, err = database.InitDatabase()
    if err != nil {
        log.Fatalf("DB init error: %v", err)
    }
    
    scraper := scraper.NewScraper(5, 10, "MyUserAgent", db)

    if err := database.MigrateDatabase(db); err != nil {
    log.Fatalf("Migration error: %v", err)
    }
    
    if err := database.MigrateParsedResults(db); err != nil {
        log.Fatalf("Migration error: %v", err)
    }

    /* route handling */
    mux := handler.SetupRoutes(db, scraper)

    /* server config */
    srv := &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }

    go func() {
        log.Printf("Server runs on %s", srv.Addr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("ListenAndServe: %v", err)
        }
    }()

    log.Println("Server started. Use the web interface to upload URLs and start scraping.")
    log.Println("Open http://localhost:8080 in your browser.")

    // Uncomment below to run scraping on startup from urls.txt
    /*
    urls, err := scanner.ReadURLsFromFile("urls.txt")
    if err != nil {
        log.Printf("Warning: Could not read urls.txt: %v", err)
    } else {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
        defer cancel()

        const maxDepth = 3
        var wg sync.WaitGroup
        for i, url := range urls {
            wg.Add(1)
            go func(i int, url string) {
                defer wg.Done()
                scraper.ScrapeWithDepth(ctx, url, maxDepth)
                fmt.Printf("Global progress: %d/%d URLs done\n", i+1, len(urls))
            }(i, url)
        }
        wg.Wait()

        keyword := "price"
        parseFunc := func(url string, html []byte) (string, string, sql.NullTime) {
            price, title, completedAt := parser.ParseFunc(url, html, keyword)
            return price, title, sql.NullTime{Time: completedAt, Valid: true}
        }

        postgres := &database.Postgres{DB: db}
        if err := postgres.ProcessRawHTML(parseFunc); err != nil {
            log.Printf("Parsing error: %v", err)
        }
    }
    */

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
    <-quit
    log.Println("Shutdown...")
    ctxServer, cancelServer := context.WithTimeout(context.Background(), 5*time.Second) 
    defer cancelServer()
    if err := srv.Shutdown(ctxServer); err != nil {
        log.Fatalf("Server Shutdown: %v", err)
    }
    log.Println("Server stopped.")
}
