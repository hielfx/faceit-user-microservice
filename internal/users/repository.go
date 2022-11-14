package users

import (
	"context"
	"user-microservice/internal/models"

	"github.com/google/uuid"
)

// Repository - users repository
type Repository interface {
	Create(ctx context.Context, user models.User) (*models.User, error)
	GetById(ctx context.Context, id uuid.UUID) (*models.User, error)
	Update(ctx context.Context, user models.User) (*models.User, error)
	DeleteById(ctx context.Context, id uuid.UUID) error
}
