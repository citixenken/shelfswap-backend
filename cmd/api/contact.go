package main

import (
	"encoding/json"
	"net/http"
)

func (app *application) contactHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if input.Name == "" || input.Email == "" || input.Subject == "" || input.Message == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	if err := app.emailService.SendContactEmail(input.Email, input.Subject, input.Message); err != nil {
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Message sent successfully"})
}
