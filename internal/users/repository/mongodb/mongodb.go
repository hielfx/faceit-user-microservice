package mongodb

import (
	"context"
	"errors"
	"math"
	"time"
	"user-microservice/internal/models"
	"user-microservice/internal/pagination"
	"user-microservice/internal/users"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	if err := r.db.FindOne(ctx, bson.M{"id": id}).Decode(&res); err != nil {
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
	if _, err := r.db.DeleteOne(ctx, bson.M{"id": id}); err != nil {
		logrus.Errorf("Error in repository/mongodb.DeleteById -> error: %s", err)
		return err
	}

	return nil
}

// GetPaginatedUsers - returns a list of paginated user
func (r mongodbRepository) GetPaginatedUsers(ctx context.Context, pag pagination.PaginationOptions) (models.PaginatedUsers, error) {
	pageSize := pag.Size
	if pageSize <= 0 {
		pageSize = pagination.DefaultSize
	}
	currentPage := pag.Page
	if currentPage <= 0 {
		currentPage = pagination.FirstPage
	}

	res := models.PaginatedUsers{
		Paginated: pagination.Paginated{
			Size:        pageSize,
			CurrentPage: currentPage,
		},
	}

	//count how many users are stored
	totalCount, err := r.db.CountDocuments(ctx, bson.M{})
	if err != nil {
		logrus.Errorf("Error in repository/mongodb.GetPaginatedUsers -> error executing count command: %s", err)
		return res, nil
	}

	//Define findOptions
	findOptions := options.Find()
	findOptions.SetLimit(int64(pageSize))
	skipValue := (int64(currentPage) - 1) * int64(pageSize)
	if skipValue < totalCount {
		findOptions.SetSkip(skipValue)
	}
	// if pag.OrderBy != "" {
	// 	sortOrder := 1
	// 	if pag.SortOrder.IsDesc() {
	// 		sortOrder = -1
	// 	}
	// 	findOptions.SetSort(bson.D{{Key: pag.OrderBy, Value: sortOrder}})
	// }

	// retrieve the users
	cursor, err := r.db.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		logrus.Errorf("Error in repository/mongodb.GetPaginatedUsers -> error executing find command: %s", err)
		return res, nil
	}
	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		logrus.Errorf("Error in repository/mongodb.GetPaginatedUsers -> error decoding cursor: %s", err)
		return res, err
	}

	//TODO: move pagination logic to its package (set total count, set total pages, set has more, etc.)
	res.Users = users
	res.TotalCount = totalCount
	res.TotalPages = int64(math.Ceil(float64(totalCount) / float64(pageSize)))
	res.HasMore = int64(currentPage) < totalCount/int64(pageSize)

	return res, nil
}
