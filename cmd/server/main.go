package main

import (
	"context"
	"os"
	"user-microservice/config"
	"user-microservice/internal/server"
	"user-microservice/pkg/db/mongodb"
	redisdb "user-microservice/pkg/db/redis"
)

// @title       Users Microservices
// @version     1.0
// @description Users Microservices
// @BasePath    /api/v1
func main() {

	filepath := os.Getenv("CONFIG_FILE")

	cfg, err := config.GetConfigFromFile(filepath)
	if err != nil {
		panic(err)
	}

	db, err := mongodb.NewMongoDatabase(cfg.Mongo)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := db.Client().Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	redisDB := redisdb.MewRedisDatabase(cfg.Redis)
	defer func() {
		if err := redisDB.Close(); err != nil {
			panic(err)
		}
	}()

	s := server.New(db, redisDB, cfg)
	defer func() {
		if err := s.Cleanup(); err != nil {
			panic(err)
		}
	}()

	if err := s.Run(); err != nil {
		panic(err)
	}

}
