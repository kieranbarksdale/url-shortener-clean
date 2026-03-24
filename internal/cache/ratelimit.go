package cache

import (
	"context"
	"time"
	"github.com/redis/go-redis/v9"
)



func IsRateLimited(redisClient *redis.Client, ip string) (bool, error) {
	key := "rate_limit:" + ip
	counter, err := redisClient.Incr(context.Background(), key).Result()
	if err != nil {
		return false, err
	}
	if counter == 1 {
		err := redisClient.Expire(context.Background(), key, 60*time.Second).Err()
		if err != nil {
			return false, err
		}
	}
	if counter >= 10 {
		return true, nil
	}
	return false, nil
}
