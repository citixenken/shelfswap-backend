package main

import (
	"encoding/json"
	"net/http"
)

func (app *application) listMembersHandler(w http.ResponseWriter, r *http.Request) {
	members, err := app.userStore.GetMembers()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

func (app *application) getWishlistHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)

	requests, err := app.requestStore.GetRequestsByUserID(userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}
