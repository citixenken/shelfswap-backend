package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"testbook-backend/internal/db"
	"testbook-backend/internal/email"
	"testbook-backend/internal/storage"
	"testbook-backend/internal/store"
)

type application struct {
	bookStore    store.BookStorer
	userStore    store.UserStore
	requestStore store.RequestStore
	emailService email.EmailService
	storageService storage.Service
}

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "🚀"})
	})

	mux.HandleFunc("/contact", app.contactHandler)

	// Auth routes
	mux.HandleFunc("/register", app.registerHandler)
	mux.HandleFunc("/login", app.loginHandler)
	mux.HandleFunc("/logout", app.logoutHandler)
	mux.HandleFunc("/forgot-password", app.forgotPasswordHandler)
	mux.HandleFunc("/reset-password", app.resetPasswordHandler)
	mux.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			app.authMiddleware(app.updateProfileHandler)(w, r)
		} else {
			app.authMiddleware(app.meHandler)(w, r)
		}
	})
	mux.HandleFunc("/my-books", app.authMiddleware(app.userBooksHandler))
	mux.HandleFunc("/members", app.authMiddleware(app.listMembersHandler))
	mux.HandleFunc("/wishlist", app.authMiddleware(app.getWishlistHandler))

	// Book routes
	mux.HandleFunc("/books", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			app.authMiddleware(app.createBookHandler)(w, r)
		} else {
			app.listBooksHandler(w, r)
		}
	})
	mux.HandleFunc("/books/top-requested", app.listTopRequestedBooksHandler)
	mux.HandleFunc("/genres", app.listGenresHandler)
	mux.HandleFunc("/genres/popular", app.listPopularGenresHandler)

	mux.HandleFunc("/books/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut || r.Method == http.MethodDelete {
			app.authMiddleware(app.bookIDHandler)(w, r)
		} else if len(r.URL.Path) > len("/books/") && r.URL.Path[len(r.URL.Path)-8:] == "/request" {
			if r.Method == http.MethodPost {
				app.authMiddleware(app.requestBookHandler)(w, r)
			} else if r.Method == http.MethodDelete {
				app.authMiddleware(app.deleteBookRequestHandler)(w, r)
			}
		} else {
			app.bookIDHandler(w, r)
		}
	})
	mux.HandleFunc("/upload", app.authMiddleware(app.uploadHandler))
	mux.HandleFunc("/stats", app.getStatsHandler)

	// Serve static files
	// Serve static files (only if using local storage, but for simplicity we remove it as we migrate to cloud)
	// If using LocalStorage in dev, you might want to keep this or serve via a separate handler.
	// For this migration, we assume Supabase or we'll add it back if LocalStorage is active.
	// Actually, let's keep it conditionally or just remove it if we are fully committing to Supabase.
	// Given the user wants to migrate, let's remove it to force usage of the new system.
	// But wait, existing images on disk won't be served. That's fine for a migration step.

	return mux
}

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	// Default DSN, can be overridden by env var in real app
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	dbConn, err := db.Connect(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

	bookStore := store.NewPostgresBookStore(dbConn)
	if err := bookStore.Migrate(); err != nil {
		log.Fatal(err)
	}

	userStore := store.NewPostgresUserStore(dbConn)
	if err := userStore.Migrate(); err != nil {
		log.Fatal(err)
	}

	var emailService email.EmailService
	resendAPIKey := os.Getenv("RESEND_API_KEY")
	if resendAPIKey != "" {
		emailService = email.NewResendEmailService(resendAPIKey)
		log.Println("Using Resend email service")
	} else {
		emailService = email.NewConsoleEmailService()
		log.Println("Using Console email service (set RESEND_API_KEY to use Resend)")
	}
	requestStore := store.NewPostgresRequestStore(dbConn)

	var storageService storage.Service
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if supabaseURL != "" && supabaseKey != "" {
		storageService = storage.NewSupabaseStorage(supabaseURL, supabaseKey, "uploads")
		log.Println("Using Supabase storage service")
	} else {
		storageService = storage.NewLocalStorage("uploads")
		log.Println("Using Local storage service")
	}

	app := &application{
		bookStore:    bookStore,
		userStore:    userStore,
		requestStore: requestStore,
		emailService: emailService,
		storageService: storageService,
	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: app.routes(),
	}

	log.Println("Server starting on :8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
