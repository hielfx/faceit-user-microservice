package mongodb

import (
	"context"
	"user-microservice/config"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// NewMongoDatabase - creates a new mongodb connection and returns the database
func NewMongoDatabase(cfg config.MongoConfig) (*mongo.Database, error) {
	mongoOptions := options.Client().ApplyURI(cfg.URI).SetRegistry(mongoRegistry)
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

	db := client.Database(cfg.DB)

	return db, nil
}
