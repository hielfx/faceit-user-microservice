package redis

import (
	"user-microservice/internal/users"

	"github.com/go-redis/redis/v8"
)

type redisRepository struct {
	rdb *redis.Client
}

var _ users.PubSubRepository = redisRepository{}
var _ users.PubSubRepository = (*redisRepository)(nil)

// NewRedisRepository - returns a new redis repository instance
func NewRedisRepository(rdb *redis.Client) redisRepository {
	return redisRepository{rdb}
}
