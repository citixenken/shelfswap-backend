package store

func (s *PostgresUserStore) GetMembers(searchQuery string) ([]User, error) {
	query := `
		SELECT id, email, COALESCE(username, ''), COALESCE(bio, ''), COALESCE(avatar_path, ''), COALESCE(location, ''), created_at
		FROM users`

	var args []interface{}

	// Add WHERE clause if search query is provided
	if searchQuery != "" {
		query += ` WHERE username ILIKE $1`
		args = append(args, "%"+searchQuery+"%")
	}

	query += ` ORDER BY created_at DESC`

	rows, err := s.db.Query(query, args...)
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
