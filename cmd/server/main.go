package main

import (
	"context"
	"net/http"
	usersHttp "user-microservice/internal/users/http"
	usersRepository "user-microservice/internal/users/repository"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
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

	v1 := e.Group("/api/v1")

	v1.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "It's alive!",
		})
	})

	usersRoute := v1.Group("/users")

	usersRepo := usersRepository.NewMongoDBRepository(db)
	usersHandler := usersHttp.NewHttpHandler(usersRepo)
	usersRoute.POST("", usersHandler.Create)

	if err := e.Start(":4040"); err != nil {
		logrus.Fatalln(err)
		return
	}

}
