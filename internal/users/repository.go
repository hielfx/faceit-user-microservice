package users

import (
	"context"
	"user-microservice/internal/models"
)

// Repository - users repository
type Repository interface {
	Create(ctx context.Context, user models.User) (*models.User, error)
}
