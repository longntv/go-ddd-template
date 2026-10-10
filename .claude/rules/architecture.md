---
paths:
  - "internal/**/*.go"
  - "cmd/**/*.go"
  - "test/integration/registry/**/*.go"
---

# Architecture rules

## Dependency direction
`handler → usecase (interfaces) ← service → domain ← infrastructure`

- `internal/domain/**` imports only the standard library, `uuid`, and the CloudEvents SDK types it
  already uses. Never gin, gorm, aws or anything under `internal/infrastructure`.
- `internal/service/**` depends on **ports only**: `domain/gateway` interfaces, `domain/entity`,
  `domain/event`, `domain/model`, `usecase/{input,output}`, plus `*zap.Logger` (injected, never the
  global) for errors it handles instead of returning. Never import `internal/infrastructure/**`
  or `internal/handler/**`. If a service needs something external, add a port in `domain/gateway`
  and bind the implementation with `wire.Bind`.
- `internal/handler/**` talks to use-case interfaces in `internal/usecase`, never to services or
  gateways directly.
- `internal/infrastructure/**` implements ports and returns domain entities. GORM models
  (`*Entity` types) never leave the `datastore` package; convert with `ToDomain()`.

## Naming and file layout (one concept per file)
| Thing | File | Exported API |
|---|---|---|
| Entity | `domain/entity/<agg>.go` | `type <Agg>`, `type <Agg>ID = uuid.UUID`, `New<Agg>(...)`, behaviour methods |
| Ports | `domain/gateway/<agg>.go` | `<Agg>QueriesGateway` (reads), `<Agg>CommandsGateway` (writes) |
| Use case | `usecase/usecase.go` | `type <Verb><Agg> interface { Execute(ctx, *input.<Verb><Agg>) (*output.<Verb><Agg>, error) }` |
| DTOs | `usecase/input/<agg>.go`, `usecase/output/<agg>.go` | struct per use case, `validate:` tags on input |
| Service | `service/<agg>_<verb>er.go` (`user_creator.go`, `user_lister.go`) | `New<Verb><Agg>(ports..., logger *zap.Logger) usecase.<Verb><Agg>` (logger only when it publishes); implementation struct unexported |
| Datastore | `datastore/<agg>_entity.go`, `_reader.go`, `_writer.go` | `New<Agg>Reader(*gorm.DB) gateway.<Agg>QueriesGateway`, `New<Agg>Writer` |
| HTTP | `handler/http/server/<agg>_<verb>.go` (`user_create.go`) | method on a handler struct; routes in `handler/http/router.go` |
| Migration | `database/migrations/NNNNNN_<desc>.up.sql` + `.down.sql` | always both |

## Errors
- Services return `*model.DomainError` via `model.NewDomainError(code, message, cause)`.
  Codes returned today: `USER_EXISTS`, `USER_NOT_FOUND`, `INVALID_INPUT` (input the handler cannot check, e.g. a password over 72 bytes), `INTERNAL`. New aggregates use
  `<AGG>_NOT_FOUND`, `<AGG>_EXISTS` and must be mapped in `handleError`
  (`handler/http/server/handler.go`): not found → 404, exists → 409, invalid → 400, else 500.
- Datastore readers translate `gorm.ErrRecordNotFound` into a domain sentinel (`model.ErrUserNotFound`).
- Compare errors with `errors.Is` / `errors.As`, never `==` or type assertions.
- Never return internal error text to HTTP clients for the `INTERNAL` case.

## Wiring
- Every package with constructors exports `var WireSet = wire.NewSet(...)`.
- Add new providers to the package `WireSet`, then run `make generate`. Never hand-edit `wire_gen.go`.
- `internal/registry/wire.go` builds production graphs; `test/integration/registry/wire.go` builds the
  test graph, takes external dependencies (DB, password hasher, publisher, logger) as **parameters**
  and lists datastore constructors explicitly.
  When a service gains a new dependency, update both injectors.

## Events
- Event type strings are constants in `domain/event` (`com.go-ddd-template.<agg>.<past-tense-verb>`, built as `<Agg>EventTypePrefix + ".<verb>"`); never
  inline the string anywhere else.
- Every type a service publishes must be registered in `ProvideConfiguredMux`
  (`infrastructure/cloudevents/provider.go`) using those constants.
- Publishing goes through `gateway.EventPublisher`.

## Interfaces and mocks
- Port and use-case interface files carry `//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -destination=mock/<file>.go -source=<file>.go`.
  Mocks live in a `mock/` subpackage (`mock_gateway`, `mock_usecase`). Regenerate, never edit.
