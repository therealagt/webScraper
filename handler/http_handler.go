package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"webScraper/database"
)

func QueryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		links := r.URL.Query()["link"]
		keywords := r.URL.Query()["keyword"]
		dateFrom := r.URL.Query().Get("from")
		dateTo := r.URL.Query().Get("to")
		sortBy := r.URL.Query().Get("sort")
		limit := 10
		if l := r.URL.Query().Get("limit"); l != "" {
			fmt.Sscanf(l, "%d", &limit)
		}

		qb := database.QueryBuilder{
			Links:    links,
			Keywords: keywords,
			DateFrom: dateFrom,
			DateTo:   dateTo,
			SortBy:   sortBy,
			Limit:    limit,
		}

		query := qb.BuildRawHTMLQuery()
			fmt.Println("SQL Query:", query)
		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, "Query error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var id int
			var url string
			if err := rows.Scan(&id, &url); err != nil {
				http.Error(w, "Scan error", http.StatusInternalServerError)
				return
			}
			fmt.Fprintln(w, "ID:", id, "URL:", url)
		}
	}
}