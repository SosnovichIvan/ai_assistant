.PHONY: build run test lint migrate docker

# Build the application
build:
	go build -o bin/server ./cmd/server

# Run the application
run:
	go run ./cmd/server

# Run tests
test:
	go test -v -race ./...

# Run tests with coverage
test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	golangci-lint run ./...

# Run database migrations
migrate:
	psql "$$DATABASE_URL" -f migrations/001_init.sql

# Run with Docker
docker-build:
	docker build -t ai_assistant .

docker-run:
	docker run -p 8080:8080 --env-file .env ai_assistant

# Development helpers
dev:
	go run ./cmd/server

# Generate mocks (if needed)
mocks:
	mockgen -source=internal/service/embedding.go -destination=internal/service/mocks/embedding_mock.go

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html
