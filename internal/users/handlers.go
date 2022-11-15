package users

import "github.com/labstack/echo/v4"

// Handler - user handlers
type Handler interface {
	CreateUser(c echo.Context) error
	GetAllUsers(c echo.Context) error
	GetUserByID(c echo.Context) error
	UpdateUserByID(c echo.Context) error
	DeleteUserByID(c echo.Context) error
}
