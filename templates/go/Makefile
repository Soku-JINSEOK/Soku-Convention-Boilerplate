.PHONY: build fmt fmt-check lint test tidy

build:
	go build ./...

fmt:
	gofmt -w .
	goimports -w .

fmt-check:
	test -z "$$(gofmt -l .)"
	test -z "$$(goimports -l .)"

lint:
	golangci-lint run ./...

test:
	go test ./...

tidy:
	go mod tidy
