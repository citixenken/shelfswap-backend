package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"testbook-backend/internal/store"
)



func (app *application) bookIDHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		app.getBookHandler(w, r)
	case http.MethodPut:
		app.updateBookHandler(w, r)
	case http.MethodDelete:
		app.deleteBookHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *application) listBooksHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	genre := r.URL.Query().Get("genre")
	sortParam := r.URL.Query().Get("sort")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 9 // Default limit
	}

	filter := store.BookFilter{
		Query:  query,
		Genre:  genre,
		Sort:   sortParam,
		Limit:  limit,
		Offset: offset,
	}

	books, err := app.bookStore.GetAll(filter)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Privacy: Redact details for unauthenticated users
	// Privacy: Redact details for unauthenticated users
	userID, _ := app.getAuthenticatedUserID(r)
	isAuthenticated := userID != 0

	if !isAuthenticated {
		for i := range books {
			books[i].UserEmail = ""
			books[i].UserUsername = ""
			books[i].UserAvatarPath = ""
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func (app *application) createBookHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title       string `json:"title"`
		Author      string `json:"author"`
		Description string `json:"description"`
		Genre       string `json:"genre"`
		ImagePath   string `json:"image_path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("userID").(int)

	book := store.Book{
		Title:       input.Title,
		Author:      input.Author,
		Description: input.Description,
		Genre:       input.Genre,
		ImagePath:   input.ImagePath,
		UserID:      userID,
	}

	createdBook, err := app.bookStore.Add(book)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdBook)
}

func (app *application) getBookHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Path[len("/books/"):])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	book, err := app.bookStore.GetByID(id)
	if err != nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Privacy: Redact details for unauthenticated users
	userID, err := app.getAuthenticatedUserID(r)
	isAuthenticated := userID != 0

	if !isAuthenticated {
		book.UserEmail = ""
		book.UserUsername = ""
		book.UserAvatarPath = ""
	} else {
		hasRequested, err := app.requestStore.HasRequested(userID, book.ID)
		if err == nil {
			book.IsRequested = hasRequested
		}
	}

	json.NewEncoder(w).Encode(book)
}

func (app *application) userBooksHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)

	books, err := app.bookStore.GetByUserID(userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func (app *application) updateBookHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Path[len("/books/"):])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	existingBook, err := app.bookStore.GetByID(id)
	if err != nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	userID := r.Context().Value("userID").(int)
	if existingBook.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var input struct {
		Title       string `json:"title"`
		Author      string `json:"author"`
		Description string `json:"description"`
		Genre       string `json:"genre"`
		ImagePath   string `json:"image_path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	book := store.Book{
		ID:          id,
		Title:       input.Title,
		Author:      input.Author,
		Description: input.Description,
		Genre:       input.Genre,
		ImagePath:   input.ImagePath,
		UserID:      userID,
	}

	if err := app.bookStore.Update(book); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}

func (app *application) deleteBookHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Path[len("/books/"):])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	existingBook, err := app.bookStore.GetByID(id)
	if err != nil {
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	userID := r.Context().Value("userID").(int)
	if existingBook.UserID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := app.bookStore.Delete(id); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) listPopularGenresHandler(w http.ResponseWriter, r *http.Request) {
	genres, err := app.bookStore.GetPopularGenres()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(genres)
}
