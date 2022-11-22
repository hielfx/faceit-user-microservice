package mongodb_test

import (
	"context"
	"testing"
	"time"
	"user-microservice/internal/models"
	"user-microservice/internal/pagination"
	"user-microservice/internal/testutils"
	"user-microservice/internal/users/repository/mongodb"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

var dbClientTest *mongo.Client

func TestMain(m *testing.M) {
	dbClientTest = new(mongo.Client)
	testutils.ExecuteTestMain(m, dbClientTest)
}

type isErrorFunc func(error) bool

func defaultIsErrorFunc(error) bool { return true }

func TestMongoDBRepository_Create(t *testing.T) {
	for _, tc := range []struct {
		name          string
		user          models.User
		expectedError bool
		isErrorFunc
	}{
		{
			"Create user successfully",
			models.User{
				FirstName: "First Name",
				LastName:  "Last name",
				Nickname:  "Nickname",
				Password:  "Password",
				Email:     "Email",
				Country:   "Country",
			},
			false,
			defaultIsErrorFunc,
		},
		{
			"Create user successfully overrides ID and Dates",
			models.User{
				ID:        uuid.New().String(),
				FirstName: "First Name",
				LastName:  "Last name",
				Nickname:  "Nickname",
				Password:  "Password",
				Email:     "Email",
				Country:   "Country",
				CreatedAt: time.Now().Add(-2 * time.Hour),
				UpdatedAt: time.Now().Add(-2 * time.Hour),
			},
			false,
			defaultIsErrorFunc,
		},
		{
			"Create user with already existing ID",
			models.User{
				ID:        "29621CF9C9894266A5A2085FD99A75E1",
				FirstName: "Already existing user id First Name",
				LastName:  "Already existing user id Last name",
				Nickname:  "Already existing user id Nickname",
				Password:  "Already existing user id Password",
				Email:     "Already existing user id Email",
				Country:   "Already existing user id Country",
				CreatedAt: time.Now().Add(-2 * time.Hour),
				UpdatedAt: time.Now().Add(-2 * time.Hour),
			},
			false,
			defaultIsErrorFunc,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			//Given
			now := time.Now().UTC().Add(-1 * time.Minute)
			db := dbClientTest
			mongoRepo := mongodb.NewMongoDBRepository(testutils.GetDatabaseFromClient(db))
			ctx := context.TODO()

			//When
			res, err := mongoRepo.Create(ctx, tc.user)

			//Then
			if tc.expectedError {
				assert.Error(t, err, "Expected error")
				assert.Nilf(t, res, "Expected res to be nil, but was %s", res)
				assert.True(t, tc.isErrorFunc(err), "Different error from expected one: %s", err)
			} else {
				assert.Nilf(t, err, "Expected err to be nil, but was %s", err)
				require.NotNil(t, res, "Expected res not to be nil")

				assert.Equalf(t, tc.user.FirstName, res.FirstName, "Expected FirstName to be %s, but was %s", tc.user.FirstName, res.FirstName)
				assert.Equalf(t, tc.user.LastName, res.LastName, "Expected LastName to be %s, but was %s", tc.user.LastName, tc.user.LastName)
				assert.Equalf(t, tc.user.Nickname, res.Nickname, "Expected Nickname to be %s, but was %s", tc.user.Nickname, tc.user.Nickname)
				assert.Equalf(t, tc.user.Password, res.Password, "Expected Password to be %s, but was %s", tc.user.Password, tc.user.Password)
				assert.Equalf(t, tc.user.Email, res.Email, "Expected Email to be %s, but was %s", tc.user.Email, tc.user.Email)
				assert.Equalf(t, tc.user.Country, res.Country, "Expected Country to be %s, but was %s", tc.user.Country, tc.user.Country)
				assert.Truef(t, res.CreatedAt.After(now), "Expected CreatedAt to be after %s, but was %s", now, res.CreatedAt)
				assert.Truef(t, res.UpdatedAt.After(now), "Expected UpdatedAt to be after %s, but was %s", now, res.UpdatedAt)

				//We assert this because the creation method should override CreatedAt, UpdatedAt and ID
				assert.NotEqual(t, tc.user.CreatedAt, res.CreatedAt, "Expected CreatedAt not to be equal")
				assert.NotEqual(t, tc.user.UpdatedAt, res.UpdatedAt, "Expected UpdatedAt not to be equal")
				assert.NotEqual(t, tc.user.ID, res.ID, "Expected ID not to be equal")
			}
		})
	}
}

