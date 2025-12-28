package store

func (s *PostgresUserStore) Update(user User) error {
	query := `UPDATE users SET username = $1, bio = $2, avatar_path = $3, location = $4, clerk_id = $5 WHERE id = $6`
	_, err := s.db.Exec(query, user.Username, user.Bio, user.AvatarPath, user.Location, user.ClerkID, user.ID)
	return err
}
