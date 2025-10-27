package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"webScraper/database"
	"webScraper/handler"
	"webScraper/scanner"
	"webScraper/scraper"

	_ "github.com/lib/pq"
)

/* Init */
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

    urls, err := scanner.ReadURLsFromFile("urls.txt")
    if err != nil {
        log.Fatalf("Error reading the urls: %v", err)
    }

    /* time ctx for go routines */
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    /* route handling */
    mux := handler.SetupRoutes(db)

    /* server config */
    srv := &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }

     /* create go routine for server parallel to scraping */
    go func() {
        log.Printf("Server runs on %s", srv.Addr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("ListenAndServe: %v", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
    <-quit
    log.Println("Shutdown...")
    ctxServer, cancel := context.WithTimeout(context.Background(), 5*time.Second) 
    defer cancel()
    if err := srv.Shutdown(ctxServer); err != nil {
        log.Fatalf("Server Shutdown: %v", err)
    }
    log.Println("Server stopped.")

    /* create go routine for scraper */
     var wg sync.WaitGroup
    for i, url := range urls {
        wg.Add(1)
        go func(i int, url string) {
            defer wg.Done()
            scraper.Scrape(ctx, url, 1)
            fmt.Printf("Global progress: %d/%d URLs done\n", i+1, len(urls))
        }(i, url)

        parseFunc := func(url string, html []byte) (string, string, time.Time) {
            // hier brauche ich parsing logic
            return "", url, time.Now()
        }

        if err := database.ProcessRawHTML(parseFunc); err != nil { //kp was hier geht
            log.Fatalf("Parsing error: %v", err)
        }
    }
    wg.Wait()
}
