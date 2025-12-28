package store

func (s *PostgresUserStore) GetByClerkID(clerkID string) (User, error) {
	query := `SELECT id, email, password, COALESCE(username, ''), COALESCE(bio, ''), COALESCE(avatar_path, ''), COALESCE(location, ''), created_at, COALESCE(clerk_id, '') FROM users WHERE clerk_id = $1`
	var user User
	err := s.db.QueryRow(query, clerkID).Scan(&user.ID, &user.Email, &user.Password, &user.Username, &user.Bio, &user.AvatarPath, &user.Location, &user.CreatedAt, &user.ClerkID)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (s *PostgresUserStore) DeleteByClerkID(clerkID string) error {
	// First get the user ID to delete related books (if cascade isn't set up)
	user, err := s.GetByClerkID(clerkID)
	if err != nil {
		return err
	}

	// Manual cascade for books (just in case)
	_, err = s.db.Exec(`DELETE FROM books WHERE user_id = $1`, user.ID)
	if err != nil {
		return err
	}

	// Manual cascade for book_requests
	_, err = s.db.Exec(`DELETE FROM book_requests WHERE requester_id = $1`, user.ID)
	if err != nil {
		return err
	}

	query := `DELETE FROM users WHERE clerk_id = $1`
	_, err = s.db.Exec(query, clerkID)
	return err
}
