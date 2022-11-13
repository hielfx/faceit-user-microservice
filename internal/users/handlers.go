package users

import "github.com/labstack/echo/v4"

type Handler interface {
	Create(c echo.Context) error
}
