package users

import (
	"context"
	"user-microservice/internal/models"
)

type Repository interface {
	Create(ctx context.Context, user models.User) (*models.User, error)
}
