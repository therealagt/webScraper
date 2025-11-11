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

	scraper := scraper.NewScraper(5, 10, "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36", db)

	if err := database.MigrateDatabase(db); err != nil {
		log.Fatalf("Migration error: %v", err)
	}

	if err := database.MigrateParsedResults(db); err != nil {
		log.Fatalf("Migration error: %v", err)
	}

	if err := database.MigrateRawHTML(db); err != nil {
		log.Fatalf("RawHTML Migration error: %v", err)
	}

	// global context for shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	/* route handling */
	mux := handler.SetupRoutes(db, scraper, ctx)

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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown...")

	// cancel context to stop all scrapers
	cancel()

	// wait briefly to allow scrapers to stop gracefully
	time.Sleep(2 * time.Second)

	ctxServer, cancelServer := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelServer()
	if err := srv.Shutdown(ctxServer); err != nil {
		log.Fatalf("Server Shutdown: %v", err)
	}
	log.Println("Server stopped.")
}
