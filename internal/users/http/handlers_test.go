package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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
			`{"firstName": "Create user FirstName"}`,
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
			echo.NewHTTPError(http.StatusBadRequest, nil),
			true, //TODO: Change to false once it's validated
		},
		{
			"Create user with invalid body",
			`invalid body`,
			nil,
			http.StatusBadRequest,
			echo.NewHTTPError(http.StatusBadRequest, nil),
			false,
		},
		//TODO: Add more test cases
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
				require.NoError(t, err, "Expected no error when unmarshaling body")

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
			http.StatusOK,
			true,
		},
		{
			"Delete non existing user with error",
			userUUID.String(),
			userUUID,
			mongo.ErrNoDocuments,
			nil,
			http.StatusOK,
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
				if assert.True(t, isOK) {
					require.NotNil(t, echoError, "Expected echoError not to be nil")
					assert.Equalf(t, tc.expectedCode, echoError.Code, "Expected error code to be %d, but was %d", tc.expectedCode, echoError.Code)
					expectedEchoError, isOK := tc.expectedError.(*echo.HTTPError)
					if assert.True(t, isOK) {
						require.NotNil(t, expectedEchoError, "Expected expectedEchoError not to be nil")
						assert.Equal(t, echoError, expectedEchoError, "Expected expectedEchoError to be equal to %s, but was %s", expectedEchoError, echoError)
					}
				} else {
					assert.Equalf(t, tc.expectedCode, rec.Code, "Expected error code to be %d, but was %d", tc.expectedCode, rec.Code)
					assert.Equal(t, tc.expectedError, err)
				}
			} else {
				assert.Equalf(t, http.StatusOK, tc.expectedCode, "Expected status code to be %d, but was %d", http.StatusOK, tc.expectedCode)
				body := rec.Body.String()
				assert.Emptyf(t, body, "Expected body to be empty, but was %s", body)
			}
		})
	}
}

func TestGetUserByID(t *testing.T) {
	userUUID := uuid.New()
	_ = userUUID //TODO: Delete this
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
		//TODO: Add test cases here
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
				if assert.True(t, isOK) {
					assert.Equalf(t, tc.expectedCode, echoError.Code, "Expected code to be %d, but was %d", tc.expectedCode, echoError.Code)
					expectedEchoError, isOK := tc.expectedError.(*echo.HTTPError)
					if assert.True(t, isOK) {
						assert.Equalf(t, expectedEchoError, echoError, "Expected expectedEchoError to be %s, but was %s", expectedEchoError, echoError)
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
					//TODO: Assert password here
				}
			}
		})
	}
}
