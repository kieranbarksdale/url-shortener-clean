package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL string
	RedisURL    string
	Port        string
}

func Load() (*Config, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	redisURL := os.Getenv("REDIS_URL")
	port := os.Getenv("PORT")

	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}
	if redisURL == "" {
		return nil, fmt.Errorf("REDIS_URL environment variable is required")
	}
	if port == "" {
		return nil, fmt.Errorf("PORT environment variable is required")
	}
	port = ":" + port

	return &Config{
		DatabaseURL: databaseURL,
		RedisURL:    redisURL,
		Port:        port,
	}, nil
}
