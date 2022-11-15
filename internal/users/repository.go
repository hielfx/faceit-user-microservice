package users

import (
	"context"
	"user-microservice/internal/models"
	"user-microservice/internal/pagination"

	"github.com/google/uuid"
)

// Repository - users repository
type Repository interface {
	Create(ctx context.Context, user models.User) (*models.User, error)
	GetById(ctx context.Context, id uuid.UUID) (*models.User, error)
	Update(ctx context.Context, user models.User) (*models.User, error)
	DeleteById(ctx context.Context, id uuid.UUID) error
	GetPaginatedUsers(ctx context.Context, pagination pagination.PaginationOptions) (models.PaginatedUsers, error)
}
