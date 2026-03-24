package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func GetURL(redisClient *redis.Client, shortCode string) (string, error) {
	originalURL, err := redisClient.Get(context.Background(), shortCode).Result()
	if err != nil {
		return "", err
	}

	return originalURL, nil
}
