package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
	httpErrors "user-microservice/internal/errors/http"
	"user-microservice/internal/models"
	"user-microservice/internal/pagination"
	"user-microservice/internal/testutils"
	userHttp "user-microservice/internal/users/http"
	"user-microservice/internal/users/mock"
	usersPubSub "user-microservice/internal/users/pubsub"

	"github.com/go-redis/redismock/v8"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestCreateUser(t *testing.T) {
	validBody := `{
		"firstName": "CreateUser FirstName",
		"lastName": "CreateUser LastName",
		"nickname": "CreateUser Nickname",
		"password": "CreateUser Password",
		"email": "CreateUser Email",
		"country": "CreateUser Country"
	}`
	for _, tc := range []struct {
		name              string
		body              string
		mockedUser        *models.User
		expectedCode      int
		mockedError       error
		expectedError     error
		shouldExecCall    bool
		shouldExecPublish bool
	}{
		{
			"Create user successfully",
			validBody,
			&models.User{
				ID:        uuid.New().String(),
				FirstName: "CreateUser FirstName",
				LastName:  "CreateUser LastName",
				Nickname:  "CreateUser Nickname",
				Password:  "CreateUser Password",
				Email:     "CreateUser Email",
				Country:   "CreateUser Country",
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			http.StatusOK,
			nil,
			nil,
			true,
			true,
		},
		{
			"Create user with empty body",
			`{}`,
			nil,
			http.StatusBadRequest,
			nil,
			echo.NewHTTPError(http.StatusBadRequest, httpErrors.ErrInvalidBody),
			false,
			false,
		},
		{
			"Create user with invalid body",
			`invalid body`,
			nil,
			http.StatusBadRequest,
			nil,
			echo.NewHTTPError(http.StatusBadRequest, nil),
			false,
			false,
		},
		{
			"Create user with invalid fields",
			`{
				"firstName": true,
				"lastName": 124.78,
				"nickname": -87,
				"password": false,
				"email": "Updated Email",	
				"country": {}
			}`,
			nil,
			http.StatusBadRequest,
			nil,
			echo.NewHTTPError(http.StatusBadRequest, nil),
			false,
			false,
		},
		{
			"Create user with internal server error",
			validBody,
			nil,
			http.StatusInternalServerError,
			errors.New("homemade error"),
			errors.New("homemade error"),
			true,
			false,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Given
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			redisDB, redisMock := redismock.NewClientMock()
			mockUserRepo := mock.NewMockRepository(ctrl)
			pubsubRepo := usersPubSub.NewPubSub(redisDB)
			userHandler := userHttp.NewHttpHandler(mockUserRepo, pubsubRepo)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(tc.body))
			req.Header.Set(echo.HeaderContentType, "application/json")
			rec := httptest.NewRecorder()
			e := echo.New()
			echoCtx := e.NewContext(req, rec)
			ctx := context.TODO()

			callTimes := 0
			if tc.shouldExecCall {
				callTimes = 1
			}
			mockUserRepo.EXPECT().Create(ctx, gomock.Any()).Return(tc.mockedUser, tc.expectedError).Times(callTimes)
			if tc.shouldExecPublish {
				encodedUser, err := json.Marshal(*tc.mockedUser)
				require.NoErrorf(t, err, "Expected no error when marshaling mocked user for publish, but was %s", err)
				redisMock.ExpectPublish(usersPubSub.TopicUserCreation, encodedUser)
			}

			// When
			err := userHandler.CreateUser(echoCtx)

			// Then
			if tc.expectedError != nil {
				testutils.AssertExpectedErrorsHttpReponse(t, tc.expectedCode, rec.Code, tc.expectedError, err)
			} else {
				assert.Equalf(t, http.StatusCreated, rec.Code, "Expected status code to be %d, but was %d", http.StatusCreated, rec.Code)
				var body models.User
				err := json.Unmarshal(rec.Body.Bytes(), &body)
				require.NoErrorf(t, err, "Expected no error when unmarshaling body, but was %s", err)

				testutils.AssertUserBody(t, *tc.mockedUser, body, testutils.AssertUserConfig{})

				if tc.shouldExecPublish {
					//Wait 3 secods for goroutine to finish
					<-time.Tick(3 * time.Second)
					redisError := redisMock.ExpectationsWereMet()
					require.NoErrorf(t, redisError, "Expected redisError not to be, but was %s", redisError)
				}
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	userUUID := uuid.New()
	for _, tc := range []struct {
		name              string
		id                string
		mockedId          string
		mockedError       error
		expectedError     error
		expectedCode      int
		shouldCallRepo    bool
		shouldExecPublish bool
	}{
		{
			"Delete user successfully",
			userUUID.String(),
			userUUID.String(),
			nil,
			nil,
			http.StatusNoContent,
			true,
			true,
		},
		{
			"Delete user with error",
			userUUID.String(),
			userUUID.String(),
			errors.New("homemade error"),
			errors.New("homemade error"),
			http.StatusInternalServerError,
			true,
			false,
		},
		{
			"Delete user with wrong id",
			"invalid-user-id",
			userUUID.String(),
			nil,
			echo.NewHTTPError(http.StatusBadRequest, "Invalid ID invalid-user-id"),
			http.StatusBadRequest,
			false,
			false,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			//Given
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			redisDB, redisMock := redismock.NewClientMock()
			pubsubRepo := usersPubSub.NewPubSub(redisDB)
			mockUserRepo := mock.NewMockRepository(ctrl)
			userHandler := userHttp.NewHttpHandler(mockUserRepo, pubsubRepo)

			req := httptest.NewRequest(http.MethodDelete, "/", nil)
			req.Header.Set(echo.HeaderContentType, "application/json")
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)
			c.SetPath("/api/v1/users/:userId")
			c.SetParamNames("userId")
			c.SetParamValues(tc.id)
			ctx := context.TODO()
			callTimes := 0
			if tc.shouldCallRepo {
				callTimes = 1
			}
			mockUserRepo.EXPECT().DeleteById(ctx, tc.mockedId).Return(tc.mockedError).Times(callTimes)
			if tc.shouldExecPublish {
				redisMock.ExpectPublish(usersPubSub.TopicUserDeletion, tc.mockedId)
			}

			//When
			err := userHandler.DeleteUserByID(c)

			//Then
			if tc.expectedError != nil {
				require.Error(t, err)
				echoError, isOK := err.(*echo.HTTPError)
				if isOK {
					require.NotNil(t, echoError, "Expected echoError not to be nil")
					assert.Equalf(t, tc.expectedCode, echoError.Code, "Expected error code to be %d, but was %d", tc.expectedCode, echoError.Code)
					expectedEchoError, isOK := tc.expectedError.(*echo.HTTPError)
					if isOK {
						require.NotNil(t, expectedEchoError, "Expected expectedEchoError not to be nil")
						assert.Equal(t, echoError, expectedEchoError, "Expected expectedEchoError to be equal to %s, but was %s", expectedEchoError, echoError)
					}
				} else {
					// assert.Equalf(t, tc.expectedCode, rec.Code, "Expected error code to be %d, but was %d", tc.expectedCode, rec.Code)
					assert.Equal(t, tc.expectedError, err)
				}
			} else {
				assert.Equalf(t, http.StatusNoContent, rec.Code, "Expected status code to be %d, but was %d", http.StatusNoContent, rec.Code)
				body := rec.Body.String()
				assert.Emptyf(t, body, "Expected body to be empty, but was %s", body)

				if tc.shouldExecPublish {
					// wait 3 seconds for goroutine to complete
					<-time.Tick(3 * time.Second)
					redisError := redisMock.ExpectationsWereMet()
					require.NoErrorf(t, redisError, "Expected redisError not to be, but was %s", redisError)
				}
			}
		})
	}
}

func TestGetUserByID(t *testing.T) {
	userUUID := uuid.New()
	now := time.Now().UTC().Add(-1 * time.Minute)
	for _, tc := range []struct {
		name           string
		id             string
		mockedID       string
		mockedUser     *models.User
		mockedError    error
		expectedCode   int
		expectedError  error
		shouldCallMock bool
	}{
		{
			"Get user successfully by id",
			userUUID.String(),
			userUUID.String(),
			&models.User{
				ID:        userUUID.String(),
				FirstName: "Retrieved user FirstName",
				LastName:  "Retrieved user LastName",
				Nickname:  "Retrieved user Nickname",
				Password:  "Retrieved user Password",
				Email:     "Retrieved user Email",
				Country:   "Retrieved user Country",
				CreatedAt: now,
				UpdatedAt: now,
			},
			nil,
			http.StatusOK,
			nil,
			true,
		},
		{
			"Get user id not found error",
			userUUID.String(),
			userUUID.String(),
			nil,
			mongo.ErrNilDocument,
			http.StatusNotFound,
			echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("User not found for ID %s", userUUID.String())),
			true,
		},
		{
			"Get user invalid id error",
			"invalid-id",
			userUUID.String(),
			nil,
			nil,
			http.StatusBadRequest,
			echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID invalid-id"),
			false,
		},
		{
			"Get user internal server error",
			userUUID.String(),
			userUUID.String(),
			nil,
			errors.New("homemade error"),
			http.StatusInternalServerError,
			errors.New("homemade error"),
			true,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			//Given
			ctrl := gomock.NewController(t)
			userRepo := mock.NewMockRepository(ctrl)
			redisDB, _ := redismock.NewClientMock()
			pubsubRepo := usersPubSub.NewPubSub(redisDB)
			userHandler := userHttp.NewHttpHandler(userRepo, pubsubRepo)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)
			c.SetPath("/api/v1/users/:userId")
			c.SetParamNames("userId")
			c.SetParamValues(tc.id)

			callTimes := 0
			if tc.shouldCallMock {
				callTimes = 1
			}
			userRepo.EXPECT().GetById(context.TODO(), tc.mockedID).Return(tc.mockedUser, tc.mockedError).Times(callTimes)

			//When
			err := userHandler.GetUserByID(c)

			//Then

			if tc.expectedError != nil {
				require.Error(t, err)
				echoError, isOK := err.(*echo.HTTPError)
				if isOK {
					assert.Equalf(t, tc.expectedCode, echoError.Code, "Expected code to be %d, but was %d", tc.expectedCode, echoError.Code)
					expectedEchoError, isOK := tc.expectedError.(*echo.HTTPError)
					if assert.True(t, isOK) {
						assert.Equalf(t, expectedEchoError, echoError, "Expected expectedEchoError to be %s, but was %s", expectedEchoError, echoError)
					}
				} else {
					require.Equalf(t, tc.expectedError, err, "Expected err to be %s, but was %s", tc.expectedError, err)
				}
			} else {
				require.Nil(t, err)
				assert.Equalf(t, http.StatusOK, rec.Code, "Expected status code to be %d, but was %d", http.StatusOK, rec.Code)

				var body models.User
				err := json.Unmarshal(rec.Body.Bytes(), &body)
				require.NoErrorf(t, err, "Expected no error when unmarshaling body, but was %s", err)

				testutils.AssertUserBody(t, *tc.mockedUser, body, testutils.AssertUserConfig{
					CreatedAt: &testutils.DateCheck{
						Option: testutils.DateCheckOptionEquals,
						Value:  tc.mockedUser.CreatedAt,
					},
					UpdatedAt: &testutils.DateCheck{
						Option: testutils.DateCheckOptionEquals,
						Value:  tc.mockedUser.UpdatedAt,
					},
					EqualPasswords: true,
				})
			}
		})
	}
}

