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
			mongoRepo := mongodb.NewMongoDBRepository(dbClientTest.Database("mongo-test"))
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
