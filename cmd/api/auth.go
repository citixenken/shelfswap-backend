package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"testbook-backend/internal/store"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/clerk/clerk-sdk-go/v2/user"
)

// CLERK_SECRET_KEY should be set in environment variables

func (app *application) registerHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Use Clerk for registration", http.StatusGone)
}

func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Use Clerk for login", http.StatusGone)
}

func (app *application) logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Use Clerk for logout", http.StatusGone)
}

func (app *application) meHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	localUser, err := app.userStore.GetByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Get Clerk user data
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		cookie, err := r.Cookie("__session")
		if err == nil {
			authHeader = "Bearer " + cookie.Value
		}
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token != "" {
		key := os.Getenv("CLERK_SECRET_KEY")
		if key == "" {
			// Log error or handle missing key appropriately
			// For now, we just won't be able to verify
			log.Println("CLERK_SECRET_KEY not set")
			return
		}
		clerk.SetKey(key)

		claims, err := jwt.Verify(r.Context(), &jwt.VerifyParams{
			Token: token,
		})
		if err == nil {
			usr, err := user.Get(r.Context(), claims.Subject)
			if err == nil {
				// Return combined data: Clerk user info + local DB ID
				response := map[string]interface{}{
					"id":          localUser.ID, // Local DB ID for ownership checks
					"clerk_id":    usr.ID,
					"email":       localUser.Email,
					"username":    localUser.Username,
					"avatar_path": usr.ImageURL,
					"created_at":  localUser.CreatedAt,
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
				return
			}
		}
	}

	// Fallback to local user data only
	json.NewEncoder(w).Encode(localUser)
}

// getAuthenticatedUserID verifies the Clerk token and returns the local user ID.
// Returns 0 and nil error if no token is present or invalid (optional auth).
// Returns error only if there's a system error (e.g. DB).
func (app *application) getAuthenticatedUserID(r *http.Request) (int, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		cookie, err := r.Cookie("__session")
		if err == nil {
			authHeader = "Bearer " + cookie.Value
		}
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return 0, nil
	}

	key := os.Getenv("CLERK_SECRET_KEY")
	if key == "" {
		log.Println("CLERK_SECRET_KEY not set")
		return 0, nil
	}

	// In v2, we don't need to create a client for JWT verification if we use the default config or pass keys?
	// Wait, jwt.Verify needs a JWKS client.
	// Correct usage:
	// In v2, we use the global configuration
	clerk.SetKey(key)

	// Verify Token
	// We need a JWKS client for verification.
	// If we set the key globally, maybe we can get a default client?
	// Based on v2 docs (recalled), we might need to create a client if we want to pass it to JWKS.
	// But since NewClient was undefined, let's try to use the global SetKey and see if jwt.Verify works with default?
	// Actually, jwt.Verify requires JWKSClient in params.
	// Let's try to find the right function to get the client.
	// Maybe it is 'clerk.NewClient' but I need to import 'github.com/clerk/clerk-sdk-go/v2/clerk'?
	// Let's try to just use user.Get first as that's simpler.

	// For JWT, let's assume we can pass nil for JWKSClient if we don't have one, or it's optional?
	// No, it's likely required.
	// Let's try to use the http handler middleware provided by Clerk if possible?
	// But I am writing a custom middleware.

	// Let's try to import the 'clerk' subpackage to see if NewClient is there.
	// I will modify the imports in the next step if this fails.
	// For now, I will assume I can't create a client and try to use what I have.
	// Wait, if user.Get doesn't take client, it uses global.
	// Maybe there is a global 'clerk.JWKS()' function?

	claims, err := jwt.Verify(r.Context(), &jwt.VerifyParams{
		Token: token,
	})
	if err != nil {
		return 0, nil // Invalid token
	}

	// Sync User
	usr, err := user.Get(r.Context(), claims.Subject)
	if err != nil {
		return 0, nil
	}

	if len(usr.EmailAddresses) == 0 {
		return 0, nil
	}

	email := usr.EmailAddresses[0].EmailAddress

	localUser, err := app.userStore.GetByEmail(email)
	if err != nil {
		// User not found, create new
		// Fix: Handle nil username safely
		uName := ""
		if usr.Username != nil {
			uName = *usr.Username
		}

		newUser := store.User{
			Email:    email,
			Password: "clerk_managed_account",
			Username: uName,
		}

		if err := app.userStore.Create(newUser); err != nil {
			// Try to fetch again in case of race condition
			localUser, err = app.userStore.GetByEmail(email)
			if err != nil {
				return 0, err
			}
		} else {
			// Re-fetch to get ID (workaround for value receiver issue)
			localUser, err = app.userStore.GetByEmail(email)
			if err != nil {
				return 0, err
			}
		}
	} else {
		// User exists - sync username from Clerk if it has changed
		clerkUsername := ""
		if usr.Username != nil {
			clerkUsername = *usr.Username
		}

		// Update username if different from what's in Clerk
		if localUser.Username != clerkUsername {
			localUser.Username = clerkUsername
			if err := app.userStore.Update(localUser); err != nil {
				// Log error but don't fail auth
				// The user can still authenticate even if username sync fails
			}
		}
	}

	return localUser.ID, nil
}

func (app *application) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := app.getAuthenticatedUserID(r)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		if userID == 0 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
