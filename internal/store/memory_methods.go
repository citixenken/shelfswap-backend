package store

import (
	"errors"
	"sort"
)

func (s *InMemoryBookStore) GetByID(id int) (Book, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, book := range s.books {
		if book.ID == id {
			return book, nil
		}
	}
	return Book{}, errors.New("book not found")
}

func (s *InMemoryBookStore) GetByUserID(userID int) ([]Book, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var userBooks []Book
	for _, book := range s.books {
		if book.UserID == userID {
			userBooks = append(userBooks, book)
		}
	}
	return userBooks, nil
}

func (s *InMemoryBookStore) Update(book Book) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, b := range s.books {
		if b.ID == book.ID {
			s.books[i] = book
			return nil
		}
	}
	return errors.New("book not found")
}

func (s *InMemoryBookStore) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, book := range s.books {
		if book.ID == id {
			s.books = append(s.books[:i], s.books[i+1:]...)
			return nil
		}
	}
	return errors.New("book not found")
}

func (s *InMemoryBookStore) GetGenres() ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	genreMap := make(map[string]bool)
	for _, book := range s.books {
		if book.Genre != "" {
			genreMap[book.Genre] = true
		}
	}

	var genres []string
	for genre := range genreMap {
		genres = append(genres, genre)
	}
	return genres, nil
}

func (s *InMemoryBookStore) GetPopularGenres() ([]GenreStats, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	genreCounts := make(map[string]int)
	for _, book := range s.books {
		if book.Genre != "" {
			genreCounts[book.Genre]++
		}
	}

	var stats []GenreStats
	for genre, count := range genreCounts {
		stats = append(stats, GenreStats{Genre: genre, BookCount: count})
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].BookCount > stats[j].BookCount
	})

	return stats, nil
}
