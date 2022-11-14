package main

import (
	"context"
	"net/http"
	usersHttp "user-microservice/internal/users/http"
	usersRepository "user-microservice/internal/users/repository/mongodb"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {

	mongoOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), mongoOptions)
	if err != nil {
		logrus.Fatalln(err)
		return
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			logrus.Fatalln(err)
		}
	}()
	db := client.Database("users-microservice")

	e := echo.New()

	router := e.Group("/api/v1")

	router.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "It's alive!",
		})
	})

	// Users
	usersRepo := usersRepository.NewMongoDBRepository(db)
	usersHandler := usersHttp.NewHttpHandler(usersRepo)
	usersHttp.AppendUsersRoutes(router.Group("/users"), usersHandler)

	if err := e.Start(":4040"); err != nil {
		logrus.Fatalln(err)
		return
	}

}
