package store

func (s *PostgresUserStore) GetMembers() ([]User, error) {
	query := `
		SELECT id, email, COALESCE(username, ''), COALESCE(bio, ''), COALESCE(avatar_path, ''), COALESCE(location, ''), created_at
		FROM users
		ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.Bio, &u.AvatarPath, &u.Location, &u.CreatedAt); err != nil {
			return nil, err
		}
		members = append(members, u)
	}
	return members, nil
}
