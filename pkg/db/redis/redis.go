package redis

import "github.com/go-redis/redis/v8"

// MewRedisDatabase - creates a new redis client
func MewRedisDatabase() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return client
}
