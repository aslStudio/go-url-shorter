package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"url-shorter/internal/link"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Save(ctx context.Context, item link.Link) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO links (code, original_url) VALUES ($1, $2)`, item.Code, item.OriginalURL)
	return saveError(err, item.Code)
}

func saveError(err error, code string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return fmt.Errorf("%w: %s", link.ErrCodeExists, code)
	}
	return err
}

func (s *Store) Get(ctx context.Context, code string) (link.Link, error) {
	var item link.Link
	err := s.pool.QueryRow(ctx, `SELECT code, original_url FROM links WHERE code = $1`, code).Scan(&item.Code, &item.OriginalURL)
	if err := getError(err); err != nil {
		return link.Link{}, err
	}
	return item, nil
}

func getError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return link.ErrNotFound
	}
	return err
}

var _ link.Repository = (*Store)(nil)
