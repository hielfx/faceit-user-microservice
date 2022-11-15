package http

import (
	"context"
	"net/http"
	"user-microservice/internal/models"
	"user-microservice/internal/users"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type httpHandler struct {
	repository users.Repository
}

var _ users.Handler = httpHandler{}
var _ users.Handler = (*httpHandler)(nil)

func NewHttpHandler(usersRepository users.Repository) users.Handler {
	return &httpHandler{usersRepository}
}

func (h httpHandler) CreateUser(c echo.Context) error {
	var body models.User

	if err := c.Bind(&body); err != nil {
		logrus.Errorf("Error in users/http.CreateUser -> error binding body: %s", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	//TODO: Validate body

	ctx := context.TODO()
	res, err := h.repository.Create(ctx, body)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, res)
}

func (h httpHandler) GetAllUsers(c echo.Context) error {
	//TODO: Implement method
	return echo.NewHTTPError(http.StatusNotImplemented, http.StatusText(http.StatusNotImplemented))
}

func (h httpHandler) GetUserByID(c echo.Context) error {
	//TODO: Implement method
	return echo.NewHTTPError(http.StatusNotImplemented, http.StatusText(http.StatusNotImplemented))
}

func (h httpHandler) UpdateUserByID(c echo.Context) error {
	//TODO: Implement method
	return echo.NewHTTPError(http.StatusNotImplemented, http.StatusText(http.StatusNotImplemented))
}

func (h httpHandler) DeleteUserByID(c echo.Context) error {
	//TODO: Implement method
	return echo.NewHTTPError(http.StatusNotImplemented, http.StatusText(http.StatusNotImplemented))
}
