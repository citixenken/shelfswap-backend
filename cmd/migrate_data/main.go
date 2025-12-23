package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Connect to DB
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Connected to database successfully")

	// Truncate tables to avoid duplicates
	log.Println("Truncating tables...")
	_, err = db.Exec("TRUNCATE users, books, book_requests, password_resets RESTART IDENTITY CASCADE")
	if err != nil {
		// Ignore error if tables don't exist (though they should)
		log.Println("Warning: Failed to truncate tables (might be empty or not exist):", err)
	}

	// Read dump file
	dumpContent, err := os.ReadFile("data_dump.sql")
	if err != nil {
		log.Fatal("Failed to read data_dump.sql:", err)
	}

	// Execute SQL
	log.Println("Executing SQL dump...")
	_, err = db.Exec(string(dumpContent))
	if err != nil {
		log.Fatal("Failed to execute SQL dump:", err)
	}

	log.Println("Data migration completed successfully!")
}
