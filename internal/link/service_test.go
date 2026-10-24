package link

import (
	"context"
	"errors"
	"testing"
)

type testRepository struct {
	links map[string]Link
	err   error
}

func (r *testRepository) Save(_ context.Context, item Link) error {
	if r.err != nil {
		return r.err
	}
	if _, exists := r.links[item.Code]; exists {
		return ErrCodeExists
	}
	r.links[item.Code] = item
	return nil
}

func (r *testRepository) Get(_ context.Context, code string) (Link, error) {
	item, exists := r.links[code]
	if !exists {
		return Link{}, ErrNotFound
	}
	return item, nil
}

func TestCreateValidatesURL(t *testing.T) {
	tests := []struct {
		url  string
		want error
	}{
		{"test", ErrInvalidURL},
		{"ftp://example.com", ErrUnsupportedScheme},
		{"https:///path", ErrMissingHostname},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			svc := NewService(&testRepository{links: make(map[string]Link)})
			_, err := svc.Create(context.Background(), tt.url)
			if !errors.Is(err, tt.want) {
				t.Fatalf("Create(%q) error = %v, want %v", tt.url, err, tt.want)
			}
		})
	}
}

func TestCreateRetriesCodeCollision(t *testing.T) {
	repo := &testRepository{links: map[string]Link{"taken": {Code: "taken"}}}
	svc := NewService(repo)
	codes := []string{"taken", "fresh"}
	svc.generateCode = func() string {
		code := codes[0]
		codes = codes[1:]
		return code
	}

	created, err := svc.Create(context.Background(), "https://example.com")
	if err != nil || created.Code != "fresh" || len(codes) != 0 {
		t.Fatalf("Create() = %+v, %v; remaining codes = %v", created, err, codes)
	}
}

func TestCreateStopsAfterFiveCollisions(t *testing.T) {
	repo := &testRepository{links: map[string]Link{"taken": {Code: "taken"}}}
	svc := NewService(repo)
	attempts := 0
	svc.generateCode = func() string {
		attempts++
		return "taken"
	}

	_, err := svc.Create(context.Background(), "https://example.com")
	if !errors.Is(err, ErrCodeAttemptsExhausted) || attempts != 5 {
		t.Fatalf("Create() error = %v, attempts = %d", err, attempts)
	}
}

func TestCreateReturnsStorageError(t *testing.T) {
	want := errors.New("database unavailable")
	svc := NewService(&testRepository{err: want})
	_, err := svc.Create(context.Background(), "https://example.com")
	if !errors.Is(err, want) {
		t.Fatalf("Create() error = %v, want %v", err, want)
	}
}
