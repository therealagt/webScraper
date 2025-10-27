package handler

import (
	"database/sql"
	"fmt"
	"net/http"
)

/* route handling */
func SetupRoutes(db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	})
	mux.HandleFunc("/query", QueryHandler(db))
	mux.HandleFunc("/health", HealthCheckHandler(db))
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