func TestMongoDBRepository_GetById(t *testing.T) {
	createdAt, err := time.Parse("2006-01-02T15:04:05Z", "2016-05-18T16:00:00Z")
	require.NoError(t, err, "Expected no error when initializing createdAt")
	updatedAt, err := time.Parse("2006-01-02T15:04:05Z", "2016-05-18T16:00:00Z")
	require.NoError(t, err, "Expected no error when initializing updatedAt")
	for _, tc := range []struct {
		name           string
		id             string
		expectedResult *models.User
		expectedError  error
	}{
		{
			"Get user by ID successfully",
			"ddd50d89-0cf4-4d35-b8e8-51a2b5a06ce4",
			&models.User{
				ID:        "ddd50d89-0cf4-4d35-b8e8-51a2b5a06ce4",
				FirstName: "Alice",
				LastName:  "Tingo",
				Nickname:  "atingo",
				Password:  "Already inserted user password 1",
				Email:     "alicetingo@example.com",
				Country:   "DE",
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			},
			nil,
		},
		{
			"Get not found user by ID",
			"C43DF343FFB343DA9BB0B08B81E6FCD9",
			nil,
			mongo.ErrNoDocuments,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Given
			now := time.Now().UTC().Add(-1 * time.Minute)
			mongoRepo := mongodb.NewMongoDBRepository(testutils.GetDatabaseFromClient(dbClientTest))
			ctx := context.TODO()

			// When
			res, err := mongoRepo.GetById(ctx, tc.id)

			// Then
			if tc.expectedError != nil {
				assert.Nilf(t, res, "Expected res to be nil but was %s", res)
				assert.Error(t, err)
				assert.Equalf(t, tc.expectedError, err, "Expected err to be %s, but was %s", tc.expectedError, err)
			} else {
				assert.NoErrorf(t, err, "Expected no error, but was %s", err)
				require.NotNil(t, res, "Expected res not to be nil")
				assert.Equalf(t, tc.id, res.ID, "Expected ID to be %s, but was %s", tc.id, res.ID)
				assert.Equalf(t, tc.expectedResult.FirstName, res.FirstName, "Expected FirstName to be %s, but was %s", tc.expectedResult.FirstName, res.FirstName)
				assert.Equalf(t, tc.expectedResult.LastName, res.LastName, "Expected LastName to be %s, but was %s", tc.expectedResult.LastName, res.LastName)
				assert.Equalf(t, tc.expectedResult.Nickname, res.Nickname, "Expected Nickname to be %s, but was %s", tc.expectedResult.Nickname, res.Nickname)
				assert.Equalf(t, tc.expectedResult.Password, res.Password, "Expected Password to be %s, but was %s", tc.expectedResult.Password, res.Password)
				assert.Equalf(t, tc.expectedResult.Email, res.Email, "Expected Email to be %s, but was %s", tc.expectedResult.Email, res.Email)
				assert.Equalf(t, tc.expectedResult.Country, res.Country, "Expected Country to be %s, but was %s", tc.expectedResult.Country, res.Country)
				assert.Equalf(t, tc.expectedResult.CreatedAt, res.CreatedAt, "Expected CreatedAt to be %s, but was %s", tc.expectedResult.CreatedAt, res.CreatedAt)
				assert.Equalf(t, tc.expectedResult.UpdatedAt, res.UpdatedAt, "Expected UpdatedAt to be %s, but was %s", tc.expectedResult.UpdatedAt, res.UpdatedAt)

				assert.True(t, res.CreatedAt.Before(now), "Expected CreatedAt to be before now but was %s", res.CreatedAt)
				assert.True(t, res.UpdatedAt.Before(now), "Expected UpdatedAt to be before now but was %s", res.UpdatedAt)
			}

		})
	}
}

