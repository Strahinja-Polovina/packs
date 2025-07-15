.PHONY: help build run test clean migrate-up migrate-down migrate-status docker-build docker-run docker-compose-up docker-compose-down swagger templ-generate up

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build the application
build: ## Build the application
	@echo "Building application..."
	go build -o bin/packs cmd/api/main.go

# Generate swagger documentation
swagger: ## Generate swagger documentation
	@echo "Generating swagger documentation..."
	swag init -g cmd/api/main.go

# Generate templ templates
templ-generate: ## Generate templ templates
	@echo "Generating templ templates..."
	templ generate

# Run the application
run: ## Run the application
	@echo "Running application..."
	go run cmd/api/main.go

# Run tests
test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Database migration commands
migrate-up: ## Run database migrations up
	@echo "Running migrations up..."
	goose -dir migrations postgres "$(DB_CONNECTION_STRING)" up

migrate-down: ## Run database migrations down
	@echo "Running migrations down..."
	goose -dir migrations postgres "$(DB_CONNECTION_STRING)" down

migrate-status: ## Check migration status
	@echo "Checking migration status..."
	goose -dir migrations postgres "$(DB_CONNECTION_STRING)" status

migrate-reset: ## Reset database (down all migrations then up)
	@echo "Resetting database..."
	goose -dir migrations postgres "$(DB_CONNECTION_STRING)" reset

# Create a new migration
migrate-create: ## Create a new migration (usage: make migrate-create NAME=migration_name)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	@echo "Creating migration: $(NAME)"
	goose -dir migrations create $(NAME) sql

# Docker commands
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t packs-service .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env packs-service

# Docker Compose commands
docker-compose-up: ## Start services with docker-compose
	@echo "Starting services with docker-compose..."
	docker-compose up -d

docker-compose-down: ## Stop services with docker-compose
	@echo "Stopping services with docker-compose..."
	docker-compose down

docker-compose-logs: ## View docker-compose logs
	@echo "Viewing docker-compose logs..."
	docker-compose logs -f

# Complete development setup and start all services
up: ## Generate templates, swagger, prepare everything, and start all services with docker-compose
	@echo "Preparing application for deployment..."
	@echo "1. Tidying Go modules..."
	go mod tidy
	@echo "2. Generating templ templates..."
	templ generate
	@echo "3. Generating swagger documentation..."
	swag init -g cmd/api/main.go
	@echo "4. Starting all services with docker-compose..."
	docker compose up --build -d
	@echo "✅ All services started successfully!"
	@echo "📖 API documentation: http://localhost:8080/swagger/index.html"
	@echo "🌐 Frontend: http://localhost:8080/"
	@echo "🏥 Health check: http://localhost:8080/health"
	@echo "📊 View logs: make docker-compose-logs"

# Install development dependencies
install-dev-deps: ## Install development dependencies
	@echo "Installing development dependencies..."
	go install github.com/pressly/goose/v3/cmd/goose@latest

# Format code
fmt: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint: ## Lint Go code (requires golangci-lint)
	@echo "Linting code..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Tidy dependencies
tidy: ## Tidy Go modules
	@echo "Tidying dependencies..."
	go mod tidy

# Vendor dependencies
vendor: ## Vendor dependencies
	@echo "Vendoring dependencies..."
	go mod vendor

# Database setup for development
db-setup: ## Setup development database
	@echo "Setting up development database..."
	@echo "Make sure PostgreSQL is running and create database 'packs_db'"
	@echo "Then run: make migrate-up"

# Environment setup
env-setup: ## Setup environment file
	@echo "Setting up environment file..."
	@if [ ! -f .env ]; then \
		echo "Creating .env file..."; \
		echo "DB_HOST=localhost" > .env; \
		echo "DB_PORT=5432" >> .env; \
		echo "DB_USER=postgres" >> .env; \
		echo "DB_PASSWORD=postgres" >> .env; \
		echo "DB_NAME=packs_db" >> .env; \
		echo "DB_SSL_MODE=disable" >> .env; \
		echo "PORT=8080" >> .env; \
		echo ".env file created. Please update with your database credentials."; \
	else \
		echo ".env file already exists."; \
	fi

# Full setup for new developers
setup: env-setup install-dev-deps db-setup ## Full setup for new developers
	@echo "Setup complete! Next steps:"
	@echo "1. Update .env file with your database credentials"
	@echo "2. Create PostgreSQL database 'packs_db'"
	@echo "3. Run 'make migrate-up' to apply migrations"
	@echo "4. Run 'make run' to start the application"

# Default database connection string for local development
DB_CONNECTION_STRING ?= host=localhost port=5432 user=postgres password=postgres dbname=packs_db sslmode=disable
