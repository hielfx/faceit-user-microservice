package mongodb

import (
	"context"
	"errors"
	"time"
	"user-microservice/internal/models"
	"user-microservice/internal/users"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
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

// Create - inserts the user into the database and returns the updated version
func (r mongodbRepository) Create(ctx context.Context, user models.User) (*models.User, error) {
	user.ID = uuid.New()
	user.CreatedAt = time.Now().UTC()
	user.UpdatedAt = time.Now().UTC()

	if _, err := r.db.InsertOne(ctx, &user); err != nil {
		logrus.Errorf("Error in repository/mongodb.Create -> error: %s", err)
		return nil, err
	}

	return &user, nil
}

// GetById - retrieves the user with the given ID
func (r mongodbRepository) GetById(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var res models.User
	if err := r.db.FindOne(ctx, bson.M{"_id": id}).Decode(&res); err != nil {
		if err != mongo.ErrNoDocuments {
			logrus.Errorf("Error in repository/mongodb.GetById -> error: %s", err)
		}
		return nil, err
	}

	return &res, nil
}

// Update - updates the user in the DB and returns the updated version
func (r mongodbRepository) Update(ctx context.Context, user models.User) (*models.User, error) {
	return nil, errors.New("Not implemented")
}

// DeleteById - removes the user with the given ID from the DB
func (r mongodbRepository) DeleteById(ctx context.Context, id uuid.UUID) error {
	return errors.New("Not implemented")
}
