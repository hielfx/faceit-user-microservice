package http

import (
	"context"
	"fmt"
	"net/http"
	httpErrors "user-microservice/internal/errors/http"
	"user-microservice/internal/models"
	"user-microservice/internal/pagination"
	"user-microservice/internal/users"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

var _ = echo.HTTPError{}

type httpHandler struct {
	repository users.Repository
}

var _ users.Handler = httpHandler{}
var _ users.Handler = (*httpHandler)(nil)

// NewHttpHandler - returns a new user http handler initialized with the repository
func NewHttpHandler(usersRepository users.Repository) users.Handler {
	return &httpHandler{usersRepository}
}

// CreateUser godoc
//
// @Summary     Create a user
// @Description Creates a new user and inserts it in the DB
// @Accept      json
// @Produce     json
// @Param       body body     models.User true "User to create"
// @Success     201  {object} models.User
// @Failure     400  {object} echo.HTTPError
// @Failure     404  {object} echo.HTTPError
// @Failure     500  {object} echo.HTTPError
// @Router      /users [post]
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

// GetAllUsers godoc
//
// @Summary     Gets paginated users
// @Description Gets a paginated users list from the db and returns it
// @Produce     json
// @Param       page    query    int    false "Page to retrieve"
// @Param       size    query    int    false "Page size"
// @Param       country query    string false "Country filter"
// @Success     200     {object} models.PaginatedUsers
// @Failure     400     {object} echo.HTTPError
// @Failure     500     {object} echo.HTTPError
// @Router      /users [get]
func (h httpHandler) GetAllUsers(c echo.Context) error {

	type params struct {
		pagination.PaginationOptions
		models.UserFilters
	}

	var pagOpts params
	if err := c.Bind(&pagOpts); err != nil {
		logrus.Errorf("Error in users/http.GetAllUsers -> error binding params: %s", err)
		return echo.NewHTTPError(http.StatusBadRequest, httpErrors.ErrInvalidParams)
	}

	res, err := h.repository.GetPaginatedUsers(context.TODO(), pagOpts.PaginationOptions, pagOpts.UserFilters)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// GetUserByID godoc
//
// @Summary     Gets a user
// @Description Gets a user by its id from the DB and returns it
// @Produce     json
// @Param       userId path     int true "User id"
// @Success     200    {object} models.User
// @Failure     400    {object} echo.HTTPError
// @Failure     404    {object} echo.HTTPError
// @Failure     500    {object} echo.HTTPError
// @Router      /users/{userId} [get]
func (h httpHandler) GetUserByID(c echo.Context) error {
	userIDstr := c.Param("userId")
	userID, err := uuid.Parse(userIDstr)
	if err != nil {
		logrus.Errorf("Error in users/http.GetUserByID -> error parsing user ID: %s", err)
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid user ID %s", userIDstr))
	}

	user, err := h.repository.GetById(context.TODO(), userID.String())
	if err != nil {
		if err == mongo.ErrNilDocument {
			return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("User not found for ID %s", userID))
		}
		return err
	}

	return c.JSON(http.StatusOK, user)
}

// GetUserByID godoc
//
// @Summary     Gets a user
// @Description Gets a user by its id from the DB and returns it
// @Produce     json
// @Accept      json
// @Param       userId path     int         true "User id"
// @Param       body   body     models.User true "Request body"
// @Success     200    {object} models.User
// @Failure     400    {object} echo.HTTPError
// @Failure     404    {object} echo.HTTPError
// @Failure     500    {object} echo.HTTPError
// @Router      /users/{userId} [post]
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

	userToModify, err := h.repository.GetById(context.TODO(), userID.String())
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

// DeleteUserByID godoc
//
// @Summary     Deletes a user
// @Description Deletes a user by its id from the DB
// @Accept      json
// @Param       userId path int true "User id"
// @Success     204
// @Failure     400 {object} echo.HTTPError
// @Failure     500 {object} echo.HTTPError
// @Router      /users/{userId} [post]
func (h httpHandler) DeleteUserByID(c echo.Context) error {
	idStr := c.Param("userId")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid ID %s", idStr))
	}

	if err := h.repository.DeleteById(context.TODO(), userID.String()); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
