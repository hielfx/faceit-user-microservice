package redis

import (
	"user-microservice/config"

	"github.com/go-redis/redis/v8"
)

// MewRedisDatabase - creates a new redis client
func MewRedisDatabase(cfg config.RedisConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return client
}
