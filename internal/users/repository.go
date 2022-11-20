//go:generate mockgen -source repository.go -destination mock/repository_mock.go -package mock
package users

import (
	"context"
	"user-microservice/internal/models"
	"user-microservice/internal/pagination"
)

// Repository - users repository
type Repository interface {
	Create(ctx context.Context, user models.User) (*models.User, error)
	GetById(ctx context.Context, id string) (*models.User, error)
	Update(ctx context.Context, user models.User) (*models.User, error)
	DeleteById(ctx context.Context, id string) error
	GetPaginatedUsers(ctx context.Context, pagination pagination.PaginationOptions, filters models.UserFilters) (models.PaginatedUsers, error)
}
