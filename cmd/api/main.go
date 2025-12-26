package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	mux.HandleFunc("/contact", app.corsMiddleware(app.contactHandler))

	// Auth routes
	mux.HandleFunc("/register", app.corsMiddleware(app.registerHandler))
	mux.HandleFunc("/login", app.corsMiddleware(app.loginHandler))
	mux.HandleFunc("/logout", app.corsMiddleware(app.logoutHandler))
	mux.HandleFunc("/forgot-password", app.corsMiddleware(app.forgotPasswordHandler))
	mux.HandleFunc("/reset-password", app.corsMiddleware(app.resetPasswordHandler))
	mux.HandleFunc("/me", app.corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			app.authMiddleware(app.updateProfileHandler)(w, r)
		} else {
			app.authMiddleware(app.meHandler)(w, r)
		}
	}))
	mux.HandleFunc("/my-books", app.corsMiddleware(app.authMiddleware(app.userBooksHandler)))
	mux.HandleFunc("/members", app.corsMiddleware(app.authMiddleware(app.listMembersHandler)))
	mux.HandleFunc("/wishlist", app.corsMiddleware(app.authMiddleware(app.getWishlistHandler)))

	// Book routes
	mux.HandleFunc("/books", app.corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			app.authMiddleware(app.createBookHandler)(w, r)
		} else {
			app.listBooksHandler(w, r)
		}
	}))
	mux.HandleFunc("/books/top-requested", app.corsMiddleware(app.listTopRequestedBooksHandler))
	mux.HandleFunc("/genres", app.corsMiddleware(app.listGenresHandler))
	mux.HandleFunc("/genres/popular", app.corsMiddleware(app.listPopularGenresHandler))

	mux.HandleFunc("/books/", app.corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
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
	}))
	mux.HandleFunc("/upload", app.corsMiddleware(app.authMiddleware(app.uploadHandler)))
	mux.HandleFunc("/stats", app.corsMiddleware(app.getStatsHandler))

	// Apply middleware chain: recovery -> logging -> security headers -> routes
	handler := recoverMiddleware(loggingMiddleware(securityHeadersMiddleware(mux)))

	return handler
}

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	// Validate required environment variables
	requiredEnvVars := []string{"DATABASE_URL"}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("%s environment variable is required", envVar)
		}
	}

	// Get configuration from environment
	dsn := os.Getenv("DATABASE_URL")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}

	// Connect to database
	dbConn, err := db.Connect(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

	// Initialize stores
	bookStore := store.NewPostgresBookStore(dbConn)
	if err := bookStore.Migrate(); err != nil {
		log.Fatal(err)
	}

	userStore := store.NewPostgresUserStore(dbConn)
	if err := userStore.Migrate(); err != nil {
		log.Fatal(err)
	}

	requestStore := store.NewPostgresRequestStore(dbConn)

	// Initialize email service
	var emailService email.EmailService
	resendAPIKey := os.Getenv("RESEND_API_KEY")
	if resendAPIKey != "" {
		emailService = email.NewResendEmailService(resendAPIKey)
		log.Println("✓ Using Resend email service")
	} else {
		emailService = email.NewConsoleEmailService()
		log.Println("⚠ Using Console email service (set RESEND_API_KEY to use Resend)")
	}

	// Initialize storage service
	var storageService storage.Service
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if supabaseURL != "" && supabaseKey != "" {
		storageService = storage.NewSupabaseStorage(supabaseURL, supabaseKey, "uploads")
		log.Println("✓ Using Supabase storage service")
	} else {
		storageService = storage.NewLocalStorage("uploads")
		log.Println("⚠ Using Local storage service (set SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY for cloud storage)")
	}

	// Create application
	app := &application{
		bookStore:      bookStore,
		userStore:      userStore,
		requestStore:   requestStore,
		emailService:   emailService,
		storageService: storageService,
	}

	// Create server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: app.routes(),
	}

	// Channel to listen for errors from the server
	serverErrors := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		log.Printf("🚀 Server starting on port %s", port)
		serverErrors <- srv.ListenAndServe()
	}()

	// Channel to listen for interrupt signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or an error
	select {
	case err := <-serverErrors:
		log.Fatalf("Error starting server: %v", err)

	case sig := <-shutdown:
		log.Printf("Received signal %v, starting graceful shutdown", sig)

		// Create a context with timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Error during shutdown: %v", err)
			// Force close after timeout
			srv.Close()
		}

		log.Println("Server stopped gracefully")
	}
}
