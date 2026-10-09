# go-ddd-template

Go 1.23 service: DDD layers + CQRS ports, Gin HTTP API (`cmd/server`), SQS CloudEvents subscriber
(`cmd/subscriber`), GORM/Postgres, Google Wire DI, gomock + go-cmp tests.

## Commands
- `make test` — unit tests (no services needed). Run after every change.
- `make docker/up` then `make test/integration` — integration tests against Postgres
  (if 5432 is busy: run Postgres elsewhere and set `DB_PORT`).
- `make generate` — Wire injectors + mocks + gofmt. Run after changing any constructor, WireSet or port interface.
- `go vet ./... && go vet -tags=integration ./...` — integration files only compile with the tag.

## Rules and skills
- Layer, naming and wiring rules: `.claude/rules/architecture.md`.
- Test rules: `.claude/rules/testing.md`.
- Adding an aggregate, use case, endpoint or event: use the `ddd-implement` skill.
- Writing unit or integration tests: use the `ddd-testing` skill.

## Canonical examples
The `User` aggregate is the reference implementation for every layer. When unsure how something
should look, copy the matching `user_*` file and its test.
