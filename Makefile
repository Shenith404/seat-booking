.PHONY: build run test clean docker-up docker-down sqlc swagger migrate-up migrate-down

# Build the application
build:
	go build -o bin/server.exe ./cmd/api

# Run the application
run:
	go run ./cmd/api

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Start Docker containers
docker-up:
	docker-compose up -d postgres redis

# Stop Docker containers
docker-down:
	docker-compose down

# Generate SQLC code
sqlc:
	docker-compose run --rm sqlc

# Generate Swagger documentation
swagger:
	swag init -g cmd/api/main.go -o docs

# Install swag CLI
install-swag:
	go install github.com/swaggo/swag/cmd/swag@latest

# Run database migrations up
migrate-up:
	goose -dir sql/migrations postgres "$(GOOSE_DBSTRING)" up

# Run database migrations down
migrate-down:
	goose -dir sql/migrations postgres "$(GOOSE_DBSTRING)" down

# Download dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Full setup
setup: docker-up deps migrate-up
	@echo "Setup complete. Run 'make run' to start the server."
