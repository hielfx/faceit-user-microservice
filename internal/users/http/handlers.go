package http

import (
	"net/http"
	"user-microservice/internal/users"

	"github.com/labstack/echo/v4"
)

type httpHandler struct {
	repository users.Repository
}

var _ users.Handler = httpHandler{}
var _ users.Handler = (*httpHandler)(nil)

func NewHttpHandler(usersRepository users.Repository) users.Handler {
	return &httpHandler{usersRepository}
}

// TODO: Add Create comments
func (h httpHandler) Create(c echo.Context) error {

	return echo.NewHTTPError(http.StatusNotImplemented, http.StatusText(http.StatusNotImplemented))
}
