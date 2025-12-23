package store

func (s *PostgresBookStore) GetGenres() ([]string, error) {
	query := `SELECT DISTINCT genre FROM books WHERE genre IS NOT NULL AND genre != '' ORDER BY genre ASC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []string
	for rows.Next() {
		var g string
		if err := rows.Scan(&g); err != nil {
			return nil, err
		}
		genres = append(genres, g)
	}
	return genres, nil
}

type GenreStats struct {
	Genre     string `json:"genre"`
	BookCount int    `json:"book_count"`
}

func (s *PostgresBookStore) GetPopularGenres() ([]GenreStats, error) {
	query := `
		SELECT genre, COUNT(*) as book_count
		FROM books
		WHERE genre IS NOT NULL AND genre != ''
		GROUP BY genre
		ORDER BY book_count DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	defer rows.Close()

	stats := []GenreStats{}
	for rows.Next() {
		var s GenreStats
		if err := rows.Scan(&s.Genre, &s.BookCount); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}
