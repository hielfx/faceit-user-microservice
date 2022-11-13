package users

import "github.com/labstack/echo/v4"

// Handler - user handlers
type Handler interface {
	CreateUser(c echo.Context) error
}
