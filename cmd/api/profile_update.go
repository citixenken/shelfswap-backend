package main

import (
	"encoding/json"
	"net/http"
)

func (app *application) updateProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB limit
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("userID").(int)
	user, err := app.userStore.GetByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	username := r.FormValue("username")
	bio := r.FormValue("bio")
	location := r.FormValue("location")
	removeAvatar := r.FormValue("remove_avatar")

	// Handle avatar upload
	file, header, err := r.FormFile("avatar")
	var avatarPath string
	if err == nil {
		defer file.Close()

		// Upload to storage service
		path, err := app.storageService.Upload(file, header)
		if err != nil {
			http.Error(w, "Failed to save avatar", http.StatusInternalServerError)
			return
		}
		avatarPath = path
	}

	// Update user fields
	if username != "" {
		user.Username = username
	}
	user.Bio = bio
	user.Location = location

	if removeAvatar == "true" {
		user.AvatarPath = ""
	} else if avatarPath != "" {
		user.AvatarPath = avatarPath
	}

	if err := app.userStore.Update(user); err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	// Return updated user
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
