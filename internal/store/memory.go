package store

import (
	"sort"
	"strings"
	"sync"
)

type InMemoryBookStore struct {
	mu     sync.Mutex
	books  []Book
	nextID int
}

func NewInMemoryBookStore() *InMemoryBookStore {
	return &InMemoryBookStore{
		books:  []Book{},
		nextID: 1,
	}
}

func (s *InMemoryBookStore) Add(book Book) (Book, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	book.ID = s.nextID
	s.nextID++
	s.books = append(s.books, book)
	return book, nil
}

func (s *InMemoryBookStore) GetAll(filter BookFilter) ([]Book, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var filtered []Book
	for _, b := range s.books {
		if filter.Query != "" {
			q := strings.ToLower(filter.Query)
			if !strings.Contains(strings.ToLower(b.Title), q) && !strings.Contains(strings.ToLower(b.Author), q) {
				continue
			}
		}
		filtered = append(filtered, b)
	}

	sort.Slice(filtered, func(i, j int) bool {
		if filter.Sort == "oldest" {
			return filtered[i].CreatedAt.Before(filtered[j].CreatedAt)
		}
		return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
	})

	return filtered, nil
}