func TestMongoDBRepository_DeleteById(t *testing.T) {
	for _, tc := range []struct {
		name          string
		id            string
		expectedError error
	}{
		{
			"Delete user by id successfully",
			"f4c9c17e-c260-4a0b-a1f1-a3f3ef6a3739",
			nil,
		},
		{
			"Delete user not found with error",
			uuid.New().String(),
			nil,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Given
			mongoRepository := mongodb.NewMongoDBRepository(testutils.GetDatabaseFromClient(dbClientTest))
			ctx := context.TODO()

			// When
			err := mongoRepository.DeleteById(ctx, tc.id)

			// Then
			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equalf(t, tc.expectedError, err, "Expected error to be %s, but was %s", tc.expectedError, err)
			} else {
				require.NoError(t, err)
				//retrieve element from id and check the error
				fromDB, err := mongoRepository.GetById(ctx, tc.id)
				require.Error(t, err)
				require.Nil(t, fromDB)
				assert.Equal(t, mongo.ErrNoDocuments, err, "Expected error to be %s, but was %s", mongo.ErrNoDocuments, err)
			}
		})
	}
}

type expected struct {
	err          error
	equalLength  bool
	equalPage    bool
	equalSize    bool
	hasMore      bool
	checkFilters bool
}

func TestMongoDBRepository_GetPaginatedUsers(t *testing.T) {
	for _, tc := range []struct {
		name       string
		pgOpts     pagination.PaginationOptions
		filterOpts models.UserFilters
		expected   expected
	}{
		{
			"Get paginated users with all pgOptions success",
			pagination.PaginationOptions{
				Page: 2,
				Size: 2,
				// OrderBy:   "_id",
				// SortOrder: pagination.SortOrderAsc,
			},
			models.UserFilters{},
			expected{
				equalLength: true,
				equalPage:   true,
				equalSize:   true,
				hasMore:     true,
			},
		},
		{
			"Get paginated users with negative page",
			pagination.PaginationOptions{
				Page: -1,
				Size: 2,
			},
			models.UserFilters{},
			expected{
				equalSize:   true,
				equalLength: true,
				hasMore:     true,
			},
		},
		{
			"Get paginated users with negative size",
			pagination.PaginationOptions{
				Page: 2,
				Size: -1,
			},
			models.UserFilters{},
			expected{
				equalPage: true,
				hasMore:   false,
			},
		},
		{
			"Get paginated users with negative page and size",
			pagination.PaginationOptions{
				Page: -1,
				Size: -1,
			},
			models.UserFilters{},
			expected{
				hasMore: false,
			},
		},
		{
			"Get paginated users with zero values",
			pagination.PaginationOptions{},
			models.UserFilters{},
			expected{
				hasMore: false,
			},
		},
		{
			"Get paginated users with high size",
			pagination.PaginationOptions{
				Page: 1,
				Size: 100,
			},
			models.UserFilters{},
			expected{
				hasMore:   false,
				equalPage: true,
				equalSize: true,
			},
		},
		{
			"Get paginated users filters by firstName",
			pagination.PaginationOptions{
				Page: 1,
				Size: 2,
			},
			models.UserFilters{
				FirstName: "Already inserted user first name 5",
			},
			expected{
				hasMore:      false,
				equalPage:    true,
				equalSize:    true,
				checkFilters: true,
			},
		},
		{
			"Get paginated users filters by lastName",
			pagination.PaginationOptions{
				Page: 1,
				Size: 2,
			},
			models.UserFilters{
				LastName: "Tingo",
			},
			expected{
				hasMore:      false,
				equalPage:    true,
				equalSize:    true,
				checkFilters: true,
				equalLength:  true,
			},
		},
		{
			"Get paginated users filters by country",
			pagination.PaginationOptions{
				Page: 1,
				Size: 2,
			},
			models.UserFilters{
				Country: "ES",
			},
			expected{
				hasMore:      false,
				equalPage:    true,
				equalSize:    true,
				checkFilters: true,
				equalLength:  true,
			},
		},
		//TODO: Check by other filters
		{
			"Get paginated users filters several filters with results",
			pagination.PaginationOptions{
				Page: 1,
				Size: 2,
			},
			models.UserFilters{
				Country:  "ES",
				LastName: "Tingo",
			},
			expected{
				hasMore:      false,
				equalPage:    true,
				equalSize:    true,
				checkFilters: true,
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			//Given
			mongoRepository := mongodb.NewMongoDBRepository(testutils.GetDatabaseFromClient(dbClientTest))
			ctx := context.TODO()

			//When
			res, err := mongoRepository.GetPaginatedUsers(ctx, tc.pgOpts, tc.filterOpts)

			//Then
			if tc.expected.err != nil {
				require.Error(t, err)
				require.Equal(t, tc.expected.err, err)
			} else {
				require.NoError(t, err)

				if tc.expected.equalPage {
					assert.Equalf(t, tc.pgOpts.Page, res.CurrentPage, "Expected CurrentPage to be equal to %d, but was %d", tc.pgOpts.Page, res.CurrentPage)
				} else {
					assert.NotEqualf(t, tc.pgOpts.Page, res.CurrentPage, "Expected CurrentPage to be different to %d", tc.pgOpts.Page)
				}
				assert.Greaterf(t, res.CurrentPage, 0, "Expected CurrentPage to be greater than %d, but was %d", 0, res.CurrentPage)

				if tc.expected.equalSize {
					assert.Equalf(t, tc.pgOpts.Size, res.Size, "Expected Size to be equal to %d, but was %d", tc.pgOpts.Size, res.Size)
				} else {
					assert.NotEqualf(t, tc.pgOpts.Size, res.Size, "Expected Size to be different to %d", tc.pgOpts.Size)
				}
				assert.Greaterf(t, res.Size, 0, "Expected size to be greater than %d, but was %d", 0, res.Size)

				if tc.expected.equalLength {
					assert.Equalf(t, tc.pgOpts.Size, len(res.Users), "Expected Users length to be equal to %d, but was %d", tc.pgOpts.Size, len(res.Users))
				} else {
					assert.NotEqualf(t, tc.pgOpts.Size, len(res.Users), "Expected Users length to be different to %d", tc.pgOpts.Size)
				}
				assert.Greaterf(t, len(res.Users), 0, "Expected Users length to be greater than %d, but was %d", 0, len(res.Users))

				assert.Equalf(t, tc.expected.hasMore, res.HasMore, "Expected HasMore to be %t but was %t", tc.expected.hasMore, res.HasMore)
				assert.Greater(t, res.TotalCount, int64(0), "Expected TotalCount to be greater than %d, but was %d", int64(0), res.TotalCount)
				assert.Greaterf(t, res.TotalPages, int64(0), "Expected TotalPages to be greater than %d, but was %d", int64(0), res.TotalPages)

				if tc.expected.checkFilters {
					for _, user := range res.Users {
						if tc.filterOpts.FirstName != "" {
							assert.Equalf(t, tc.filterOpts.FirstName, user.FirstName, "Expected FirstName to be %s, but was %s", tc.filterOpts.FirstName, user.FirstName)
						}
						if tc.filterOpts.LastName != "" {
							assert.Equalf(t, tc.filterOpts.LastName, user.LastName, "Expected LastName to be %s, but was %s", tc.filterOpts.LastName, user.LastName)
						}
						if tc.filterOpts.Email != "" {
							assert.Equalf(t, tc.filterOpts.Email, user.Email, "Expected Email to be %s, but was %s", tc.filterOpts.Email, user.Email)
						}
						if tc.filterOpts.Nickname != "" {
							assert.Equalf(t, tc.filterOpts.Nickname, user.Nickname, "Expected Nickname to be %s, but was %s", tc.filterOpts.Nickname, user.Nickname)
						}
						if tc.filterOpts.Country != "" {
							assert.Equalf(t, tc.filterOpts.Country, user.Country, "Expected Country to be %s, but was %s", tc.filterOpts.Country, user.Country)
						}
					}
				}
			}

		})
	}
}

