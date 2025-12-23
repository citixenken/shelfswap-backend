package main

import (
	"encoding/json"
	"net/http"
	"testbook-backend/internal/store"
)

type Stats struct {
	TotalBooks  int `json:"total_books"`
	ActiveUsers int `json:"active_users"`
	TotalGenres int `json:"total_genres"`
}

func (app *application) getStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats := Stats{}

	// Get total books count - use GetAll with empty filter
	books, err := app.bookStore.GetAll(store.BookFilter{Limit: 999999})
	if err == nil {
		stats.TotalBooks = len(books)

		// Count unique users from books
		userMap := make(map[int]bool)
		genreMap := make(map[string]bool)

		for _, book := range books {
			if book.UserID != 0 {
				userMap[book.UserID] = true
			}
			if book.Genre != "" {
				genreMap[book.Genre] = true
			}
		}

		stats.ActiveUsers = len(userMap)
		stats.TotalGenres = len(genreMap)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
