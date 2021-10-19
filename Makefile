GO111MODULE=on

.PHONY: all test test-short build

build:
	export GO111MODULE on; \
	go build ./...

lint: build
	go get -u golang.org/x/lint/golint
	golint -set_exit_status ./...

test-short: lint
	go test ./... -v -covermode=count -coverprofile=coverage.out -short

test: lint
	go test ./... -covermode=count -coverprofile=coverage.out

test-coverage: test
	go tool cover -html=coverage.out