func TestMongoDBRepository_UpdateUser(t *testing.T) {
	for _, tc := range []struct {
		name          string
		user          models.User
		expectedError error
	}{
		{
			"Update user successfully",
			models.User{
				ID:        "7f598128-fb35-4ced-b80f-c5b5f66bd583",
				FirstName: "Modified FirstName",
				LastName:  "Modified LastName",
				Nickname:  "Modified Nickname",
				Password:  "Modified Password",
				Email:     "Modified Email",
				Country:   "Modified Country",
			},
			nil,
		},
		{
			"Update user successfully with empty values",
			models.User{
				ID:        "ca1b8cc6-c34d-4959-aa10-4f6fe30b6a75",
				FirstName: "",
				LastName:  "",
				Nickname:  "",
				Password:  "",
				Email:     "",
				Country:   "",
			},
			nil,
		},
		{
			"Update user successfully without modifying dates",
			models.User{
				ID:        "2b875247-90a1-4d66-8c22-1a53a265d180",
				FirstName: "Modified FirstName",
				LastName:  "Modified LastName",
				Nickname:  "Modified Nickname",
				Password:  "Modified Password",
				Email:     "Modified Email",
				Country:   "Modified Country",
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			nil,
		},
		{
			"Update user with not found ID",
			models.User{
				ID: "B535F5174BE244DF820D3C3F886DA2A2",
			},
			mongo.ErrNoDocuments,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			//Given
			mongoRepo := mongodb.NewMongoDBRepository(testutils.GetDatabaseFromClient(dbClientTest))
			ctx := context.TODO()
			now := time.Now().UTC().Add(-1 * time.Minute)

			//When
			res, err := mongoRepo.Update(ctx, tc.user)

			//Then
			if tc.expectedError != nil {
				require.Error(t, err)
				assert.Nil(t, res)
				assert.Equalf(t, tc.expectedError, err, "Expected err to be %v, but was %v", tc.expectedError, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, res, "Expected res not to be nil")

				assert.Equalf(t, tc.user.ID, res.ID, "Expected ID to be %s, but was %s", tc.user.ID, res.ID)
				assert.Equalf(t, tc.user.FirstName, res.FirstName, "Expected FirstName to be %s, but was %s", tc.user.FirstName, res.FirstName)
				assert.Equalf(t, tc.user.LastName, res.LastName, "Expected LastName to be %s, but was %s", tc.user.LastName, res.LastName)
				assert.Equalf(t, tc.user.Nickname, res.Nickname, "Expected Nickname to be %s, but was %s", tc.user.Nickname, res.Nickname)
				assert.Equalf(t, tc.user.Password, res.Password, "Expected Password to be %s, but was %s", tc.user.Password, res.Password)
				assert.Equalf(t, tc.user.Email, res.Email, "Expected Email to be %s, but was %s", tc.user.Email, res.Email)
				assert.Equalf(t, tc.user.Country, res.Country, "Expected Country to be %s, but was %s", tc.user.Country, res.Country)

				// assert.Truef(t, res.CreatedAt.Before(now), "Expected CreatedAt to be before %s but was %s", now, res.CreatedAt)
				assert.Truef(t, res.UpdatedAt.After(now), "Expected UpdatedAt to be after %s but was %s", now, res.UpdatedAt)
			}
		})
	}
}
