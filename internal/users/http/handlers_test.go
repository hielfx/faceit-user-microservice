package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	httpErrors "user-microservice/internal/errors/http"
	"user-microservice/internal/models"
	"user-microservice/internal/testutils"
	userHttp "user-microservice/internal/users/http"
	"user-microservice/internal/users/mock"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestCreateUser(t *testing.T) {
	for _, tc := range []struct {
		name           string
		body           string
		mockedUser     *models.User
		expectedCode   int
		expectedError  error
		shouldExecCall bool
	}{
		{
			"Create user successfully",
			`{
				"firstName": "CreateUser FirstName",
				"lastName": "CreateUser LastName",
				"nickname": "CreateUser Nickname",
				"password": "CreateUser Password",
				"email": "CreateUser Email",
				"country": "CreateUser Country"
			}`,
			&models.User{
				ID:        uuid.New(),
				FirstName: "CreateUser FirstName",
				LastName:  "CreateUser LastName",
				Nickname:  "CreateUser Nickname",
				Password:  "CreateUser Password",
				Email:     "CreateUser Email",
				Country:   "CreateUser Country",
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			0,
			nil,
			true,
		},
		{
			"Create user with empty body",
			`{}`,
			nil,
			http.StatusBadRequest,
			echo.NewHTTPError(http.StatusBadRequest, httpErrors.ErrInvalidBody),
			false,
		},
		{
			"Create user with invalid body",
			`invalid body`,
			nil,
			http.StatusBadRequest,
			echo.NewHTTPError(http.StatusBadRequest, nil),
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
			echo.NewHTTPError(http.StatusBadRequest),
			false,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Given
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserRepo := mock.NewMockRepository(ctrl)
			userHandler := userHttp.NewHttpHandler(mockUserRepo)

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

			// When
			err := userHandler.CreateUser(echoCtx)

			// Then
			if tc.expectedError != nil {
				echoError, isOK := err.(*echo.HTTPError)
				if assert.True(t, isOK) {
					assert.Equal(t, tc.expectedCode, echoError.Code)
				} else {
					assert.Equalf(t, tc.expectedCode, rec.Code, "Expected error code to be %d, but was %d", tc.expectedCode, rec.Code)
					require.Error(t, err)
					assert.Equal(t, tc.expectedError, err)
				}
			} else {
				assert.Equalf(t, http.StatusCreated, rec.Code, "Expected status code to be %d, but was %d", http.StatusCreated, rec.Code)
				var body models.User
				err := json.Unmarshal(rec.Body.Bytes(), &body)
				require.NoErrorf(t, err, "Expected no error when unmarshaling body, but was %s", err)

				testutils.AssertUserBody(t, *tc.mockedUser, body, testutils.AssertUserConfig{})
				//TODO: Assert password here
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	userUUID := uuid.New()
	for _, tc := range []struct {
		name           string
		id             string
		mockedId       uuid.UUID
		mockedError    error
		expectedError  error
		expectedCode   int
		shouldCallRepo bool
	}{
		{
			"Delete user successfully",
			userUUID.String(),
			userUUID,
			nil,
			nil,
			http.StatusNoContent,
			true,
		},
		{
			"Delete user with error",
			userUUID.String(),
			userUUID,
			errors.New("homemade error"),
			errors.New("homemade error"),
			http.StatusInternalServerError,
			true,
		},
		{
			"Delete user with wrong id",
			"invalid-user-id",
			userUUID,
			nil,
			echo.NewHTTPError(http.StatusBadRequest, "Invalid ID invalid-user-id"),
			http.StatusBadRequest,
			false,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			//Given
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserRepo := mock.NewMockRepository(ctrl)
			userHandler := userHttp.NewHttpHandler(mockUserRepo)

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
		mockedID       uuid.UUID
		mockedUser     *models.User
		mockedError    error
		expectedCode   int
		expectedError  error
		shouldCallMock bool
	}{
		{
			"Get user successfully by id",
			userUUID.String(),
			userUUID,
			&models.User{
				ID:        userUUID,
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
			userUUID,
			nil,
			mongo.ErrNilDocument,
			http.StatusNotFound,
			echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("User not found for ID %s", userUUID.String())),
			true,
		},
		{
			"Get user invalid id error",
			"invalid-id",
			userUUID,
			nil,
			nil,
			http.StatusBadRequest,
			echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID invalid-id"),
			false,
		},
		{
			"Get user internal server error",
			userUUID.String(),
			userUUID,
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
			userHandler := userHttp.NewHttpHandler(userRepo)

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
	for _, tc := range []struct {
		name             string
		id               string
		mockedID         uuid.UUID
		body             string
		mockedUser       models.User
		expectedCode     int
		mockedError      error
		expectedError    error
		shouldCallCreate bool
		shouldCallGet    bool
	}{
		{
			"Update user by ID successfully",
			userID.String(),
			userID,
			`{
				"firstName": "Update user FirstName",
				"lastName": "Update user LastName",
				"nickname": "Update user Nickname",
				"password": "Updated Password",
				"email": "Updated Email",	
				"country": "Updated Country"
			}`,
			models.User{
				ID:        userID,
				FirstName: "Update user FirstName",
				LastName:  "Update user LastName",
				Nickname:  "Update user Nickname",
				Password:  "Updated Password",
				Email:     "Updated Email",
				Country:   "Updated Country",
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
			"Update user with wrong id",
			"wrong-id",
			userID,
			"{}",
			models.User{},
			http.StatusBadRequest,
			nil,
			echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID wrong-id"),
			false,
			false,
		},
		{
			"Update user with empty body",
			userID.String(),
			userID,
			`{}`,
			models.User{},
			http.StatusBadRequest,
			nil,
			echo.NewHTTPError(http.StatusBadRequest, httpErrors.ErrInvalidBody),
			false,
			false,
		},
		{
			"Update user with invalid body",
			userID.String(),
			userID,
			"invalid-body",
			models.User{},
			http.StatusBadRequest,
			nil,
			echo.NewHTTPError(http.StatusBadRequest, nil),
			false,
			false,
		},
		{
			"Update user with invalid fields",
			userID.String(),
			userID,
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
			echo.NewHTTPError(http.StatusBadRequest, nil),
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
			userHandler := userHttp.NewHttpHandler(userRepo)

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
			userRepo.EXPECT().GetById(context.TODO(), tc.mockedID).Return(&tc.mockedUser, tc.mockedError).AnyTimes()

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
			}
		})
	}
}
