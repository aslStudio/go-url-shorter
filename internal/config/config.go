package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	AppPort     string
	RedisAddr   string
	RedisTTL    time.Duration
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}

	ttl, _ := time.ParseDuration(os.Getenv("REDIS_TTL"))

	cfg := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		AppPort:     os.Getenv("APP_PORT"),
		RedisAddr:   os.Getenv("REDIS_ADDR"),
		RedisTTL:    ttl,
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.AppPort == "" {
		return Config{}, errors.New("APP_PORT is required")
	}
	return cfg, nil
}
