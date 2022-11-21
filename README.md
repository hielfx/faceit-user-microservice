<!-- omit in toc -->
# Users microservice

<!-- omit in toc -->
## Table of content

- [Project structure](#project-structure)
- [Asumptions, Desitions and Things to change](#asumptions-desitions-and-things-to-change)

## Project structure

```
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

## Asumptions, Desitions and Things to change

- User password will be clear text for simplicity: no bcrypt, no hashing, no hiding in JSON responses, etc. This is for the same reasoning the login is not provided.
- There should be more edge cases when testing
- Despite the text saying we must use "id", I used "_id". There are some workarounds that could be done but for simplicity for this challenge, I didn't do it. Some workarounds:
  - Switching to MySQL
  - Adding another field called "id", making it unique and "forgetting" about "_id" (also creating and index for search)
  - Create 2 fields (_id and id) and sync them with the same value
- There should be a "use case" layer, in which we test and execute the business logic for each use case. In this project this layer does not exists, all the logic has been done in the handlers layer. This has some side effects, such as:
  - We cannot correctly test the handler logic and the use case logic without changing the same test
  - We cannot reuse the logic in other parts of the application if needed.
  - If the use case layer has a new dependency, we have to modify the handlers instead (for example, the redis dependency; this dependency forced us to include it in the handlers instead on its corresponding layer)
  - Initially I used ObjectId() for \_id and string for id but I switched it to google uuid for \_id and dropped the id field. In the end  I used a simple string for the _id for simplicity.
  - I thought of using a configuration file in yaml format and parsing it using viper, but I leave that undone for this version
