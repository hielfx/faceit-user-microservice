package pubsub

import (
	"context"
	"user-microservice/internal/models"

	"github.com/go-redis/redis/v8"
)

type redisPubSub struct {
	rc *redis.Client
}

var _ PubSub = redisPubSub{}
var _ PubSub = (*redisPubSub)(nil)

// NewPubSub - returns a new User PubSub
func NewPubSub(rc *redis.Client) *redisPubSub {
	return &redisPubSub{rc}
}

// NotifyUserCreation - publish to the TopicUserCreation topic
func (rps redisPubSub) NotifyUserCreation(ctx context.Context, created models.User) error {
	return rps.rc.Publish(ctx, TopicUserCreation, created).Err()
}

// NotifyUserUpdate - publish to the TopicUserUpdate topic
func (rps redisPubSub) NotifyUserUpdate(ctx context.Context, updatedUser models.User) error {
	return rps.rc.Publish(ctx, TopicUserUpdate, updatedUser).Err()
}

// NotifyUserDeletion - publish to the TopicUserDeletion topic
func (rps redisPubSub) NotifyUserDeletion(ctx context.Context, deletedUserID string) error {
	return rps.rc.Publish(ctx, TopicUserDeletion, deletedUserID).Err()
}
