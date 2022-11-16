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
	userHttp "user-microservice/internal/users/http"
	"user-microservice/internal/users/mock"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//TODO: Add tests here

func TestCreateUser(t *testing.T) {
	for _, tc := range []struct {
		name              string
		body              string
		mockedUser        *models.User
		expectedErrorCode int
		expectedError     error
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
		},
		{
			"Create user with empty body",
			`{}`,
			nil,
			http.StatusBadRequest,
			echo.NewHTTPError(http.StatusBadRequest, nil),
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

			mockUserRepo.EXPECT().Create(ctx, gomock.Any()).Return(tc.mockedUser, tc.expectedError).Times(1)

			// When
			err := userHandler.CreateUser(echoCtx)

			// Then
			if tc.expectedError != nil {
				echoError, isOK := err.(*echo.HTTPError)
				if assert.True(t, isOK) {
					assert.Equal(t, tc.expectedErrorCode, echoError.Code)
				} else {
					assert.Equalf(t, tc.expectedErrorCode, rec.Code, "Expected error code to be %d, but was %d", tc.expectedErrorCode, rec.Code)
					require.Error(t, err)
					assert.Equal(t, tc.expectedError, err)
				}
			} else {
				assert.Equalf(t, http.StatusCreated, rec.Code, "Expected status code to be %d, but was %d", http.StatusCreated, rec.Code)
				var body models.User
				err := json.Unmarshal(rec.Body.Bytes(), &body)
				require.NoError(t, err, "Expected no error when unmarshaling body")

				//TODO: Assert body here
				assert.NotEmpty(t, body.ID.String(), "Expected ID not to be empty")
				assert.Equalf(t, tc.mockedUser.FirstName, body.FirstName, "Expected FirstName to be equal to %s, but was %s", tc.mockedUser.FirstName, body.FirstName)
			}
		})
	}
}
