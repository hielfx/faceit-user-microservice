package http

import (
	"user-microservice/internal/users"

	"github.com/labstack/echo/v4"
)

// AppendUsersRoutes - Sets the users routes for the given echo group
func AppendUsersRoutes(e *echo.Group, h users.Handler) {
	e.GET("", h.GetAllUsers)
	e.POST("", h.CreateUser)
	e.GET("/:userId", h.GetUserByID)
	e.POST("/:userId", h.UpdateUserByID)
	e.DELETE("/:userId", h.DeleteUserByID)
}
