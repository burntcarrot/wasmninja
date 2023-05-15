package module

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
)

type ModuleCache struct {
	redisClient *redis.Client
}

func NewModuleCache(redisClient *redis.Client) *ModuleCache {
	return &ModuleCache{
		redisClient: redisClient,
	}
}

func (c *ModuleCache) GetModule(moduleName string) ([]byte, error) {
	moduleBytes, err := c.redisClient.Get(context.Background(), moduleName).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("module not found in cache")
		}
		return nil, fmt.Errorf("failed to get module from cache: %w", err)
	}

	return moduleBytes, nil
}

func (c *ModuleCache) CacheModule(moduleName string, moduleBytes []byte) error {
	err := c.redisClient.Set(context.Background(), moduleName, moduleBytes, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to cache module: %w", err)
	}

	return nil
}
