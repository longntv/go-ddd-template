# Go DDD Template

A Domain-Driven Design (DDD) template with CQRS pattern, using HTTP server, AWS services, and CloudEvents.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              HTTP API Server                                 │
├─────────────────────────────────────────────────────────────────────────────┤
│  HTTP Request → Handler → UseCase (Command/Query) → Gateway → DataStore     │
│                              ↓                                               │
│                      Domain Event Published (SNS)                            │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                           SQS Subscriber                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│  SQS Message → CloudEvents Subscriber → Handler → Process Event             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Project Structure

```
.
├── cmd/
│   ├── server/          # HTTP API server
│   └── subscriber/      # SQS message subscriber
├── internal/
│   ├── domain/          # Domain layer (entities, events, gateways)
│   ├── usecase/         # Use case layer (CQRS commands/queries)
│   ├── infrastructure/  # Infrastructure (AWS, DB, CloudEvents)
│   └── handler/         # HTTP & CloudEvent handlers
├── database/            # SQL migrations (golang-migrate)
├── test/integration/    # End-to-end tests + test Wire registry
├── testdata/fixtures/   # YAML rows loaded into the test template DB
└── docker/             # Docker configuration
```

## Getting Started

### Prerequisites

- Go 1.23+
- Docker & Docker Compose
- PostgreSQL 16 (via Docker Compose)

### Local Development

1. Copy environment variables:
```bash
cp .env.example .env
```

2. Start infrastructure:
```bash
make docker/up
```

3. Run database migrations:
```bash
make db/migrate
```

4. Generate Wire code:
```bash
make wire/build
```

5. Run the server:
```bash
make run/server
```

6. Run the subscriber (in another terminal):
```bash
make run/subscriber
```

## Development

| Command | Description |
|---------|-------------|
| `make run/server` | Run HTTP server |
| `make run/subscriber` | Run SQS subscriber |
| `make test` | Run unit tests (no services needed) |
| `make test/integration` | Run integration tests against Postgres |
| `make test/coverage` | Run all tests with coverage |
| `make wire/build` | Generate Wire code |
| `make db/migrate` | Run database migrations |
| `make docker/up` | Start Docker services |
| `make docker/down` | Stop Docker services |

## Testing

### Unit tests (`make test`)

Table-driven tests next to the code, using gomock mocks of the ports
(`internal/domain/gateway/mock`, `internal/usecase/mock`) and `go-cmp` for diffs.

| What | Where | Mocks |
|------|-------|-------|
| Use cases | `internal/service/*_test.go` | gateways + `EventPublisher` |
| HTTP handlers | `internal/handler/http/server/handler_test.go` | use cases, via `httptest` |
| Event routing | `internal/infrastructure/cloudevents/provider_test.go` | none |
| Entities | `internal/domain/entity/*_test.go` | none |

### Integration tests (`make test/integration`)

Files carry `//go:build integration` and need Postgres (`make docker/up`, or set
`DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD`).

- `internal/testutil` creates a **template database** once per test binary: it runs
  `database/migrations/*.up.sql` and loads `testdata/fixtures/<table>.yml`.
- Each test gets its own copy with `CREATE DATABASE ... TEMPLATE`, dropped on cleanup,
  so tests run in parallel without sharing state.
  - Reader tests share one copy (`testutil.InitReadDB` in `TestMain`).
  - Writer tests call `testutil.InitDB(t)` per test case.
- `internal/infrastructure/datastore/*_test.go` exercises the GORM readers/writers.
- `test/integration/http` drives the real router, services and datastore through
  `test/integration/registry` (a Wire injector that takes the DB and a mock
  `EventPublisher` as parameters, so no AWS is needed).

To add a table: add its migration, a `testdata/fixtures/<table>.yml`, then
reader/writer tests following `user_reader_test.go` / `user_writer_test.go`.

## DDD Layers

### Domain Layer (`internal/domain/`)
- **Entity**: Core business entities
- **Event**: Domain events
- **Gateway**: Interfaces for external dependencies

### UseCase Layer (`internal/usecase/`)
- **Input**: Command/Query DTOs
- **Output**: Response DTOs
- **Use Case**: Business logic implementation (CQRS)

### Infrastructure Layer (`internal/infrastructure/`)
- **AWS**: Bedrock, S3, SNS, SQS clients
- **DataStore**: GORM with read/write splitting
- **CloudEvents**: Event pub/sub with CNCF spec

### Handler Layer (`internal/handler/`)
- **HTTP**: REST API handlers with Gin
- **CloudEvents**: Event processing handlers
