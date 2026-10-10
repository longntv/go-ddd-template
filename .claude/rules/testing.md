---
paths:
  - "**/*_test.go"
  - "internal/testutil/**"
  - "test/**"
  - "testdata/**"
---

# Testing rules

## What goes where
| Code under test | Test kind | Location | Real | Mocked |
|---|---|---|---|---|
| Service (use case) | unit | `internal/service/<file>_test.go`, package `service` | logic | every gateway + `EventPublisher` |
| HTTP handler | unit | `internal/handler/http/server/*_test.go`, package `server_test` | gin binding + status mapping | use cases |
| Entity / event routing | unit | next to the code, `_test` package | everything | nothing |
| Datastore reader/writer | integration | `internal/infrastructure/datastore/<agg>_{reader,writer}_test.go` | GORM + Postgres | nothing |
| HTTP end-to-end | integration | `test/integration/http/<agg>_test.go` | router → service → Postgres | `EventPublisher` and other external ports |

## Rules for every test
- Table-driven with a **map**: `tests := map[string]testcase{...}`; name keys as behaviour
  ("email already exists"), not as function names.
- `t.Parallel()` in the test function **and** in each `t.Run`.
- One `gomock.NewController(t)` per subtest, created inside `t.Run`. No `ctrl.Finish()` (automatic).
- Expectations are set in a `prepare func(*args, *fields)` per case, with explicit `.Times(n)`.
  Use `gomock.InOrder` when order is part of the behaviour; `DoAndReturn` to inspect arguments.
- Compare results with `cmp.Diff(want, got)` and report `(-want +got)`. Use
  `cmpopts.IgnoreFields` for wall-clock fields, `cmp.Comparer(time.Time.Equal)` (see `equalTime` in `datastore/user_reader_test.go`) for DB timestamps.
- Service errors: assert the `DomainError.Code` with `assertDomainErrorCode` (`service/user_creator_test.go`),
  not the message text.
- Cover at least: happy path, each domain error code the code can return, and each dependency failure.
- No `time.Sleep`, no real network, no AWS. Anything external is a port, and ports are mocked.

## Integration-only rules
- First line of every integration file: `//go:build integration`. Unit tests must pass with no services.
- Each package with integration tests has a `main_test.go` whose `TestMain` calls
  `testutil.InitTemplateDB` and `Release`s it after `m.Run()`.
- **Readers** use the package-level `readDB` (from `testutil.InitReadDB`) and may run in parallel.
  **Writers and HTTP tests** call `testutil.InitDB(t)` inside each subtest. Never write to `readDB`.
- Fixtures live in `testdata/fixtures/<table>.yml` (YAML list of rows keyed by column). IDs are
  fixed, readable UUIDs (`11111111-…`). Never change or delete an existing fixture row other tests
  rely on; add new rows instead, and update count assertions such as `fixtureUserCount`.
- Integration tests assert side effects in the DB (`CountUsers`, reading back via the reader), not
  only the HTTP response.

## Before finishing
`make test` must pass. If datastore, migrations, fixtures or `test/integration` changed, also run
`make test/integration` (needs Postgres) and `go vet -tags=integration ./...`.
