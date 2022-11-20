package pubsub

import (
	"context"
	"user-microservice/internal/models"
)

type PubSub interface {
	NotifyUserCreation(ctx context.Context, created models.User) error
	NotifyUserUpdate(ctx context.Context, updatedUser models.User) error
	NotifyUserDeletion(ctx context.Context, deletedUserID string) error
}
