package main

import (
	"encoding/json"
	"net/http"
)

func (app *application) listGenresHandler(w http.ResponseWriter, r *http.Request) {
	genres, err := app.bookStore.GetGenres()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(genres)
}
