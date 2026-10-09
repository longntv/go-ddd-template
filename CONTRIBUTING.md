# Contributing

Thanks for your interest in improving go-ddd-template.

## Before you start
- For anything larger than a small fix, open an issue first to agree on the approach.
- Keep pull requests small and focused on one change.

## Development setup
```bash
make tools/install                       # wire, golangci-lint, migrate
make test                                # unit tests
make docker/up && make test/integration  # integration tests (Postgres)
```

## Guidelines
- Follow the layer and naming rules in [`.claude/rules/architecture.md`](.claude/rules/architecture.md).
  Services depend on ports only, and every constructor goes into a `WireSet`.
- Every behaviour change comes with tests that follow [`.claude/rules/testing.md`](.claude/rules/testing.md):
  unit tests for services and handlers, and integration tests for datastore and HTTP changes.
- Regenerate code with `make generate` after changing constructors or interfaces; never edit
  `wire_gen.go` or files under `mock/` by hand.
- Run `make lint`, `make test` and `make test/integration` before opening a pull request.

## Commit messages
Use short imperative subjects ("Add order aggregate", "Fix subscriber shutdown"), with a body
explaining *why* when it isn't obvious.
