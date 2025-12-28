package store

import (
	"database/sql"
	"time"
)

type User struct {
	ID         int       `json:"id"`
	ClerkID    string    `json:"clerk_id,omitempty"`
	Email      string    `json:"email"`
	Password   string    `json:"-"` // Don't return password in JSON
	Username   string    `json:"username,omitempty"`
	Bio        string    `json:"bio,omitempty"`
	AvatarPath string    `json:"avatar_path,omitempty"`
	Location   string    `json:"location,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type UserStore interface {
	Create(user User) error
	GetByEmail(email string) (User, error)
	GetByID(id int) (User, error)
	Update(user User) error
	SaveResetToken(token string, userID int, expiry time.Time) error
	GetResetToken(token string) (int, time.Time, error)
	DeleteResetToken(token string) error
	UpdatePassword(userID int, password string) error
	GetMembers() ([]User, error)
}

type PostgresUserStore struct {
	db *sql.DB
}

func NewPostgresUserStore(db *sql.DB) *PostgresUserStore {
	return &PostgresUserStore{
		db: db,
	}
}

func (s *PostgresUserStore) Migrate() error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);
		ALTER TABLE users ADD COLUMN IF NOT EXISTS username TEXT;
		ALTER TABLE users ADD COLUMN IF NOT EXISTS bio TEXT;
		ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_path TEXT;
		ALTER TABLE users ADD COLUMN IF NOT EXISTS location TEXT;
		ALTER TABLE users ADD COLUMN IF NOT EXISTS clerk_id TEXT UNIQUE;

		CREATE TABLE IF NOT EXISTS password_resets (
			token TEXT PRIMARY KEY,
			user_id INTEGER REFERENCES users(id),
			expiry TIMESTAMP NOT NULL
		);

		CREATE TABLE IF NOT EXISTS book_requests (
			id SERIAL PRIMARY KEY,
			book_id INTEGER REFERENCES books(id),
			requester_id INTEGER REFERENCES users(id),
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(book_id, requester_id)
		);
		`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresUserStore) Create(user User) error {
	query := `
		INSERT INTO users (email, password, username, avatar_path, clerk_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	return s.db.QueryRow(query, user.Email, user.Password, user.Username, user.AvatarPath, user.ClerkID).Scan(&user.ID, &user.CreatedAt)
}

func (s *PostgresUserStore) GetByEmail(email string) (User, error) {
	query := `SELECT id, email, password, COALESCE(username, ''), COALESCE(bio, ''), COALESCE(avatar_path, ''), COALESCE(location, ''), created_at, COALESCE(clerk_id, '') FROM users WHERE email = $1`
	var user User
	err := s.db.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.Password, &user.Username, &user.Bio, &user.AvatarPath, &user.Location, &user.CreatedAt, &user.ClerkID)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (s *PostgresUserStore) GetByID(id int) (User, error) {
	query := `SELECT id, email, password, COALESCE(username, ''), COALESCE(bio, ''), COALESCE(avatar_path, ''), COALESCE(location, ''), created_at, COALESCE(clerk_id, '') FROM users WHERE id = $1`
	var user User
	err := s.db.QueryRow(query, id).Scan(&user.ID, &user.Email, &user.Password, &user.Username, &user.Bio, &user.AvatarPath, &user.Location, &user.CreatedAt, &user.ClerkID)
	if err != nil {
		return User{}, err
	}
	return user, nil
}
