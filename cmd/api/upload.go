package main

import (
	"encoding/json"
	"net/http"
)

func (app *application) uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit upload size (e.g., 10MB)
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Upload to storage service
	imagePath, err := app.storageService.Upload(file, handler)
	if err != nil {
		http.Error(w, "Failed to upload image", http.StatusInternalServerError)
		return
	}

	// Return the path (or full URL)
	response := map[string]string{
		"image_path": imagePath,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
