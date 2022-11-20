MAINPATH := ./cmd/server
BINDIR := $(CURDIR)/bin
BINNAME ?= users-microservice

GOBIN ?= $(shell which go)

PKG := ./...
LDFLAGS := -w -s
CGO_ENABLED ?= 0
TEST_FLAGS := -race -failfast -cover -coverprofile=./test/coverage/c.out
EXTRA_TEST_FLAGS ?= -v

DC_DEV_FILE ?= docker-compose.dev.yaml
DC_LOCAL_FILE ?= docker-compose.local.yaml
DC_ARGS ?= ""

CONFIG_FILE_LOCAL := $(CURDIR)/config/local.yaml
CONFIG_FILE_DEV := $(CURDIR)/config/dev.yaml

.PHONY: all
all: build

.PHONY: build
build:
	GO111MODULE=on CGO_ENABLED=$(CGO_ENABLED) $(GOBIN) build -trimpath -ldflags '$(LDFLAGS)' -o $(BINDIR)/$(BINNAME) $(MAINPATH)

.PHONY: run
run:
ifeq ($(CONFIG_FILE),)
run: CONFIG_FILE=$(CONFIG_FILE_LOCAL)
endif
run:
	CONFIG_FILE=$(CONFIG_FILE) $(GOBIN) run $(MAINPATH)

.PHONY: tidy
tidy:
	$(GOBIN) mod tidy

.PHONY: vendor
vendor:
	$(GOBIN) mod vendor

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

.PHONY: dev-up
dev-up:
	docker compose -f $(DC_DEV_FILE) up --remove-orphans --build -d

.PHONY: dev-down
dev-down:
	docker compose -f $(DC_DEV_FILE) down --remove-orphans

.PHONY: dev-exec
dev-exec:
	docker compose -f $(DC_DEV_FILE) $(DC_ARGS)
	
.PHONY: local-up
local-up:
	docker compose -f $(DC_LOCAL_FILE) up --remove-orphans --build -d

.PHONY: local-down
local-down:
	docker compose -f $(DC_LOCAL_FILE) down --remove-orphans

.PHONY: local-exec
local-exec:
	docker compose -f $(DC_LOCAL_FILE) $(DC_ARGS)

$(SWAGGOSWAG):
	(cd /; GO111MODULE=on $(GOBIN) install github.com/swaggo/swag/cmd/swag@latest)

.PHONY: swag
swag: $(SWAGGOSWAG) vendor
	@echo "Running swaggo-swag"
	swag init --parseDependency --parseVendor --parseInternal  -g **/**/*.go --exclude ./vendor
	swag fmt

.PHONY: subscriber
ifeq ($(CONFIG_FILE),)
subscriber: CONFIG_FILE=$(CONFIG_FILE_LOCAL)
endif
subscriber:
	CONFIG_FILE=$(CONFIG_FILE) $(GOBIN) run ./cmd/subscriber

.PHONY: build-subscriber
build-subscriber:
	GO111MODULE=on CGO_ENABLED=$(CGO_ENABLED) $(GOBIN) build -trimpath -ldflags '$(LDFLAGS)' -o $(BINDIR)/subscriber ./cmd/subscriber