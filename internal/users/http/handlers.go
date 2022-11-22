// @tag.name        Users
// @tag.description Users API
package http

import (
	"context"
	"fmt"
	"net/http"
	httpErrors "user-microservice/internal/errors/http"
	"user-microservice/internal/models"
	"user-microservice/internal/pagination"
	"user-microservice/internal/users"
	userPS "user-microservice/internal/users/pubsub"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

var _ = echo.HTTPError{}

type httpHandler struct {
	repository       users.Repository
	pubsubRepository userPS.PubSub
}

var _ users.Handler = httpHandler{}
var _ users.Handler = (*httpHandler)(nil)

// NewHttpHandler - returns a new user http handler initialized with the repository
func NewHttpHandler(usersRepository users.Repository, pubsubRepository userPS.PubSub) users.Handler {
	return &httpHandler{usersRepository, pubsubRepository}
}

// CreateUser godoc
//
// @Summary     Create a user
// @Description Creates a new user and inserts it in the DB
// @Tags        Users
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

	// Notify user creation
	go func() {
		if err := h.pubsubRepository.NotifyUserCreation(ctx, *res); err != nil {
			logrus.Errorf("Error in users/http.CreateUser -> could not notify user creation: %s", err)
		}
	}()

	return c.JSON(http.StatusCreated, res)
}

// GetAllUsers godoc
//
// @Summary     Gets paginated users
// @Description Gets a paginated users list from the db and returns it
// @Tags        Users
// @Produce     json
// @Param       page      query    int    false "Page to retrieve" default(1)  minimum(1) example(2)
// @Param       size      query    int    false "Page size"        default(10) minimum(1) example(3)
// @Param       firstName query    string false "FirstName filter" example(Alice)
// @Param       lastName  query    string false "LastName filter"  example(Tingo)
// @Param       email     query    string false "Email filter"     example(alicetingo@example.com) format(email)
// @Param       nickname  query    string false "Nickname filter"  example(atingo)
// @Param       country   query    string false "Country filter"   example(DE)
// @Success     200       {object} models.PaginatedUsers
// @Failure     400       {object} echo.HTTPError
// @Failure     500       {object} echo.HTTPError
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
// @Tags        Users
// @Produce     json
// @Param       userId path     string true "User id" example(ddd50d89-0cf4-4d35-b8e8-51a2b5a06ce4) format(uuid)
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
		logrus.Errorf("Testing: %s", err)
		return err
	}

	return c.JSON(http.StatusOK, user)
}

// UpdateUserByID godoc
//
// @Summary     Updates a user
// @Description Updates a user by its id with the given body data
// @Tags        Users
// @Produce     json
// @Accept      json
// @Param       userId path     string      true "User id" example(7f598128-fb35-4ced-b80f-c5b5f66bd583) format(uuid)
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

	ctx := context.TODO()
	userToModify, err := h.repository.GetById(ctx, userID.String())
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

	res, err := h.repository.Update(ctx, *userToModify)
	if err != nil {
		return err
	}

	// Notify user update
	go func() {
		if err := h.pubsubRepository.NotifyUserUpdate(ctx, *res); err != nil {
			logrus.Errorf("Error in users/http.UpdateUserByID -> could not notify user update -> %s", err)
		}
	}()

	return c.JSON(http.StatusOK, res)
}

// DeleteUserByID godoc
//
// @Summary     Deletes a user
// @Description Deletes a user by its id from the DB
// @Tags        Users
// @Accept      json
// @Param       userId path string true "User id" format(uuid) example(5cace01f-45c3-49f0-a725-c22866874095)
// @Success     204
// @Failure     400 {object} echo.HTTPError
// @Failure     500 {object} echo.HTTPError
// @Router      /users/{userId} [delete]
func (h httpHandler) DeleteUserByID(c echo.Context) error {
	idStr := c.Param("userId")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid ID %s", idStr))
	}

	ctx := context.TODO()
	if err := h.repository.DeleteById(ctx, userID.String()); err != nil {
		return err
	}

	// Notify user deletion
	go func() {
		if err := h.pubsubRepository.NotifyUserDeletion(ctx, userID.String()); err != nil {
			logrus.Errorf("Error in users/http.DeleteUserByID -> could not notify user deletion: %s", err)
		}
	}()

	return c.NoContent(http.StatusNoContent)
}
