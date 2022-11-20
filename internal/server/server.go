package server

import (
	"context"
	"net/http"
	"user-microservice/docs"
	_ "user-microservice/docs"
	usersHttp "user-microservice/internal/users/http"
	usersRepo "user-microservice/internal/users/repository/mongodb"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	CurrentApiVersion = "/api/v1"
	UsersPath         = "/users"
)

// Server - server main struct
type Server struct {
	db   *mongo.Database
	echo *echo.Echo
}

// New - returns a newly initialized server
func New(db *mongo.Database) *Server {
	return NewWithEcho(db, echo.New())
}

// NewWithEcho - same as New but with a given echo.Echo
func NewWithEcho(db *mongo.Database, e *echo.Echo) *Server {
	return &Server{db, e}
}

// Run - Executes the server and starts it
func (s *Server) Run() error {
	router := s.echo.Group(CurrentApiVersion)

	//middlewares
	// router.Use(middleware.RemoveTrailingSlash())
	router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		// AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderContentType},
	}))
	router.Use(middleware.Recover())
	router.Use(middleware.Secure())

	//Health check route
	router.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "It's alive!",
		})
	})

	// Swagger route
	docs.SwaggerInfo.BasePath = CurrentApiVersion
	router.GET("/swagger/*", echoSwagger.WrapHandler)

	// Initialize repositories
	usersR := usersRepo.NewMongoDBRepository(s.db)

	//Initialize http handlers
	usersHandler := usersHttp.NewHttpHandler(usersR)

	// Append routes
	usersHttp.AppendUsersRoutes(router.Group(UsersPath), usersHandler)

	//Start the server
	if err := s.echo.Start(":4040"); err != nil {
		return err
	}

	return nil
}

// Cleanup - performs the needed cleanups for the server.
// Should be sed as a defered function
func (s *Server) Cleanup() error {
	if s.db != nil {
		if err := s.db.Client().Disconnect(context.TODO()); err != nil {
			logrus.Errorf("Error in server.Cleanup -> error disconnecting MongoDB: %s", err)
			return err
		}
	}

	return nil
}
