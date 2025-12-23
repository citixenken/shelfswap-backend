package store

import (
	"database/sql"
	"time"
)

type BookRequest struct {
	ID          int       `json:"id"`
	BookID      int       `json:"book_id"`
	RequesterID int       `json:"requester_id"`
	CreatedAt   time.Time `json:"created_at"`
	BookTitle   string    `json:"book_title,omitempty"`
	BookAuthor  string    `json:"book_author,omitempty"`
	BookImage   string    `json:"book_image,omitempty"`
}

type RequestStore interface {
	AddRequest(req BookRequest) error
	GetRequestsByUserID(userID int) ([]BookRequest, error)
	GetTopRequestedBooks(limit int) ([]BookRequestStats, error)
	DeleteRequest(userID, bookID int) error
	HasRequested(userID, bookID int) (bool, error)
}

type BookRequestStats struct {
	BookID       int    `json:"book_id"`
	Title        string `json:"title"`
	Author       string `json:"author"`
	ImagePath    string `json:"image_path"`
	RequestCount int    `json:"request_count"`
}

type PostgresRequestStore struct {
	db *sql.DB
}

func NewPostgresRequestStore(db *sql.DB) *PostgresRequestStore {
	return &PostgresRequestStore{db: db}
}

func (s *PostgresRequestStore) AddRequest(req BookRequest) error {
	query := `
		INSERT INTO book_requests (book_id, requester_id)
		VALUES ($1, $2)
		ON CONFLICT (book_id, requester_id) DO NOTHING`
	_, err := s.db.Exec(query, req.BookID, req.RequesterID)
	return err
}

func (s *PostgresRequestStore) GetRequestsByUserID(userID int) ([]BookRequest, error) {
	query := `
		SELECT br.id, br.book_id, br.requester_id, br.created_at, b.title, b.author, COALESCE(b.image_path, '')
		FROM book_requests br
		JOIN books b ON br.book_id = b.id
		WHERE br.requester_id = $1
		ORDER BY br.created_at DESC`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := []BookRequest{}
	for rows.Next() {
		var r BookRequest
		if err := rows.Scan(&r.ID, &r.BookID, &r.RequesterID, &r.CreatedAt, &r.BookTitle, &r.BookAuthor, &r.BookImage); err != nil {
			return nil, err
		}
		requests = append(requests, r)
	}
	return requests, nil
}

func (s *PostgresRequestStore) GetTopRequestedBooks(limit int) ([]BookRequestStats, error) {
	query := `
		SELECT b.id, b.title, b.author, COALESCE(b.image_path, ''), COUNT(br.id) as request_count
		FROM books b
		JOIN book_requests br ON b.id = br.book_id
		GROUP BY b.id
		ORDER BY request_count DESC
		LIMIT $1`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := []BookRequestStats{}
	for rows.Next() {
		var s BookRequestStats
		if err := rows.Scan(&s.BookID, &s.Title, &s.Author, &s.ImagePath, &s.RequestCount); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

func (s *PostgresRequestStore) DeleteRequest(userID, bookID int) error {
	query := `DELETE FROM book_requests WHERE requester_id = $1 AND book_id = $2`
	_, err := s.db.Exec(query, userID, bookID)
	return err
}

func (s *PostgresRequestStore) HasRequested(userID, bookID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM book_requests WHERE requester_id = $1 AND book_id = $2)`
	var exists bool
	err := s.db.QueryRow(query, userID, bookID).Scan(&exists)
	return exists, err
}
