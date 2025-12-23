package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (app *application) forgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	user, err := app.userStore.GetByEmail(input.Email)
	if err != nil {
		// Don't reveal if user exists
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "If an account exists, a reset email has been sent."})
		return
	}

	// Generate token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	token := hex.EncodeToString(b)

	// Save token (valid for 1 hour)
	expiry := time.Now().Add(1 * time.Hour)
	if err := app.userStore.SaveResetToken(token, user.ID, expiry); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Send email
	app.emailService.SendPasswordReset(user.Email, token)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "If an account exists, a reset email has been sent."})
}

func (app *application) resetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	userID, expiry, err := app.userStore.GetResetToken(input.Token)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusBadRequest)
		return
	}

	if time.Now().After(expiry) {
		app.userStore.DeleteResetToken(input.Token)
		http.Error(w, "Token expired", http.StatusBadRequest)
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := app.userStore.UpdatePassword(userID, string(hashedPassword)); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Delete token
	app.userStore.DeleteResetToken(input.Token)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Password updated successfully"})
}
