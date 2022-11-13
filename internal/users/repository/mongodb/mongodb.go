package mongodb

import (
	"context"
	"errors"
	"user-microservice/internal/models"
	"user-microservice/internal/users"

	"go.mongodb.org/mongo-driver/mongo"
)

const mongodbCollection = "users"

type mongodbRepository struct {
	db *mongo.Collection
}

// NewMongoDBRepository - returns a new instance for the mongodb repository
func NewMongoDBRepository(db *mongo.Database) users.Repository {
	return &mongodbRepository{db.Collection(mongodbCollection)}
}

// Create - inserts the user into the database and returns it
func (r mongodbRepository) Create(ctx context.Context, user models.User) (*models.User, error) {

	return nil, errors.New("Error not implemented")
}
