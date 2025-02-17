package config

import (
	"os"

	"github.com/redis/go-redis/v9"
)

func NewRedisOptions() *redis.Options {
	return &redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_USER_PASSWORD"),
		Username: os.Getenv("REDIS_USER"),
	}
}
