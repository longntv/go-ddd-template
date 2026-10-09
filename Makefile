.PHONY: run/server run/subscriber test test/unit test/integration test/all lint wire/build docker/up docker/down docker/build db/migrate/up db/migrate/down db/migrate/create

# Application
APP_NAME := go-ddd-template
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Go
GO := go
GO_FLAGS := -v
WIRE := $(GO) run github.com/google/wire/cmd/wire@v0.6.0

# Docker
DOCKER_COMPOSE := docker compose -f docker/docker-compose.yaml

# Database
MIGRATE := migrate -path database/migrations -database "postgresql://$$DB_USER:$$DB_PASSWORD@$$DB_HOST:$$DB_PORT/$$DB_NAME?sslmode=$$DB_SSLMODE"

# Server targets
run/server:
	$(GO) run $(GO_FLAGS) $(LDFLAGS) ./cmd/server

run/subscriber:
	$(GO) run $(GO_FLAGS) $(LDFLAGS) ./cmd/subscriber

# Build
build/server:
	$(GO) build $(GO_FLAGS) $(LDFLAGS) -o bin/server ./cmd/server

build/subscriber:
	$(GO) build $(GO_FLAGS) $(LDFLAGS) -o bin/subscriber ./cmd/subscriber

# Wire dependency injection
wire/build:
	$(WIRE) ./internal/registry
	$(WIRE) ./test/integration/registry

# Test
# Unit tests need nothing running. Integration tests (build tag
# "integration") need Postgres: `make docker/up`, or point DB_HOST/DB_PORT/
# DB_USER/DB_PASSWORD at any Postgres 13+ where the user may CREATE DATABASE.
test: test/unit

test/unit:
	$(GO) test -race ./...

test/integration:
	$(GO) test -race -count=1 -tags=integration ./internal/infrastructure/datastore/... ./test/integration/...

test/all:
	$(GO) test -race -count=1 -tags=integration ./...

test/coverage:
	$(GO) test -race -count=1 -tags=integration -coverpkg=./internal/... -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# Lint
lint:
	golangci-lint run ./...

# Docker
docker/up:
	$(DOCKER_COMPOSE) up -d

docker/down:
	$(DOCKER_COMPOSE) down

docker/logs:
	$(DOCKER_COMPOSE) logs -f

docker/build:
	docker build -t $(APP_NAME):$(VERSION) -f docker/Dockerfile .

# Database (golang-migrate)
db/migrate/up:
	$(MIGRATE) up

db/migrate/down:
	$(MIGRATE) down 1

db/migrate/version:
	$(MIGRATE) version

db/migrate/create:
	@read -p "Enter migration name: " name; \
	timestamp=$$(date +%Y%m%d%H%M%S); \
	mkdir -p database/migrations; \
	touch database/migrations/$${timestamp}_$${name}.up.sql; \
	touch database/migrations/$${timestamp}_$${name}.down.sql; \
	echo "Created migration files: $${timestamp}_$${name}.up.sql and $${timestamp}_$${name}.down.sql"

# Clean
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install tools
tools/install:
	$(GO) install github.com/google/wire/cmd/wire@v0.6.0
	$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.3.0
	$(GO) install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1

.PHONY: generate
generate:
	$(GO) generate ./...
	$(GO) fmt ./...
	$(GO) mod tidy