func TestUpdateUserByID(t *testing.T) {
	userID := uuid.New()
	validBody := `{
		"firstName": "Update user FirstName",
		"lastName": "Update user LastName",
		"nickname": "Update user Nickname",
		"password": "Updated Password",
		"email": "Updated Email",	
		"country": "Updated Country"
	}`
	validUser := models.User{
		ID:        userID.String(),
		FirstName: "Update user FirstName",
		LastName:  "Update user LastName",
		Nickname:  "Update user Nickname",
		Password:  "Updated Password",
		Email:     "Updated Email",
		Country:   "Updated Country",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	for _, tc := range []struct {
		name              string
		id                string
		mockedID          string
		body              string
		mockedUser        models.User
		expectedCode      int
		mockedError       error
		mockedGetError    error
		expectedError     error
		shouldCallCreate  bool
		shouldCallGet     bool
		shouldExecPublish bool
	}{
		{
			"Update user by ID successfully",
			userID.String(),
			userID.String(),
			validBody,
			validUser,
			http.StatusOK,
			nil,
			nil,
			nil,
			true,
			true,
			true,
		},
		{
			"Update user with wrong id",
			"wrong-id",
			userID.String(),
			"{}",
			models.User{},
			http.StatusBadRequest,
			nil,
			nil,
			echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID wrong-id"),
			false,
			false,
			false,
		},
		{
			"Update user with empty body",
			userID.String(),
			userID.String(),
			`{}`,
			models.User{},
			http.StatusBadRequest,
			nil,
			nil,
			echo.NewHTTPError(http.StatusBadRequest, httpErrors.ErrInvalidBody),
			false,
			false,
			false,
		},
		{
			"Update user with invalid body",
			userID.String(),
			userID.String(),
			"invalid-body",
			models.User{},
			http.StatusBadRequest,
			nil,
			nil,
			echo.NewHTTPError(http.StatusBadRequest, nil),
			false,
			false,
			false,
		},
		{
			"Update user with invalid fields",
			userID.String(),
			userID.String(),
			`{
				"firstName": true,
				"lastName": 124.78,
				"nickname": -87,
				"password": false,
				"email": "Updated Email",	
				"country": {}
			}`,
			models.User{},
			http.StatusBadRequest,
			nil,
			nil,
			echo.NewHTTPError(http.StatusBadRequest, nil),
			false,
			false,
			false,
		},
		{
			"Update user with not found error",
			userID.String(),
			userID.String(),
			`{}`,
			models.User{},
			http.StatusNotFound,
			nil,
			mongo.ErrNilDocument,
			echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("User not found for ID %s", userID.String())),
			false,
			false,
			false,
		},
		{
			"Update user with internal server error by get",
			userID.String(),
			userID.String(),
			`{}`,
			models.User{},
			http.StatusInternalServerError,
			nil,
			errors.New("homemade error"),
			errors.New("homemade error"),
			false,
			true,
			false,
		},
		{
			"Update user with internal server error by update",
			userID.String(),
			userID.String(),
			validBody,
			validUser,
			http.StatusInternalServerError,
			errors.New("homemade error"),
			nil,
			errors.New("homemade error"),
			true,
			false,
			false,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			//Given
			ctrl := gomock.NewController(t)
			userRepo := mock.NewMockRepository(ctrl)
			redisDB, redisMock := redismock.NewClientMock()
			pubsubRepo := usersPubSub.NewPubSub(redisDB)
			userHandler := userHttp.NewHttpHandler(userRepo, pubsubRepo)

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.body))
			req.Header.Set(echo.HeaderContentType, "application/json")
			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)
			c.SetPath("/api/v1/users/:userId")
			c.SetParamNames("userId")
			c.SetParamValues(tc.id)

			callTimes := 0
			if tc.shouldCallCreate {
				callTimes = 1
			}
			userRepo.EXPECT().Update(context.TODO(), gomock.Any()).Return(&tc.mockedUser, tc.mockedError).Times(callTimes)
			userRepo.EXPECT().GetById(context.TODO(), tc.mockedID).Return(&tc.mockedUser, tc.mockedGetError).AnyTimes()
			if tc.shouldExecPublish {
				encodedUser, err := json.Marshal(tc.mockedUser)
				require.NoErrorf(t, err, "Expected no error when marshaling user to publish, but was %s", err)
				redisMock.ExpectPublish(usersPubSub.TopicUserUpdate, encodedUser)
			}

			//when
			err := userHandler.UpdateUserByID(c)

			//Then
			if tc.expectedError != nil {
				testutils.AssertExpectedErrorsHttpReponse(t, tc.expectedCode, rec.Code, tc.expectedError, err)
			} else {
				require.NoError(t, err)
				assert.Equalf(t, http.StatusOK, rec.Code, "Expecred status code to be %d, but was %d", http.StatusOK, rec.Code)

				var body models.User
				err := json.Unmarshal(rec.Body.Bytes(), &body)
				require.NoErrorf(t, err, "Expected no error when unmarshaling body, but was %s", err)

				testutils.AssertUserBody(t, tc.mockedUser, body, testutils.AssertUserConfig{
					EqualPasswords: true,
					UpdatedAt: &testutils.DateCheck{
						Option: testutils.DateCheckOptionEquals,
						Value:  tc.mockedUser.UpdatedAt,
					},
					CreatedAt: &testutils.DateCheck{
						Option: testutils.DateCheckOptionEquals,
						Value:  tc.mockedUser.CreatedAt,
					},
				})

				if tc.shouldExecPublish {
					//Wait 3 secods for goroutine to finish
					<-time.Tick(3 * time.Second)
					redisError := redisMock.ExpectationsWereMet()
					require.NoErrorf(t, redisError, "Expected redisError not to be, but was %s", redisError)
				}
			}
		})
	}
}

