package store

import "database/sql"

func (s *PostgresBookStore) GetByID(id int) (Book, error) {
	query := `
		SELECT b.id, b.title, b.author, COALESCE(b.description, ''), COALESCE(b.image_path, ''), b.created_at, b.user_id, COALESCE(u.email, ''), COALESCE(u.username, ''), COALESCE(u.avatar_path, '')
		FROM books b
		LEFT JOIN users u ON b.user_id = u.id
		WHERE b.id = $1`
	var book Book
	var userID sql.NullInt64
	err := s.db.QueryRow(query, id).Scan(&book.ID, &book.Title, &book.Author, &book.Description, &book.ImagePath, &book.CreatedAt, &userID, &book.UserEmail, &book.UserUsername, &book.UserAvatarPath)
	if err != nil {
		return Book{}, err
	}
	if userID.Valid {
		book.UserID = int(userID.Int64)
	}
	return book, nil
}

func (s *PostgresBookStore) GetByUserID(userID int) ([]Book, error) {
	query := `
		SELECT id, title, author, COALESCE(description, ''), COALESCE(image_path, ''), created_at, user_id
		FROM books
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	books := []Book{}
	for rows.Next() {
		var b Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Description, &b.ImagePath, &b.CreatedAt, &b.UserID); err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	return books, nil
}

func (s *PostgresBookStore) Update(book Book) error {
	query := `UPDATE books SET title = $1, author = $2, description = $3, genre = $4, image_path = $5 WHERE id = $6`
	_, err := s.db.Exec(query, book.Title, book.Author, book.Description, book.Genre, book.ImagePath, book.ID)
	return err
}

func (s *PostgresBookStore) Delete(id int) error {
	query := `DELETE FROM books WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}
