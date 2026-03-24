package cache

import (
	"context"
	"time"
	"github.com/redis/go-redis/v9"
)

func SetURL(redisClient *redis.Client, shortCode string, originalURL string) error {
	err := redisClient.Set(context.Background(), shortCode, originalURL, 24*time.Hour).Err()
	if err != nil {
		return err
	}

	return nil
}
