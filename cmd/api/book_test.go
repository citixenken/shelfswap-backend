package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"testbook-backend/internal/store"
)

func TestCreateBook(t *testing.T) {
	// Initialize store and application
	bookStore := store.NewInMemoryBookStore()
	app := &application{
		bookStore: bookStore,
	}

	// Create a new book payload
	payload := map[string]string{
		"title":  "The Go Programming Language",
		"author": "Alan A. A. Donovan",
	}
	body, _ := json.Marshal(payload)

	// Create request
	req, _ := http.NewRequest("POST", "/books", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call the handler (we need to define routes first, but for unit test we can call handler directly if exposed,
	// or better, setup the router like in main)

	// For this test, let's assume we have a routes() method or similar,
	// or we can just test the handler function if we export it.
	// Let's refactor main to make it testable.

	handler := app.routes()
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	// Check response body
	var response store.Book
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if response.Title != payload["title"] {
		t.Errorf("handler returned unexpected title: got %v want %v",
			response.Title, payload["title"])
	}
	if response.ID == 0 {
		t.Error("handler returned 0 ID, expected generated ID")
	}
}
