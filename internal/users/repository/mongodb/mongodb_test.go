package mongodb_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
	"user-microservice/internal/models"
	"user-microservice/internal/users/repository/mongodb"

	"github.com/google/uuid"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var dbClientTest *mongo.Client

const databaseName = "test"

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		logrus.Fatalf("Could not connect to docker: %s", err)
		return
	}

	dir, err := os.Getwd()
	if err != nil {
		logrus.Fatalf("Error retrieving working directory: %s", err)
		return
	}

	//pull mongodb docker image
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mongo",
		Tag:        "6",
		Env:        []string{},
		Mounts: []string{
			dir + "/init_db.js:/docker-entrypoint-initdb.d/init_db.js",
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		logrus.Fatalf("Could start docker reouser: %s", err)
		return
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	err = pool.Retry(func() error {
		var err error
		dbClientTest, err = mongo.Connect(
			context.TODO(),
			options.Client().ApplyURI(
				fmt.Sprintf("mongodb://localhost:%s", resource.GetPort("27017/tcp")),
			),
		)
		if err != nil {
			return err
		}
		return dbClientTest.Ping(context.TODO(), nil)
	})
	if err != nil {
		logrus.Fatalf("Could not connect to docker: %s", err)
	}

	// run tests
	code := m.Run()

	// When you're done, kill and remove the container
	if err = pool.Purge(resource); err != nil {
		logrus.Fatalf("Could not purge resource: %s", err)
	}

	// disconnect mongodb client
	if err = dbClientTest.Disconnect(context.TODO()); err != nil {
		panic(err)
	}

	os.Exit(code)

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
				ID:        uuid.New(),
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
				ID:        uuid.MustParse("29621CF9-C989-4266-A5A2-085FD99A75E1"),
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
			now := time.Now()
			mongoRepo := mongodb.NewMongoDBRepository(dbClientTest.Database(databaseName))
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
				assert.Equalf(t, tc.user.LastName, tc.user.LastName, "Expected LastName to be %s, but was %s", tc.user.LastName, tc.user.LastName)
				assert.Equalf(t, tc.user.Nickname, tc.user.Nickname, "Expected Nickname to be %s, but was %s", tc.user.Nickname, tc.user.Nickname)
				assert.Equalf(t, tc.user.Password, tc.user.Password, "Expected Password to be %s, but was %s", tc.user.Password, tc.user.Password)
				assert.Equalf(t, tc.user.Email, tc.user.Email, "Expected Email to be %s, but was %s", tc.user.Email, tc.user.Email)
				assert.Equalf(t, tc.user.Country, tc.user.Country, "Expected Country to be %s, but was %s", tc.user.Country, tc.user.Country)
				assert.Truef(t, res.CreatedAt.After(now), "Expecrted CreatedAt to be after %s, but was %s", now, res.CreatedAt)
				assert.Truef(t, res.UpdatedAt.After(now), "Expecrted UpdatedAt to be after %s, but was %s", now, res.UpdatedAt)

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
		id             uuid.UUID
		expectedResult *models.User
		expectedError  error
	}{
		{
			"Get user by ID successfully",
			uuid.MustParse("29621CF9-C989-4266-A5A2-085FD99A75E1"),
			&models.User{
				ID:        uuid.MustParse("29621CF9-C989-4266-A5A2-085FD99A75E1"),
				FirstName: "Already inserted user first name 1",
				LastName:  "Already inserted user last name 1",
				Nickname:  "Already inserted user nickname 1",
				Password:  "Already inserted user password 1",
				Email:     "Already inserted user email 1",
				Country:   "Already inserted user country 1",
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
			},
			nil,
		},
		{
			"Get not found user by ID",
			uuid.MustParse("C43DF343-FFB3-43DA-9BB0-B08B81E6FCD9"),
			nil,
			mongo.ErrNoDocuments,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Given
			now := time.Now()
			mongoRepo := mongodb.NewMongoDBRepository(dbClientTest.Database(databaseName))
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
		id            uuid.UUID
		expectedError error
	}{
		{
			"Delete user by id successfully",
			uuid.MustParse("19957751-A789-44D4-BC3B-87390B0E7C0A"),
			nil,
		},
		{
			"Delete user not found with error",
			uuid.New(),
			nil,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Given
			mongoRepository := mongodb.NewMongoDBRepository(dbClientTest.Database(databaseName))
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
