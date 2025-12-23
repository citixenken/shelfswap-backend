package store

func (s *PostgresUserStore) Update(user User) error {
	query := `UPDATE users SET username = $1, bio = $2, avatar_path = $3, location = $4 WHERE id = $5`
	_, err := s.db.Exec(query, user.Username, user.Bio, user.AvatarPath, user.Location, user.ID)
	return err
}
