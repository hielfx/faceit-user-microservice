package mongodb

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// NewMongoDatabase - creates a new mongodb connection and returns the database
// FIXME: Add configuration as param instead of using hardcoded strings
func NewMongoDatabase() (*mongo.Database, error) {
	mongoOptions := options.Client().ApplyURI("mongodb://localhost:27017").SetRegistry(mongoRegistry)
	client, err := mongo.Connect(context.TODO(), mongoOptions)
	if err != nil {
		logrus.Errorf("Error in db/mongodb.NewMongoDatabase -> error connecting to db: %s", err)
		return nil, err
	}
	// defer func() {
	// 	if err := client.Disconnect(context.TODO()); err != nil {
	// 		logrus.Fatalln(err)
	// 	}
	// }()

	//make ping and check if ok
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		logrus.Errorf("Error in db/mongodb.NewMongoDatabase -> error pinging db: %s", err)
		return nil, err
	}

	db := client.Database("users-microservice")

	return db, nil
}
