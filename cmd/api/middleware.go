package main

import (
	"log"
	"net/http"
	"strings"
	"time"
)

// corsMiddleware adds CORS headers to allow frontend access
func (app *application) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get allowed origins from environment or use defaults
		allowedOrigins := []string{
			"http://localhost:5173",           // Local development
			"http://localhost:3000",           // Alternative local port
			"https://*.pages.dev",             // Cloudflare Pages preview
			"https://shelfswap.io",            // Production custom domain
			"https://www.shelfswap.io",        // Production custom domain (www)
		}

		origin := r.Header.Get("Origin")
		allowed := false

		for _, allowedOrigin := range allowedOrigins {
			if strings.Contains(allowedOrigin, "*") {
				// Handle wildcard subdomains
				// Example: https://*.pages.dev
				// We want to match: https://anything.pages.dev

				// 1. Check if it starts with the same protocol
				parts := strings.Split(allowedOrigin, "*")
				if len(parts) != 2 {
					continue
				}
				prefix := parts[0] // e.g., "https://"
				suffix := parts[1] // e.g., ".pages.dev"

				if strings.HasPrefix(origin, prefix) && strings.HasSuffix(origin, suffix) {
					allowed = true
					break
				}
			} else if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// securityHeadersMiddleware adds security headers to all responses
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// Enable XSS protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Content Security Policy
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:;")

		// Referrer Policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// HSTS (only in production with HTTPS)
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs all incoming requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture status code
		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(lrw, r)

		duration := time.Since(start)
		log.Printf(
			"%s %s %d %s %s",
			r.Method,
			r.RequestURI,
			lrw.statusCode,
			duration,
			r.RemoteAddr,
		)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// recoverMiddleware recovers from panics and logs the error
func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
