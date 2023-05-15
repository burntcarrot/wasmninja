package cache

import (
	"github.com/burntcarrot/wasmninja/internal/config"

	"github.com/go-redis/redis/v8"
)

type Cache struct {
	Connection *redis.Client
}

func NewConnection(cfg config.CacheConfig) *Cache {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &Cache{
		Connection: redisClient,
	}
}
