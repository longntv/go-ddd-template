---
name: ddd-implement
description: Implement features in this Go DDD/CQRS template — a new aggregate, use case, HTTP endpoint, domain event or event handler — following its layer, naming, Wire and error conventions, with tests. Use when asked to add or change business functionality in a go-ddd-template based service.
---

# Implement a feature in go-ddd-template

The `User` aggregate is the reference for every step. Before writing a file, open its `user_*`
counterpart and mirror it. Rules that always apply are in `.claude/rules/architecture.md`.

## 0. Decide the scope

| Request | Steps to do |
|---|---|
| New aggregate (e.g. `Order`) with CRUD | 1 → 9 |
| New use case on an existing aggregate | 2 (if new port methods), 4, 5, 6 (if new queries), 7, 9 |
| New endpoint for an existing use case | 7, 9 |
| New domain event + subscriber handler | 3, 8, 9 |

Work inside-out (domain first, HTTP last) so each step compiles against the previous one.
Run `go build ./...` after each step.

## 1. Entity — `internal/domain/entity/<agg>.go`
Mirror `entity/user.go`:
```go
type OrderID = uuid.UUID

type Order struct {
	ID        OrderID
	// fields...
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewOrder(/* required fields */) *Order { now := time.Now(); return &Order{ID: uuid.New(), /*...*/ CreatedAt: now, UpdatedAt: now} }
```
Put invariants and state changes in methods on the entity, not in services.

## 2. Ports — `internal/domain/gateway/<agg>.go`
Mirror `gateway/user.go`: split reads and writes, keep the `//go:generate` line.
```go
//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination=mock/order.go -source=order.go

type OrderQueriesGateway interface {
	Get(ctx context.Context, id entity.OrderID) (*entity.Order, error)
	List(ctx context.Context, limit, offset int) ([]*entity.Order, int, error)
}

type OrderCommandsGateway interface {
	Create(ctx context.Context, o *entity.Order) error
	Update(ctx context.Context, o *entity.Order) error
	Delete(ctx context.Context, id entity.OrderID) error
}
```
Add a sentinel `ErrOrderNotFound` in `domain/model/error.go`.

## 3. Events — `internal/domain/event/<agg>_event.go`
Create `<agg>_event.go` with `<Agg>EventTypePrefix = "com.go-ddd-template.<agg>"`, the event type
constants and `<Agg>EventData`, following `user_event.go`:
```go
const (
	OrderEventTypePrefix = "com.go-ddd-template.order"
	OrderCreatedEvent    = OrderEventTypePrefix + ".created"
)
```
Add an `<Agg>EventData` payload struct (JSON tags, no secrets). Events themselves are the shared
`event.DomainEvent` (`domain_event.go`), so the publisher port needs no change:
```go
evt := event.NewDomainEvent(event.OrderCreatedEvent, event.Source, order.ID.String(), &event.OrderEventData{...})
```

## 4. Use case contract — `internal/usecase`
- Add the interface to `usecase/usecase.go` (one `Execute` method), then `make generate` for mocks.
- Add `input.<Verb><Agg>` with `validate:` tags and `output.<Verb><Agg>` in `usecase/input|output/<agg>.go`.

## 5. Service — `internal/service/<agg>_<verb>er.go`
Mirror `service/user_creator.go`:
- Constructor takes **ports only** (plus `*zap.Logger` last when it publishes) and returns the
  use-case interface; the struct is unexported.
- Map every failure to `model.NewDomainError(code, message, cause)`:
  not found → `<AGG>_NOT_FOUND`, conflict → `<AGG>_EXISTS`, everything else → `INTERNAL`.
- Check existence with the queries port, mutate through the entity, persist with the commands port,
  re-read if the response needs DB-generated fields, then publish the event with
  `publishBestEffort(ctx, s.eventPublisher, s.logger, evt)`: the change is saved, so a publish
  failure is logged, not returned.
- Add the constructor to `service.WireSet` in `service/wire.go`.

## 6. Datastore — `internal/infrastructure/datastore/`
Mirror the three user files:
- `<agg>_entity.go`: GORM model with `TableName()` and `ToDomain()`.
- `<agg>_reader.go`: `New<Agg>Reader(db *gorm.DB) gateway.<Agg>QueriesGateway`; translate
  `gorm.ErrRecordNotFound` into the domain sentinel with `errors.Is`.
- `<agg>_writer.go`: `New<Agg>Writer(db *gorm.DB) gateway.<Agg>CommandsGateway`; return
  `gorm.ErrRecordNotFound` when `RowsAffected == 0` on update/delete.
- Add both constructors to `datastore.WireSet` **and** list them in `test/integration/registry/wire.go`,
  which names datastore constructors explicitly instead of using `datastore.WireSet` (that set needs config).
- Migration pair `database/migrations/<next number>_create_<table>.up.sql` / `.down.sql`.
  Use `UUID PRIMARY KEY DEFAULT gen_random_uuid()` and `TIMESTAMP(3)` like the users table.

## 7. HTTP — `internal/handler/http/`
- The user endpoints hang off `server.Handler` (`server/handler.go`, which also holds the shared
  `handleError`). For a new aggregate, add a separate struct in `server/<agg>_handler.go`
  (`type OrderHandler struct` + `NewOrderHandler(useCases...)`), add it to `server.WireSet`, and add
  it as a parameter of `Router` in `router.go`. Keep `handleError` shared.
- One file per endpoint, `server/<agg>_<verb>.go`. Bind with a local request struct using
  `binding:` tags, build the `input.*`, call `Execute`, respond with explicit `gin.H` fields.
  Never serialize entities directly (it would leak fields such as passwords).
- Route under `/api/v1/<plural>`.
- Add the new error codes to `handleError` in `server/handler.go`.

## 8. Subscriber handler (events consumed)
- Add `Handle<Agg><Verb>(ctx, *event.Event) error` in `internal/handler/cloudevents/handler.go`.
- Register it in `ProvideConfiguredMux` (`infrastructure/cloudevents/provider.go`) with the
  **constant** from `domain/event`, never a string literal.

## 9. Wire, tests, verify
1. `make generate` (Wire + mocks). If a new external dependency was added, also add it as a
   parameter of `test/integration/registry/wire.go` and regenerate.
2. Write tests with the `ddd-testing` skill: service unit tests, handler unit tests, datastore
   reader/writer integration tests, an HTTP integration test, plus fixtures for the new table.
3. Run and fix until green:
   ```bash
   go build ./... && go vet ./... && go vet -tags=integration ./...
   make test
   make test/integration   # needs Postgres: make docker/up (or set DB_PORT)
   ```
4. Report which commands you ran and their results. If integration tests could not run
   (no Postgres), say so explicitly; don't claim they passed.

## Checklist before finishing
- [ ] No service imports `internal/infrastructure` or `internal/handler`.
- [ ] Every new constructor is in a `WireSet`; `wire_gen.go` was regenerated, not edited.
- [ ] Every new error code is mapped in `handleError`.
- [ ] Every published event type is registered in `ProvideConfiguredMux`.
- [ ] Migration has both `up` and `down`.
- [ ] Unit tests and integration tests added; `make test` green.
