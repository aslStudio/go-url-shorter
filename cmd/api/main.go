package main

import (
	"context"
	"log"
	"net/http"
	"time"
	redisstore "url-shorter/internal/storage/redis"

	"url-shorter/internal/config"
	"url-shorter/internal/httpapi"
	"url-shorter/internal/link"
	"url-shorter/internal/storage/postgres"

	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Redis unavailable: %v", err)
	}

	service := link.NewService(
		postgres.NewStore(pool),
		redisstore.NewCache(redisClient, cfg.RedisTTL),
	)
	handler := httpapi.NewHandler(service, "http://localhost:"+cfg.AppPort)
	if err := http.ListenAndServe(":"+cfg.AppPort, httpapi.NewRouter(handler)); err != nil {
		log.Fatal(err)
	}
}
