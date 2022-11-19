package http

import (
	"context"
	"fmt"
	"net/http"
	httpErrors "user-microservice/internal/errors/http"
	"user-microservice/internal/models"
	"user-microservice/internal/users"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

type httpHandler struct {
	repository users.Repository
}

var _ users.Handler = httpHandler{}
var _ users.Handler = (*httpHandler)(nil)

// NewHttpHandler - returns a new user http handler initialized with the repository
func NewHttpHandler(usersRepository users.Repository) users.Handler {
	return &httpHandler{usersRepository}
}

func (h httpHandler) CreateUser(c echo.Context) error {
	var body models.User

	if err := c.Bind(&body); err != nil {
		logrus.Errorf("Error in users/http.CreateUser -> error binding body: %s", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if !body.Valid() {
		return echo.NewHTTPError(http.StatusBadRequest, httpErrors.ErrInvalidBody)
	}
	// pwd, err := sec.HashPassword(body.Password)
	// if err != nil {
	// 	return err
	// }
	// body.User.Password = pwd

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
	userIDstr := c.Param("userId")
	userID, err := uuid.Parse(userIDstr)
	if err != nil {
		logrus.Errorf("Error in users/http.GetUserByID -> error parsing user ID: %s", err)
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid user ID %s", userIDstr))
	}

	user, err := h.repository.GetById(context.TODO(), userID)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("User not found for ID %s", userID))
		}
		return err
	}

	return c.JSON(http.StatusOK, user)
}

func (h httpHandler) UpdateUserByID(c echo.Context) error {
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid user ID %s", userIDStr))
	}

	var body models.User
	if err := c.Bind(&body); err != nil {
		logrus.Errorf("Error in users/http.UpdateUserByID -> error binding body: %s", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	userToModify, err := h.repository.GetById(context.TODO(), userID)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("User not found for ID %s", userID))
		}
		return err
	}

	userToModify.Modify(body)
	if !userToModify.Valid() {
		return echo.NewHTTPError(http.StatusBadRequest, httpErrors.ErrInvalidBody)
	}

	res, err := h.repository.Update(context.TODO(), *userToModify)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (h httpHandler) DeleteUserByID(c echo.Context) error {
	idStr := c.Param("userId")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid ID %s", idStr))
	}

	if err := h.repository.DeleteById(context.TODO(), userID); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
