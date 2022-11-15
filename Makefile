MAINPATH := ./cmd/server
BINDIR := $(CURDIR)/bin
BINNAME ?= users-microservice

PKG := ./...
LDFLAGS := -w -s
CGO_ENABLED ?= 0
TEST_FLAGS := -v -race -failfast -cover -coverprofile=./test/coverage/c.out
TEST_EXTRA_FLAGS ?= 

.PHONY: all
all: build

.PHONY: build
build: $(BINDIR)/$(BINNAME)

$(BINDIR)/$(BINNAME):
	GO111MODULE=on CGO_ENABLED=$(CGO_ENABLED) go build -trimpath -ldflags '$(LDFLAGS)' -o '$(BINDIR)'/$(BINNAME) $(MAINPATH)

.PHONY: run
run:
	go run $(MAIN_PATH)

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: test
test:
	go test $(TEST_FLAGS) $(TEST_EXTRA_FLAGS) ./...

.PHONY: cover
cover: test
	go tool cover -html=./test/coverage/c.out

$(MOCKGEN):
	(cd /; GO111MODULE=on go install github.com/golang/mock/mockgen@v1.6.0)

.PHONY: generate
generate: $(MOCKGEN)
	go generate ./...