package store

import (
	"time"
)

func (s *PostgresUserStore) SaveResetToken(token string, userID int, expiry time.Time) error {
	query := `INSERT INTO password_resets (token, user_id, expiry) VALUES ($1, $2, $3)`
	_, err := s.db.Exec(query, token, userID, expiry)
	return err
}

func (s *PostgresUserStore) GetResetToken(token string) (int, time.Time, error) {
	query := `SELECT user_id, expiry FROM password_resets WHERE token = $1`
	var userID int
	var expiry time.Time
	err := s.db.QueryRow(query, token).Scan(&userID, &expiry)
	if err != nil {
		return 0, time.Time{}, err
	}
	return userID, expiry, nil
}

func (s *PostgresUserStore) DeleteResetToken(token string) error {
	query := `DELETE FROM password_resets WHERE token = $1`
	_, err := s.db.Exec(query, token)
	return err
}

func (s *PostgresUserStore) UpdatePassword(userID int, password string) error {
	query := `UPDATE users SET password = $1 WHERE id = $2`
	_, err := s.db.Exec(query, password, userID)
	return err
}
