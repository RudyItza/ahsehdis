# Makefile for ahsehdis project

# Variables
APP_NAME = ahsehdis
DB_NAME = ahsehdis
DB_USER = ahsehdis
DB_PASSWORD = rudy
MIGRATION_PATH = migrations
DB_DSN = postgres://$(DB_USER):$(DB_PASSWORD)@localhost/$(DB_NAME)?sslmode=disable
CSRF_KEY = "32-byte-long-csrf-key-here!"  # Generate with: openssl rand -base64 32
SESSION_SECRET = "your-32-byte-long-secret-key-here!"  # Generate with: openssl rand -base64 32

.PHONY: run migrateup resetdb


migrateup:
	@echo "Applying migrations..."
	goose -dir $(MIGRATION_PATH) postgres "$(DB_DSN)" up

resetdb:
	@echo "Resetting database..."
	-dropdb -U $(DB_USER) --if-exists $(DB_NAME)
	createdb -U $(DB_USER) $(DB_NAME)
	$(MAKE) migrateup

# ... rest of the Makefile remains the same ...

run:
	@echo "Starting application..."
	CSRF_KEY=$(CSRF_KEY) \
	SESSION_SECRET=$(SESSION_SECRET) \
	DB_DSN=$(DB_DSN) \
	go run cmd/web/main.go

test:
	@echo "Running tests..."
	go test -v -cover ./...

lint:
	@echo "Linting code..."
	golangci-lint run

clean:
	@echo "Cleaning up..."
	go clean
	rm -f coverage.out

# Database management shortcuts
createdb:
	createdb $(DB_NAME)

dropdb:
	dropdb $(DB_NAME)

migrate-create:
	@read -p "Enter migration name: " name; \
	goose -dir $(MIGRATION_PATH) create $${name} sql

# Helpers
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out