package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

    scraper := scraper.NewScraper(5, 10, "MyUserAgent", db)
    scraper.Scrape("https://de.wikipedia.org/wiki/Webscraping", 100)
    
    repos := setupRepositories(db)

    mux := http.NewServeMux()
    setupRoutes(mux, repos)

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
    ctx, cancel := context.WithTimeout(context.Background(), 5_000_000_000) 
    defer cancel()
    if err := srv.Shutdown(ctx); err != nil {
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
    _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS raw_html (
            id SERIAL PRIMARY KEY,
            url TEXT,
            max_pages INT, 
            concurrency INT,
            html BYTEA,
            totalResults INT,
            scraped_at TIMESTAMP DEFAULT NOW(),
            completed_at TIMESTAMP
        )
    `)
    return err
}

type Repositories struct {
}

func setupRepositories(db *sql.DB) *Repositories {
    return &Repositories{}
}

func setupRoutes(mux *http.ServeMux, repos *Repositories) {
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Hello, World!")
    })
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
    if err := db.Ping(); err != nil {
        http.Error(w, "DB not healthy", http.StatusServiceUnavailable)
        return
    }
    fmt.Fprintln(w, "OK")
}