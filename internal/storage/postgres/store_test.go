package postgres

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"url-shorter/internal/link"
)

func TestSaveError(t *testing.T) {
	if err := saveError(&pgconn.PgError{Code: "23505"}, "taken"); !errors.Is(err, link.ErrCodeExists) {
		t.Fatalf("unique violation = %v", err)
	}
	other := errors.New("database unavailable")
	if err := saveError(other, "code"); !errors.Is(err, other) {
		t.Fatalf("other error = %v", err)
	}
}

func TestGetError(t *testing.T) {
	if err := getError(pgx.ErrNoRows); !errors.Is(err, link.ErrNotFound) {
		t.Fatalf("no rows = %v", err)
	}
	other := errors.New("database unavailable")
	if err := getError(other); !errors.Is(err, other) {
		t.Fatalf("other error = %v", err)
	}
}
