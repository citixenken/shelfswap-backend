package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testbook-backend/internal/store"
)

func (app *application) requestBookHandler(w http.ResponseWriter, r *http.Request) {
	// Extract book ID from URL /books/{id}/request
	pathParts := strings.Split(r.URL.Path, "/")
	// Expected path: /books/{id}/request
	// Split: ["", "books", "{id}", "request"]
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	idStr := pathParts[2]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	book, err := app.bookStore.GetByID(id)
	if err != nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	userID := r.Context().Value("userID").(int)
	if book.UserID == userID {
		http.Error(w, "Cannot request your own book", http.StatusBadRequest)
		return
	}

	requester, err := app.userStore.GetByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	owner, err := app.userStore.GetByID(book.UserID)
	if err != nil {
		http.Error(w, "Owner not found", http.StatusInternalServerError)
		return
	}

	// Send email notification first
	ownerName := owner.Username
	if ownerName == "" {
		ownerName = "there"
	}
	if err := app.emailService.SendRequestNotification(owner.Email, ownerName, book.Title, requester.Email); err != nil {
		http.Error(w, "Failed to send email notification", http.StatusInternalServerError)
		return
	}

	// Save request to DB only if email sent successfully
	req := store.BookRequest{
		BookID:      book.ID,
		RequesterID: requester.ID,
	}
	if err := app.requestStore.AddRequest(req); err != nil {
		http.Error(w, "Failed to save request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Request sent successfully."})
}

func (app *application) listTopRequestedBooksHandler(w http.ResponseWriter, r *http.Request) {
	limit := 5
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	books, err := app.requestStore.GetTopRequestedBooks(limit)
	if err != nil {
		http.Error(w, "Failed to fetch top requested books", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func (app *application) deleteBookRequestHandler(w http.ResponseWriter, r *http.Request) {
	// Extract book ID from URL /books/{id}/request
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	idStr := pathParts[2]
	bookID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("userID").(int)

	if err := app.requestStore.DeleteRequest(userID, bookID); err != nil {
		http.Error(w, "Failed to delete request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
