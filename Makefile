MAINPATH := ./cmd/server
BINDIR := $(CURDIR)/bin
BINNAME ?= users-microservice

GOBIN ?= $(shell which go)

PKG := ./...
LDFLAGS := -w -s
CGO_ENABLED ?= 0
TEST_FLAGS := -race -failfast -cover -coverprofile=./test/coverage/c.out
EXTRA_TEST_FLAGS ?= -v

.PHONY: all
all: build

.PHONY: build
build: $(BINDIR)/$(BINNAME)

$(BINDIR)/$(BINNAME):
	GO111MODULE=on CGO_ENABLED=$(CGO_ENABLED) $(GOBIN) build -trimpath -ldflags '$(LDFLAGS)' -o $(BINDIR)/$(BINNAME) $(MAINPATH)

.PHONY: run
run:
	$(GOBIN) run $(MAINPATH)

.PHONY: tidy
tidy:
	$(GOBIN) mod tidy

.PHONY: test
test:
	$(GOBIN) test $(TEST_FLAGS) $(EXTRA_TEST_FLAGS) ./...

.PHONY: cover
cover: test cover-only

.PHONY: coverage
coverage: cover

.PHONY: coverage-only
coverage-only:
	$(GOBIN) tool cover -html=./test/coverage/c.out

.PHONY: cover-only
cover-only: coverage-only

$(MOCKGEN):
	(cd /; GO111MODULE=on $(GOBIN) install github.com/golang/mock/mockgen@v1.6.0)

.PHONY: generate
generate: $(MOCKGEN)
	$(GOBIN) generate ./...

.PHONY: compose-up-dev
compose-up-dev:
	docker compose -f docker-compose.dev.yaml up --remove-orphans -d

.PHONY: compose-dev
compose-dev: compose-up-dev

.PHONY: compose-down-dev
compose-down-dev:
	docker compose -f docker-compose.dev.yaml down
.PHONY: compose-ps-dev
compose-ps-dev:
	docker compose -f docker-compose.dev.yaml ps

$(SWAGGOSWAG):
	(cd /; GO111MODULE=on $(GOBIN) install github.com/swaggo/swag/cmd/swag@latest)

.PHONY: swag
swag: $(SWAGGOSWAG)
	@echo "Running swaggo-swag"
	swag init -g **/**/*.go
	swag fmt
