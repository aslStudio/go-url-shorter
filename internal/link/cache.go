package link

import "context"

type Cache interface {
	Get(ctx context.Context, code string) (string, bool, error)
	Set(ctx context.Context, code, originalURL string) error
}
