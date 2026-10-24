package link

import (
	"context"
	"errors"
)

type Repository interface {
	Save(ctx context.Context, link Link) error
	Get(ctx context.Context, code string) (Link, error)
}

var ErrNotFound = errors.New("link not found")
var ErrCodeExists = errors.New("code already exists")
