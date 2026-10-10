---
name: ddd-testing
description: Write and run unit and integration tests in this Go DDD template — gomock table-driven service and handler tests, datastore reader/writer tests and end-to-end HTTP tests on per-test Postgres databases cloned from a template DB. Use when adding tests, fixing failing tests, adding fixtures, or testing a newly implemented feature.
---

# Testing in go-ddd-template

Rules that always apply are in `.claude/rules/testing.md`. Reference tests to copy:

| Kind | Copy from |
|---|---|
| Service unit test | `internal/service/user_creator_test.go` (has `assertDomainErrorCode`, `ignoreTimestamps`) |
| Handler unit test | `internal/handler/http/server/handler_test.go` |
| Datastore reader (integration) | `internal/infrastructure/datastore/user_reader_test.go` |
| Datastore writer (integration) | `internal/infrastructure/datastore/user_writer_test.go` |
| HTTP end-to-end (integration) | `test/integration/http/users_test.go` + `helper_test.go` |

## 1. Service unit test
Package `service` (same package, so it can use the shared helpers). Skeleton:
```go
import (
	// context, errors, testing, cmp, gomock, input, output ...
	mockgateway "github.com/longntv/go-ddd-template/internal/domain/gateway/mock"
)

func Test_createOrder_Execute(t *testing.T) {
	t.Parallel()

	var ( /* shared inputs, entities, errDB := errors.New("database connection error") */ )

	type fields struct {
		mockCommands  *mockgateway.MockOrderCommandsGateway
		mockQueries   *mockgateway.MockOrderQueriesGateway
		mockPublisher *mockgateway.MockEventPublisher
	}
	type args struct {
		ctx context.Context
		in  *input.CreateOrder
	}
	type testcase struct {
		prepare     func(*args, *fields)
		args        args
		expected    *output.CreateOrder
		wantErrCode string // "" = no error
		wantPublishErrLogged string // event type whose failed Publish must be logged; "" = none
	}

	tests := map[string]testcase{
		"successfully create order and publish event": { /* prepare: EXPECT()...Times(1) */ },
		"<dependency> returns error":                    { /* wantErrCode: "INTERNAL" */ },
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			logCore, logs := observer.New(zap.ErrorLevel) // go.uber.org/zap/zaptest/observer
			f := &fields{ /* mockgateway.NewMock...(ctrl) */ }
			if tt.prepare != nil {
				tt.prepare(&tt.args, f)
			}

			actual, err := NewCreateOrder(f.mockCommands, f.mockQueries, f.mockPublisher, zap.New(logCore)).Execute(tt.args.ctx, tt.args.in)

			assertDomainErrorCode(t, err, tt.wantErrCode)
			assertPublishErrLogged(t, logs, tt.wantPublishErrLogged)
			if diff := cmp.Diff(tt.expected, actual); diff != "" {
				t.Errorf("createOrder.Execute() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
```
Cases to include: happy path (assert the published event type with `DoAndReturn`), every domain
error code, every dependency error, and "publish failure does not fail the request"
(`wantPublishErrLogged: event.OrderCreatedEvent`): services publish through `publishBestEffort`, which logs the
failure instead of returning it. Use `gomock.InOrder` when the order of calls matters.

## 2. Handler unit test
Package `server_test`. Build a `gin.New()` engine with the handler and only the routes under test,
mock the use cases with `mockusecase`, drive it with `httptest`. Assert status codes for binding
errors (400), each mapped domain code (404/409) and unexpected errors (500 with
`"internal server error"`), and the JSON body for success.

## 3. Datastore integration tests
- First line `//go:build integration`, package `datastore_test`.
- Fixtures first: add `testdata/fixtures/<table>.yml`. The file name must equal the table name.
  It's a list of rows keyed by column, with fixed readable UUIDs and timestamps like
  `"2025-01-01 10:00:00"`. Rows must satisfy FKs and unique constraints. Fixture files load in
  alphabetical order; if a table has an FK to another one, name and order them so parents come first.
- Reader test: use the package-level `readDB`, `t.Parallel()` everywhere, compare with
  `cmp.Diff(want, got, equalTime)`; cover found / not found / pagination / empty.
- Writer test: inside each subtest call `db, _ := testutil.InitDB(t)` (fresh clone), perform the
  write, then **read it back** through the reader. Cover success, constraint violation, not found.
- The package already has `main_test.go` with `TestMain` and `readDB`; don't add a second one.

## 4. HTTP end-to-end integration test
- If the feature added services or ports, first update `test/integration/registry/wire.go`:
  add providers (`datastore.New<Agg>Reader`, ...) and take external ports as parameters. Then run
  `make generate`.
- If `HTTPTestHelper` needs a new mock, add the field and construct it in `NewHTTPTestHelper`.
- In `test/integration/http/<agg>_test.go`, use one `NewHTTPTestHelper(t)` per subtest. Call
  `h.Do(t, method, path, body)`, then assert the status, the response fields, and the DB side
  effect (count rows or read back through the API).
- Set publisher expectations with `expectPublished(t, h, event.<X>Event)`.

## 5. Run
```bash
make test                                  # unit, always
go vet -tags=integration ./...             # integration files compile
make docker/up && make test/integration    # needs Postgres
```
If port 5432 is taken, start a throwaway Postgres and point the tests at it:
```bash
docker run --rm -d --name ddd-test-pg -e POSTGRES_PASSWORD=postgres -p 55432:5432 postgres:16-alpine
DB_PORT=55432 make test/integration
docker stop ddd-test-pg
```

## Troubleshooting
| Symptom | Cause / fix |
|---|---|
| `ping postgres ... connection refused` | Postgres not running or wrong `DB_HOST`/`DB_PORT`. |
| `load fixture <file>: insert row N` | Column name typo, constraint violation, or FK parent loaded later (file order). |
| `source database ... is being accessed by other users` | Something holds a connection to the template DB; only `testutil` may connect to it. |
| Reader test sees data written by another test | That test wrote to `readDB`; writers must use `testutil.InitDB(t)`. |
| `missing call(s)` / `unexpected call` from gomock | The `prepare` expectations don't match what the code does. Check arguments and `.Times`. |
| Leftover `go_ddd_template_test_*` databases | A test binary was killed. Drop them with `psql -c 'DROP DATABASE ... WITH (FORCE)'`. |

Report the exact commands run and their outcome; say explicitly if integration tests were not run.
