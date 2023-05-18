package cache

import (
	"context"
	"log"
	"time"

	"github.com/burntcarrot/wasmninja/internal/config"

	"github.com/go-redis/redis/v8"
)

type Cache struct {
	Connection *redis.Client
}

func NewConnection(cfg config.CacheConfig) (*Cache, error) {
	redisClient, err := connectRedisWithRetry(cfg)
	if err != nil {
		return nil, err
	}

	return &Cache{
		Connection: redisClient,
	}, nil
}

func connectRedisWithRetry(cfg config.CacheConfig) (*redis.Client, error) {
	var (
		client *redis.Client
		err    error
	)

	// Retry parameters
	retryAttempts := 5
	retryInterval := time.Second

	for i := 0; i < retryAttempts; i++ {
		client = redis.NewClient(&redis.Options{
			Addr:     cfg.Address,
			Password: cfg.Password,
			DB:       cfg.DB,
		})

		ctx := context.Background()
		_, err = client.Ping(ctx).Result()
		if err == nil {
			// Connection successful, break out of the retry loop
			break
		}

		log.Printf("Failed to connect to Redis. Retrying in %s...", retryInterval)

		time.Sleep(retryInterval)
		retryInterval *= 2 // Exponential backoff
	}

	if err != nil {
		return nil, err
	}

	return client, nil
}
