package store

import (
	"database/sql"
	"strconv"
	"time"
)

type Book struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	Description string    `json:"description"`
	Genre       string    `json:"genre"`
	ImagePath   string    `json:"image_path"`
	CreatedAt   time.Time `json:"created_at"`
	UserID         int       `json:"user_id"`
	UserEmail      string    `json:"user_email,omitempty"`       // For display purposes
	UserUsername   string    `json:"user_username,omitempty"`    // For display purposes
	UserAvatarPath string    `json:"user_avatar_path,omitempty"` // For display purposes
	IsRequested    bool      `json:"is_requested"`
}

type BookFilter struct {
	Query  string // Search by title or author
	Genre  string // Filter by genre
	Sort   string // "newest" or "oldest"
	Limit  int
	Offset int
}

type BookStorer interface {
	Add(book Book) (Book, error)
	GetAll(filter BookFilter) ([]Book, error)
	GetByID(id int) (Book, error)
	GetByUserID(userID int) ([]Book, error)
	Update(book Book) error
	Delete(id int) error
	GetGenres() ([]string, error)
	GetPopularGenres() ([]GenreStats, error)
}

type PostgresBookStore struct {
	db *sql.DB
}

func NewPostgresBookStore(db *sql.DB) *PostgresBookStore {
	return &PostgresBookStore{
		db: db,
	}
}

func (s *PostgresBookStore) Migrate() error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS books (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			author TEXT NOT NULL,
			description TEXT,
			genre TEXT,
			image_path TEXT,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			user_id INTEGER REFERENCES users(id)
		);
		ALTER TABLE books ADD COLUMN IF NOT EXISTS description TEXT;
		ALTER TABLE books ADD COLUMN IF NOT EXISTS genre TEXT;
		ALTER TABLE books ADD COLUMN IF NOT EXISTS image_path TEXT;
		ALTER TABLE books ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;
		ALTER TABLE books ADD COLUMN IF NOT EXISTS user_id INTEGER REFERENCES users(id);
		`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresBookStore) Add(book Book) (Book, error) {
	query := `
		INSERT INTO books (title, author, description, genre, image_path, user_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	err := s.db.QueryRow(query, book.Title, book.Author, book.Description, book.Genre, book.ImagePath, book.UserID).Scan(&book.ID, &book.CreatedAt)
	if err != nil {
		return Book{}, err
	}

	return book, nil
}

func (s *PostgresBookStore) GetAll(filter BookFilter) ([]Book, error) {
	query := `
		SELECT b.id, b.title, b.author, COALESCE(b.description, ''), COALESCE(b.genre, ''), COALESCE(b.image_path, ''), b.created_at, b.user_id, COALESCE(u.email, ''), COALESCE(u.username, ''), COALESCE(u.avatar_path, '')
		FROM books b
		LEFT JOIN users u ON b.user_id = u.id
		WHERE 1=1`
	var args []interface{}

	if filter.Query != "" {
		query += ` AND (b.title ILIKE $` + strconv.Itoa(len(args)+1) + ` OR b.author ILIKE $` + strconv.Itoa(len(args)+1) + `)`
		args = append(args, "%"+filter.Query+"%")
	}

	if filter.Genre != "" {
		query += ` AND b.genre = $` + strconv.Itoa(len(args)+1)
		args = append(args, filter.Genre)
	}

	switch filter.Sort {
	case "oldest":
		query += ` ORDER BY b.created_at ASC`
	case "newest":
		query += ` ORDER BY b.created_at DESC`
	default:
		query += ` ORDER BY b.created_at DESC`
	}

	if filter.Limit > 0 {
		query += ` LIMIT $` + strconv.Itoa(len(args)+1)
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		query += ` OFFSET $` + strconv.Itoa(len(args)+1)
		args = append(args, filter.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var b Book
		var userID sql.NullInt64 // Handle nullable user_id for existing records
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Description, &b.Genre, &b.ImagePath, &b.CreatedAt, &userID, &b.UserEmail, &b.UserUsername, &b.UserAvatarPath); err != nil {
			return nil, err
		}
		if userID.Valid {
			b.UserID = int(userID.Int64)
		}
		books = append(books, b)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return books, nil
}
