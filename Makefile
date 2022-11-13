
.PHONY: tidy
tidy:
	go mod tidy

.PHONY: test
test:
	go test -v -race -failfast -cover -coverprofile=./test/coverage/c.out ./...

.PHONY: coverage
coverage: test
	go tool cover -html=./test/coverage/c.out