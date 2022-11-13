MAINPATH := ./cmd/server
BINDIR := $(CURDIR)/bin
BINNAME ?= users-microservice

PKG := ./...
LDFLAGS := -w -s
CGO_ENABLED ?= 0
TESTFLAGS := -v -race -failfast -cover -coverprofile=./test/coverage/c.out

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
	go test $(TESTFLAGS) ./...

.PHONY: cover
cover: test
	go tool cover -html=./test/coverage/c.out