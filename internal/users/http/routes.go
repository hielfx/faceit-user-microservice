package http

import (
	"user-microservice/internal/users"

	"github.com/labstack/echo/v4"
)

// AppendUsersRoutes - Sets the users routes for the given echo group
func AppendUsersRoutes(e *echo.Group, h users.Handler) {
	e.POST("", h.CreateUser)
}
