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
)

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

    repos := setupRepositories(db)

    mux := http.NewServeMux()
    setupRoutes(mux, repos)

    mux.HandleFunc("/health", healthCheckHandler)

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

func initDatabase() (*sql.DB, error) {
    connStr := "postgres://placeholder" //placeholder
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, err
    }
    if err := db.Ping(); err != nil {
        return nil, err
    }
    return db, nil
}

func migrateDatabase(db *sql.DB) error {
    _, err := db.Exec(`CREATE TABLE IF NOT EXISTS migrations (id SERIAL PRIMARY KEY, name TEXT)`)
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