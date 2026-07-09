.PHONY: build test lint docker-build docker-run clean tidy generate help

## help: Show this help message
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

## build: Compile the API binary
build:
	go build -ldflags="-w -s" -o bin/api ./cmd/api/

## test: Run all tests with race detection and coverage
test:
	go test ./... -race -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out

## lint: Run golangci-lint
lint:
	@if command -v golangci-lint > /dev/null 2>&1; then \
		golangci-lint run --timeout=5m ./...; \
	else \
		echo "golangci-lint not installed. Install with: brew install golangci-lint"; \
		go vet ./...; \
	fi

## docker-build: Build the Docker image
docker-build:
	docker build -t go-production-api -f deploy/Dockerfile .

## docker-run: Run the API in Docker
docker-run:
	docker run --rm -p 8080:8080 --env-file .env go-production-api

## tidy: Run go mod tidy
tidy:
	go mod tidy

## generate: Run go generate (creates mocks)
generate:
	go generate ./...

## clean: Remove build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html
