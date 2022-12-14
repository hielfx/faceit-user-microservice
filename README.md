<!-- omit in toc -->
# Users microservice

<!-- omit in toc -->
## Table of content

- [Simple Architecture Diagram](#simple-architecture-diagram)
- [Project folder structure](#project-folder-structure)
- [Prerequisites](#prerequisites)
- [Building the project binary](#building-the-project-binary)
- [Running the project](#running-the-project)
  - [Running in "local" mode](#running-in-local-mode)
  - [Running in "development" mode](#running-in-development-mode)
  - [Accesing the routes](#accesing-the-routes)
- [Configuring the project](#configuring-the-project)
- [Testing the project](#testing-the-project)
- [Generating a new Swagger documentation (update the documentation)](#generating-a-new-swagger-documentation-update-the-documentation)
- [Generating new repository mocks](#generating-new-repository-mocks)
- [Asumptions, Desisions and Things to change/improve](#asumptions-desisions-and-things-to-changeimprove)
- [Possible way to deploy to production](#possible-way-to-deploy-to-production)

## Simple Architecture Diagram

```ascii
Response▲ │Request
        │ │
  ┌─────┼─┼──────────────────────────┐
  │     │ │                          │  Server:     Main application (the actual API).
  │   ┌─┴─▼──┐ Publish ┌───────┐     │
  │   │Server├─────────► Redis │     │  MongoDB:    Storage database.
  │   └─▲──┬─┘         └──▲────┘     │
  │     │  │              │          │  Redis:      Pub-sub database.
  │ Read│  │Insert/       │Subscribe │
  │     │  │Update        │          │  Subscriber: Just prints the payloads that
  │     │  │              │          │              receives from the subscribed
  │   ┌─┴──▼────┐     ┌───┴──────┐   │              topics (used to show a working
  │   │ MongoDB │     │Subscriber│   │              pub-sub environment).
  │   └─────────┘     └──────────┘   │              Also refered to as "sidecar".
  │                                  │
  └──────────────────────────────────┘
```

## Project folder structure

```ascii
.
├── Dockerfile                      # Dockerfile for the main app (./cmd/server)
├── Makefile
├── README.md
├── cmd
│   ├── server
│   │   └── main.go                 # Main application (the actual server)
│   └── subscriber
│       └── main.go                 # Sidecar application to check the pub-sub flows (this is a subscriber/listener)
├── config
│   ├── config.go                   # Configuration reader and parser
│   ├── dev.yaml                    # Configuration for development environment (docker-compose with applications)
│   └── local.yaml                  # Configuration for local development (docker-compose with db only)
├── docker
│   ├── dev
│   │   ├── Dockerfile.subscriber   # Dockerfile for the subscriber sidecar
│   │   └── init-db.js              # Initialize mongodb with some data for development environment
│   └── local
│       └── init-db.js              # Initialize mongodb with some data for local environment
├── docker-compose.dev.yaml         # Docker compose file for development environment
├── docker-compose.local.yaml       # Docker compose file for local environment
├── docs                            # Generated Swagger API documentation (`make swag`)
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── go.mod
├── go.sum
├── internal                        # Project internal files (main application code lies here)
│   ├── errors
│   │   └── http
│   │       └── errors.go           # HTTP shared errors
│   ├── models                      # Domain/model layer
│   │   └── user.go                 # User data
│   ├── pagination                  # Pagination package
│   │   ├── pagination.go
│   │   ├── pagination_test.go
│   │   ├── sortOrder.go            # UNUSED
│   │   └── sortOrder_test.go
│   ├── server
│   │   └── server.go               # Main application code (the server)
│   ├── testutils                   # Utilities for testing purposes
│   │   ├── dateCheck.go
│   │   ├── errors.go
│   │   ├── testutils.go
│   │   └── users.go
│   └── users                       # Users package
│       ├── handlers.go             # User handler (http methods) interface
│       ├── http                    # User handlers implementation
│       │   ├── handlers.go
│       │   ├── handlers_test.go
│       │   └── routes.go
│       ├── mock                    # User interfaces mock (generated with `make generate`)
│       │   ├── handlers_mock.go    # Mocked handlers
│       │   └── repository_mock.go  # Mocked repository
│       ├── pubsub                  #
│       │   ├── pubsub.go           # Pubsub interface
│       │   ├── redis.go            # Redis pubsub implementation
│       │   └── topics.go           # Subscription topics
│       ├── repository              # User repository implementation
│       │   └── mongodb             
│       │       ├── init_db.js
│       │       ├── mongodb.go      # Mongodb repository implementation
│       │       └── mongodb_test.go
│       ├── repository.go           # User repository interface
│       └── sec
│           └── password.go         # Security utility for passwords (UNUSED)
├── pkg                             # External packages with no internal dependencies
│   └── db
│       ├── mongodb                 # Mongodb database access/connection implementation
│       │   ├── mongo_registry.go
│       │   └── mongodb.go
│       └── redis                   # Redis access/connection implementation          
│           └── redis.go
└── test
    └── coverage                    # Test coverage output folder
```

## Prerequisites

- **Docker** (used `Docker version 20.10.21, build baeda1f`)
- **Docker compose** (used `Docker Compose version v2.12.2`)
- **Make** (used `GNU Make 3.81`)
- **Go** (used `go version go1.19.2 darwin/amd64`)

## Building the project binary

In order to build the project binary, execute the following command:

```sh
make # make build also works.
```

The binary will be created in `./bin/users-microservice` and can be run with

```sh
CONFIG_FILE=config_file_location.yaml ./bin/users-microservice
```

There's a sidecar for testing the integration. It can be built with

```sh
make build-subscriber
```

And run with

```sh
CONFIG_FILE=config_file_location.yaml ./bin/subscriber
```

The `CONFIG_FILE` variable is needed for the project to run, also both redis and mongodb database already up and running.

You can run the project with an easier way in the following section.

## Running the project

For the project to run, it's necessary to have an already running mongodb instance and a redis instance.

For ease of use, there are 2 ways of running this project: local and development mode.

The main difference is development mode builds the project, along with a subscriber sidecar and the databases; while the local mode just starts the mongodb and redis database, using both docker-compose

### Running in "local" mode

Local mode means "database only": the docker-compose will only start the databases, the server and sidecar should be run separately. This is for the developer convenience.

To run the project in "local" mode, first the databses should be already up. The following commands will start the databases and the project:

```sh
make local-up # starts mongodb and redis
make run # starts the application server 
```

In order to run the subscriber/listener sidecar, run the following command in a separate terminal:

```sh
make subscriber
```

The `make local-up` command will create a `volumes` folder in `./docker/local/volumes` which contains both volumes for mongodb and redis.

To stop the databases:

```sh
make local-stop
```

To execute docker-compose down run:

```sh
make local-down
```

*NOTE: It will NOT delete the volumes folder, it has to be done manually if required*

### Running in "development" mode

Development mode will start the databases, along with the server and the sidecar. "This is like a deployment, but in the local machine".

To run the project in "development" mode, run the following command:

```sh
make dev-up
```

It will create a `volumes` folder in `./docker/dev/volumes` which contains both volumes for mongodb and redis.

To stop the project:

```sh
make dev-stop
```

To execute docker-compose down run:

```sh
make dev-down
```

*NOTE: It will NOT delete the volumes folder, it has to be done manually if required*

### Accesing the routes

By default, the project runs on `http://localhost:4040`, it could be modified in the [configuration section](#configuring-the-project). The project routes are the following ones:

- `GET /api/v1/health` -> Simple health check, just display a status message (we could thisplay some server info here, but it's not implemented)
  
- `GET /api/v1/swagger/index.html` -> Swagger documentation (the API documentation)

- `GET /api/v1/users` -> Gets the paginated users
- `GET /api/v1/users/:userId` -> Gets the user by its id
- `POST /api/v1/users` -> Creates a new user
- `POST /api/v1/users/:userId` -> Updates the user by its id
- `DELETE /api/v1/users/:userId` -> Deletes the user by its id

## Configuring the project

The project needs a `CONFIG_FILE` environment variable for it to run. This environment variable must have the path to a configuration yaml file (the file must exists).

Currently, there are 2 configuration files: one for "local mode" and other one for "development mode" inside the folder `config`, located at the root of the project. The main difference is the databases addresses, one using `docker dns` and other using `localhost`.

## Testing the project

To test the project, run the following command:

```sh
make test
```

*NOTE: Docker is necessary to run the repository tests, because it will start a new mongodb docker container for the mongodb repository testing*

For the test coverage, run the following command:

```sh
make coverage # `make cover` also works
```

This command will retrigger the `make test` command and the execute the coverage, so it's not necessary to execute the test first. In case you wanted to execute the coverage only, you can do it with the following command:

```sh
make coverage-only # `make cover-only` also works
```

There's a variable called `EXTRA_TEST_FLAGS` that, by default, as `-v` value. In order to deactivate the `verbosity`, you can run the test as following (you can use this to pass extra flags, but the current `TEST_FLAGS` cannot be overwritten):

```sh
EXTRA_TEST_FLAGS= make test # disable verbose mode
EXTRA_TEST_FLAGS="-run ^TestGetUserById" make test # just runs the matching test
```

## Generating a new Swagger documentation (update the documentation)

In order to update the Swagger documentation, it's necessary to run the following command

```sh
make swag
```

This command will read the project files and generate the documentation based on the [swaggo-swag declarative comments](https://github.com/swaggo/swag#declarative-comments-format).

The url can be accesible (once the server is up and running) at `/api/v1/swagger/index.html`

## Generating new repository mocks

The repository mocks have been generated using [gomock and mockgen](https://github.com/golang/mock) and the generated files should not be modified. To generate the mocks again, execute the following command:

```sh
make generate
```

## Asumptions, Desisions and Things to change/improve

This section contains the asumptions, desitions made during the development and things to change/improve in no particular order, just as they came to mi mind. It would've been better for us if I organized this section a little bit but Iit came ut like this.

- This project uses [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/)
- User password will be clear text for simplicity: no bcrypt, no hashing, no hiding in JSON responses, etc. This is for the same reasoning the login is not provided, but it should be handled.
- There should be more edge cases when testing, and I would've liked to do integration testing for the whole flow (making a complete request flow).
- Despite the text saying we must use "id", I used "_id". There are some workarounds that could be done but for simplicity for this challenge, I didn't do it. Some workarounds:
  - Switching to MySQL/PostgreSQL
  - Adding another field called "id", making it unique and "forgetting" about "_id" (also creating and index for search)
  - Create 2 fields (`_id` and `id`) and sync them with the same value
- A "use case" layer could've been a good addition, in which we test and execute the business logic for each use case. In this project this layer does not exists, all the logic has been done in the handlers layer. This has some side effects, such as:
  - We cannot correctly test the handler logic and the use case logic without changing the same test
  - We cannot reuse the logic in other parts of the application if needed.
  - If the use case layer has a new dependency, we have to modify the handlers instead. For example, the redis dependency; this dependency forced us to include it in the handlers instead on its corresponding layer.
- In the beggining I used mongo `ObjectId` (`primitive.ObjectId`) for `_id` and string for `id` but I switched it to Google's `UUID` package for `_id` and dropped the `id` field. In the end, I used a regular string as the `_id` for simplicity.
- There should be a gracefuly shutdown flow to stop the server, but it's not imlpemented yet.
- It should be good to inject some values in build time, such as the git tag, architecture, os, etc. to the binary, providing a way to print it and check it, but it wasn't implemented.
- MongoDB was selected instead of MySQL to use a different database than the one I usually use. This derived in some troubles with the use of the `_id` and the new `mongo-go` driver (`mgo.v2` is now unmaintained so I decided to use the official one). This driver is not so compatible with Google's UUID package and it was being stored as a binary. To solve this, I used a mongodb repository, that converts Google's `UUID` into MongoDB `ObjectId`. This came with it's own caveats such as the FindOne and the Find method because the documents weren't matching, resulting in a nil document or an empty slice. To solve this I used the string I meantioned earlier.
- Currently, you can only filter by the exact string match, it should be case insensitive, but it's not been implemented yet.
- For the API documentation I used Swagger ([`swaggo/swag`](https://github.com/swaggo/swag)) so the documentation could be generated with comments in the code. Maybe it's a good idea to have a separate document with more information, but I went this way so I could learn more about Swagger and OpenAPI.
- In the swagger documentation, for simplicity a whole `models.User` has been used, "requiring uncesserary fields".
- Currently, the binary only builds for the current system. I don't see this as a flaw per se, because, in the end, it will be run and built inside a docker container.
- [Golangci-lint](https://github.com/golangci/golangci-lint) and [SonarQube](https://www.sonarqube.org/) could be used to improve the code quality and security.
- There may be more things, but I can't recall now.
  
## Possible way to deploy to production

Currently it hasn't been provided a way to deploy the application to production, so I will focus on "how I would do it":

1. You can still use the Dockerfile to build the application docker image.
2. The image could be pushed to AWS ECR, so it could be accesible later.
3. The application could be deployed to a Kubernetes cluster. For that we would need at least:
   1. A namespace
   2. A config map to store the config file
   3. A Deployment specification with the config file as volume and the CONFIG_FILE environment variable.
   4. A service specification to expose the deployment
   5. A ingress specification for us the access the application with a domain (we would need a ingress controller. One option could be the [INGRESS-NGINX controller](https://kubernetes.github.io/ingress-nginx/)).
   6. If we want to generate the domain certificates, we could use [cert-manager with NGINX-ingress](https://cert-manager.io/docs/tutorials/acme/nginx-ingress/)
   7. We would also need access to a MongoDB and a Redis database, already configured in the config file.
4. In order to deploy all of this, we could us [Github Actions](https://docs.github.com/en/actions) or something similar (e.g. Jenkins) to automate the docker build, push and K8s deployment every time we push a new git tag, with a full CI/CD cyle.

The Kubernetes cluster could be easily deployed to AWS using [KOPS](https://kops.sigs.k8s.io/) or deploying a smaller cluster with [K3s](https://k3s.io/).
