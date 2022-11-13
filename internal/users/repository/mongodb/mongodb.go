package mongodb

import (
	"context"
	"time"
	"user-microservice/internal/models"
	"user-microservice/internal/users"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
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
	user.ID = uuid.New()
	user.CreatedAt = time.Now().UTC()
	user.UpdatedAt = time.Now().UTC()

	if _, err := r.db.InsertOne(ctx, user); err != nil {
		logrus.Errorf("Error in repository/mongodb.Create -> error: %s", err)
		return nil, err
	}

	return &user, nil
}
