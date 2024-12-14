# Variables
BINARY_NAME = github.com/eth-bridging
MIGRATIONS_DIR = db/migrations
DATABASE_URL ?= $(shell grep DATABASE_URL .env | sed 's/^DATABASE_URL=//')

# Default target
.PHONY: all
all: run

## ------------------------------
## Setup and Dependencies
## ------------------------------

# Install dependencies
.PHONY: setup
setup:
	@echo "Installing dependencies..."
	go mod tidy

## ------------------------------
## Database Migrations
## ------------------------------

# Run database migrations up
.PHONY: migrate-up
migrate-up:
ifndef DATABASE_URL
	$(error DATABASE_URL is not set. Use 'DATABASE_URL=your_url make migrate-up' or set it in the .env file)
endif
	@echo "Applying database migrations..."
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

# Run database migrations down
.PHONY: migrate-down
migrate-down:
ifndef DATABASE_URL
	$(error DATABASE_URL is not set. Use 'DATABASE_URL=your_url make migrate-down' or set it in the .env file)
endif
	@echo "Reverting database migrations..."
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down

# Create a new migration file
.PHONY: new-migration
new-migration:
ifndef NAME
	$(error NAME variable is required. Usage: make new-migration NAME=your_migration_name)
endif
	@echo "Creating new migration: $(NAME)..."
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)

## ------------------------------
## Running the Server
## ------------------------------

# Build the Go application
.PHONY: build
build:
	@echo "Building the application..."
	go build -o $(BINARY_NAME) ./cmd/main.go

# Run the Go application
.PHONY: run
run:
	@echo "Running the application..."
	go run ./cmd/main.go

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning up build artifacts..."
	rm -f $(BINARY_NAME)

## ------------------------------
## Testing
## ------------------------------

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test ./... -cover

## ------------------------------
## Help
## ------------------------------

# Display help
.PHONY: help
help:
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'