func TestGetAllUsers(t *testing.T) {
	for _, tc := range []struct {
		name           string
		pagination     pagination.PaginationOptions
		mockedRes      models.PaginatedUsers
		filters        map[string]string
		mockedFilters  models.UserFilters
		statusCode     int
		expectedError  error
		mockedError    error
		shouldCallRepo bool
	}{
		{
			"Get paginated users successfully",
			pagination.PaginationOptions{
				Page: 1,
				Size: 2,
			},
			models.PaginatedUsers{
				Paginated: pagination.Paginated{
					TotalCount:  10,
					TotalPages:  5,
					CurrentPage: 1,
					Size:        2,
					HasMore:     true,
				},
				Users: []models.User{
					{ID: uuid.New().String()},
					{ID: uuid.New().String()},
				},
			},
			map[string]string{},
			models.UserFilters{},
			http.StatusOK,
			nil,
			nil,
			true,
		},
		{
			"Get paginated users with filters",
			pagination.PaginationOptions{
				Page: 1,
				Size: 2,
			},
			models.PaginatedUsers{
				Paginated: pagination.Paginated{
					TotalCount:  10,
					TotalPages:  5,
					CurrentPage: 1,
					Size:        2,
					HasMore:     true,
				},
				Users: []models.User{
					{ID: uuid.New().String()},
					{ID: uuid.New().String()},
				},
			},
			map[string]string{
				"country": "UK",
			},
			models.UserFilters{
				Country: "UK",
			},
			http.StatusOK,
			nil,
			nil,
			true,
		},
		{
			"Get paginated users with get error",
			pagination.PaginationOptions{},
			models.PaginatedUsers{},
			map[string]string{},
			models.UserFilters{},
			http.StatusInternalServerError,
			errors.New("homemade error"),
			errors.New("homemade error"),
			true,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			//Given
			q := make(url.Values)
			q.Set("page", strconv.Itoa(tc.pagination.Page))
			q.Set("size", strconv.Itoa(tc.pagination.Size))
			for k, v := range tc.filters {
				q.Set(k, v)
			}
			req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)

			rec := httptest.NewRecorder()
			e := echo.New()
			c := e.NewContext(req, rec)
			// c.SetPath("/users?" + q.Encode())
			// paramNames := []string{"page", "size"}
			// paramValues := []string{strconv.Itoa(tc.pagination.Page), strconv.Itoa(tc.pagination.Size)}
			// for k, v := range tc.filters {
			// 	paramNames = append(paramNames, k)
			// 	paramValues = append(paramValues, v)
			// }
			// c.SetParamNames(paramNames...)
			// c.SetParamValues(paramValues...)

			ctrl := gomock.NewController(t)
			userRepo := mock.NewMockRepository(ctrl)
			redisDB, _ := redismock.NewClientMock()
			pubsubRepo := usersPubSub.NewPubSub(redisDB)
			h := userHttp.NewHttpHandler(userRepo, pubsubRepo)

			callTimes := 0
			if tc.shouldCallRepo {
				callTimes = 1
			}
			userRepo.EXPECT().GetPaginatedUsers(context.TODO(), tc.pagination, tc.mockedFilters).Return(tc.mockedRes, tc.mockedError).Times(callTimes)

			//When
			err := h.GetAllUsers(c)

			//Then
			if tc.expectedError != nil {
				testutils.AssertExpectedErrorsHttpReponse(t, tc.statusCode, rec.Code, tc.expectedError, err)
			} else {
				assert.Equalf(t, http.StatusOK, rec.Code, "Expected status code to be %d, but was %d", http.StatusOK, rec.Code)
				require.NoErrorf(t, err, "Expected no error but was %s", err)

				var body models.PaginatedUsers
				err := json.Unmarshal(rec.Body.Bytes(), &body)
				require.NoErrorf(t, err, "Expected no error when unmarshaling body, but was %s", err)
				assert.Equalf(t, tc.mockedRes, body, "Expected body to be %s, but was %s", tc.mockedRes, body)
			}
		})
	}
}
