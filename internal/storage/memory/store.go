package memory

import (
	"context"
	"fmt"
	"sync"

	"url-shorter/internal/link"
)

type Store struct {
	mu    sync.RWMutex
	links map[string]link.Link
}

func NewStore() *Store {
	return &Store{links: make(map[string]link.Link)}
}

func (s *Store) Save(_ context.Context, item link.Link) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.links[item.Code]; exists {
		return fmt.Errorf("%w: %s", link.ErrCodeExists, item.Code)
	}
	s.links[item.Code] = item
	return nil
}

func (s *Store) Get(_ context.Context, code string) (link.Link, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.links[code]
	if !exists {
		return link.Link{}, link.ErrNotFound
	}
	return item, nil
}

var _ link.Repository = (*Store)(nil)
