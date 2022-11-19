package main

import (
	"context"
	"user-microservice/internal/server"
	"user-microservice/pkg/db/mongodb"
)

// @title       Faceit Users Microservices
// @version     1.0
// @description Faceit Users Microservices
// @BasePath    /api/v1
func main() {

	db, err := mongodb.NewMongoDatabase()
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := db.Client().Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	s := server.New(db)
	defer func() {
		if err := s.Cleanup(); err != nil {
			panic(err)
		}
	}()

	if err := s.Run(); err != nil {
		panic(err)
	}

}
