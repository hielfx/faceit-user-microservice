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
cover: test
	$(GOBIN) tool cover -html=./test/coverage/c.out

.PHONY: coverage
coverage: cover

$(MOCKGEN):
	(cd /; GO111MODULE=on $(GOBIN) install github.com/golang/mock/mockgen@v1.6.0)

.PHONY: generate
generate: $(MOCKGEN)
	$(GOBIN) generate ./...

.PHONY: compose-dev
compose-dev:
	docker compose -f docker-compose.dev.yaml up -d