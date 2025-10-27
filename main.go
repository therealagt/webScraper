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

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

/* Init */
var db *sql.DB

func main() {
    var err error
    db, err = initDatabase()
    if err != nil {
        log.Fatalf("DB init error: %v", err)
    }
    
    if err := migrateDatabase(db); err != nil {
    log.Fatalf("Migration error: %v", err)
    }
    
    if err := database.MigrateParsedResults(db); err != nil {
        log.Fatalf("Migration error: %v", err)
    }

    urls, err := scanner.ReadURLsFromFile("urls.txt")
    if err != nil {
        log.Fatalf("Error reading the urls: %v", err)
    }

    scraper := scraper.NewScraper(5, 10, "MyUserAgent", db)

    ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
    defer cancel()

    var wg sync.WaitGroup
    for i, url := range urls {
        wg.Add(1)
        go func(i int, url string) {
            defer wg.Done()
            scraper.Scrape(ctx, url, 1)
            fmt.Printf("Global progress: %d/%d URLs done\n", i+1, len(urls))
        }(i, url)
    }
    wg.Wait()

    mux := http.NewServeMux()
    setupRoutes(mux)

    mux.HandleFunc("/health", healthCheckHandler)

    srv := &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }

    /* create go routine */
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
}

/* Init db */
func initDatabase() (*sql.DB, error) {
    godotenv.Load()

    host := os.Getenv("DB_HOST")
    port := os.Getenv("DB_PORT")
    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASSWORD")
    dbname := os.Getenv("DB_NAME")

    connStr := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        host, port, user, password, dbname,
    )
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, err
    }
    if err := db.Ping(); err != nil {
        return nil, err
    }
    return db, nil
}

/* Helper functions */
func migrateDatabase(db *sql.DB) error {
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

func setupRoutes(mux *http.ServeMux) {
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Hello, World!")
    })
    mux.HandleFunc("/query", handler.QueryHandler(db))
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
    if err := db.Ping(); err != nil {
        http.Error(w, "DB not healthy", http.StatusServiceUnavailable)
        return
    }
    fmt.Fprintln(w, "OK")
}

