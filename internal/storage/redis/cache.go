package redisstore

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewCache(
	client *redis.Client,
	ttl time.Duration,
) *Cache {
	return &Cache{
		client: client,
		ttl:    ttl,
	}
}

const linkPrefix = "link:"

func (c *Cache) Get(
	ctx context.Context,
	code string,
) (string, bool, error) {
	result, err := c.client.Get(ctx, linkPrefix+code).Result()

	if errors.Is(err, redis.Nil) {
		return "", false, nil
	}

	if err != nil {
		return "", false, err
	}

	return result, true, nil
}

func (c *Cache) Set(
	ctx context.Context,
	code string,
	originalURL string,
) error {
	err := c.client.Set(ctx, linkPrefix+code, originalURL, c.ttl).Err()

	return err
}
