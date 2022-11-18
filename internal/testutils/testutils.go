package testutils

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const testingDatabaseName = "test"

// ExecuteTestMain - executes a custom TestMain function
func ExecuteTestMain(m *testing.M, client *mongo.Client) {

	pool, resource, err := SetupDB(client)
	if err != nil {
		logrus.Panic(err)
	}

	// run tests
	code := m.Run()

	if err := Cleanup(pool, resource, client); err != nil {
		logrus.Panic(err)
	}

	os.Exit(code)
}

// SetupDB - helper function that starts a new mongo docker container and set the mongo client
func SetupDB(client *mongo.Client) (*dockertest.Pool, *dockertest.Resource, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		logrus.Fatalf("Could not connect to docker: %s", err)
		return nil, nil, err
	}

	dir, err := os.Getwd()
	if err != nil {
		logrus.Fatalf("Error retrieving working directory: %s", err)
		return nil, nil, err
	}

	logrus.Infof("Current test dir: %s", dir)

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
		logrus.Fatalf("Could start docker container: %s", err)
		return nil, nil, err
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	err = pool.Retry(func() error {
		var err error
		_client, err := mongo.Connect(
			context.TODO(),
			options.Client().ApplyURI(
				fmt.Sprintf("mongodb://localhost:%s", resource.GetPort("27017/tcp")),
			),
		)
		if err != nil {
			return err
		}
		*client = *_client
		return client.Ping(context.TODO(), nil)
	})
	if err != nil {
		logrus.Fatalf("Could not connect to docker: %s", err)
		return nil, nil, err
	}

	return pool, resource, nil
}

func Cleanup(pool *dockertest.Pool, resource *dockertest.Resource, client *mongo.Client) error {
	// When you're done, kill and remove the container
	if err := pool.Purge(resource); err != nil {
		logrus.Fatalf("Could not purge resource: %s", err)
		return err
	}

	// disconnect mongodb client
	if err := client.Disconnect(context.TODO()); err != nil {
		logrus.Errorf("Could not disconnect from db: %s", err)
		return err
	}

	return nil
}

// GetDatabaseFromClient - returns a new database with the testingDatabaseName name for testing purposes
func GetDatabaseFromClient(client *mongo.Client) *mongo.Database {
	return client.Database(testingDatabaseName)
}
