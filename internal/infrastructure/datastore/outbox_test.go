//go:build integration

package datastore_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"
	"gorm.io/gorm"

	appconfig "github.com/longntv/go-ddd-template/internal/config"
	"github.com/longntv/go-ddd-template/internal/domain/entity"
	"github.com/longntv/go-ddd-template/internal/domain/event"
	"github.com/longntv/go-ddd-template/internal/infrastructure/datastore"
	"github.com/longntv/go-ddd-template/internal/testutil"

	mockgateway "github.com/longntv/go-ddd-template/internal/domain/gateway/mock"
)

// Outbox rows from testdata/fixtures/outbox_events.yml.
var (
	fixtureOutboxDue       = uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001")
	fixtureOutboxNotDueYet = uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000002")
	fixtureOutboxPublished = uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000003")
)

// equalToMillisecond compares timestamps the database stores with
// millisecond precision.
var equalToMillisecond = cmp.Comparer(func(a, b time.Time) bool { return a.Sub(b).Abs() < time.Millisecond })

// newUserEvent builds an event about u, as the services do.
func newUserEvent(eventType string, u *entity.User) *event.DomainEvent {
	return event.NewDomainEvent(eventType, event.Source, u.ID.String(),
		&event.UserEventData{ID: u.ID.String(), Name: u.Name, Email: u.Email})
}

// addToOutbox saves events in one committed transaction.
func addToOutbox(t *testing.T, db *gorm.DB, events ...*event.DomainEvent) {
	t.Helper()

	err := datastore.NewTransactor(db).RunInTx(context.Background(), func(ctx context.Context) error {
		for _, evt := range events {
			if err := datastore.NewEventOutbox(db).Add(ctx, evt); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("add events to outbox: %v", err)
	}
}

// outboxRow reads one outbox row, failing the test when it is missing.
func outboxRow(t *testing.T, db *gorm.DB, id uuid.UUID) *datastore.OutboxEventEntity {
	t.Helper()

	var row datastore.OutboxEventEntity
	if err := db.Where("id = ?", id).First(&row).Error; err != nil {
		t.Fatalf("read outbox row %s: %v", id, err)
	}
	return &row
}

func newRelay(t *testing.T, db *gorm.DB, publisher *mockgateway.MockEventPublisher, batchSize int, logger *zap.Logger) *datastore.OutboxRelay {
	t.Helper()

	cfg := &appconfig.Config{Outbox: appconfig.OutboxConfig{PollInterval: time.Hour, BatchSize: batchSize}}
	relay, err := datastore.NewOutboxRelay(db, publisher, cfg, logger)
	if err != nil {
		t.Fatalf("NewOutboxRelay() error = %v", err)
	}
	return relay
}

func Test_transactor_RunInTx(t *testing.T) {
	t.Parallel()

	dave := &entity.User{ID: uuid.MustParse("44444444-4444-4444-4444-444444444444"), Name: "Dave", Email: "dave@example.com", PasswordHash: "$2a$04$dave-hash"}
	errFail := errors.New("later step failed")

	tests := map[string]struct {
		failAfterWrites bool
		wantErr         error
		wantSaved       int64
	}{
		"commits the user and its event together":        {wantSaved: 1},
		"rolls back the user and its event on any error": {failAfterWrites: true, wantErr: errFail},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db, _ := testutil.InitDB(t)
			evt := newUserEvent(event.UserCreatedEvent, dave)

			err := datastore.NewTransactor(db).RunInTx(context.Background(), func(ctx context.Context) error {
				if err := datastore.NewUserWriter(db).Create(ctx, dave); err != nil {
					return err
				}
				// Readers see the transaction's own uncommitted writes.
				if _, err := datastore.NewUserReader(db).Get(ctx, dave.ID); err != nil {
					t.Errorf("Get() inside the transaction error = %v", err)
				}
				if err := datastore.NewEventOutbox(db).Add(ctx, evt); err != nil {
					return err
				}
				if tt.failAfterWrites {
					return errFail
				}
				return nil
			})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("RunInTx() error = %v, want %v", err, tt.wantErr)
			}

			var users, events int64
			if err := db.Model(&datastore.UserEntity{}).Where("id = ?", dave.ID).Count(&users).Error; err != nil {
				t.Fatalf("count users: %v", err)
			}
			if err := db.Model(&datastore.OutboxEventEntity{}).Where("id = ?", evt.ID).Count(&events).Error; err != nil {
				t.Fatalf("count outbox events: %v", err)
			}
			if users != tt.wantSaved || events != tt.wantSaved {
				t.Errorf("saved users = %d, events = %d, want %d of each", users, events, tt.wantSaved)
			}
		})
	}
}

func Test_eventOutbox_Add(t *testing.T) {
	t.Parallel()

	t.Run("stores the event as a pending row", func(t *testing.T) {
		t.Parallel()

		db, _ := testutil.InitDB(t)
		evt := newUserEvent(event.UserUpdatedEvent, fixtureAlice)
		addToOutbox(t, db, evt)

		got := outboxRow(t, db, evt.ID)
		want := &datastore.OutboxEventEntity{
			ID: evt.ID, Seq: got.Seq, Type: event.UserUpdatedEvent, Source: event.Source,
			Subject: fixtureAlice.ID.String(), Data: got.Data, OccurredAt: evt.Timestamp,
		}
		// NextAttemptAt comes from the database clock (checked below).
		if diff := cmp.Diff(want, got, equalToMillisecond, cmpopts.IgnoreFields(datastore.OutboxEventEntity{}, "NextAttemptAt")); diff != "" {
			t.Errorf("outbox row mismatch (-want +got):\n%s", diff)
		}
		assertJSONEqual(t, `{"id":"11111111-1111-1111-1111-111111111111","name":"Alice","email":"alice@example.com"}`, got.Data)
		if due := secondsUntilNextAttempt(t, db, evt.ID); due > 0 {
			t.Errorf("new event is due in %.3fs, want due at once", due)
		}
	})

	t.Run("refuses to run outside a transaction", func(t *testing.T) {
		t.Parallel()

		db, _ := testutil.InitDB(t)
		evt := newUserEvent(event.UserUpdatedEvent, fixtureAlice)

		if err := datastore.NewEventOutbox(db).Add(context.Background(), evt); err == nil {
			t.Fatal("Add() outside RunInTx error = nil, want an error")
		}
		var n int64
		if err := db.Model(&datastore.OutboxEventEntity{}).Where("id = ?", evt.ID).Count(&n).Error; err != nil {
			t.Fatalf("count outbox events: %v", err)
		}
		if n != 0 {
			t.Errorf("Add() outside RunInTx stored %d rows, want 0", n)
		}
	})
}

func Test_OutboxRelay_RelayBatch(t *testing.T) {
	t.Parallel()

	t.Run("publishes due events oldest first and marks them published", func(t *testing.T) {
		t.Parallel()

		db, _ := testutil.InitDB(t)
		first := newUserEvent(event.UserUpdatedEvent, fixtureAlice)
		second := newUserEvent(event.UserDeletedEvent, fixtureBob)
		addToOutbox(t, db, first, second)

		publisher := mockgateway.NewMockEventPublisher(gomock.NewController(t))
		var published []uuid.UUID
		publisher.EXPECT().
			Publish(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, evt *event.DomainEvent) error {
				published = append(published, evt.ID)
				if evt.ID == first.ID {
					assertPublishedAsSaved(t, first, evt)
				}
				return nil
			}).
			Times(3)

		relay := newRelay(t, db, publisher, 10, zaptest.NewLogger(t))
		if n, err := relay.RelayBatch(context.Background()); err != nil || n != 3 {
			t.Fatalf("RelayBatch() = %d, %v, want 3, nil", n, err)
		}
		if diff := cmp.Diff([]uuid.UUID{fixtureOutboxDue, first.ID, second.ID}, published); diff != "" {
			t.Errorf("publish order mismatch (-want +got):\n%s", diff)
		}
		for _, id := range []uuid.UUID{fixtureOutboxDue, first.ID, second.ID} {
			if outboxRow(t, db, id).PublishedAt == nil {
				t.Errorf("event %s not marked published", id)
			}
		}
		if outboxRow(t, db, fixtureOutboxNotDueYet).PublishedAt != nil {
			t.Error("event whose retry is not due was published")
		}

		// Nothing is left: a second batch publishes nothing (the mock allows 3 calls).
		if n, err := relay.RelayBatch(context.Background()); err != nil || n != 0 {
			t.Errorf("second RelayBatch() = %d, %v, want 0, nil", n, err)
		}
	})

	t.Run("a failed publish is logged and scheduled for retry", func(t *testing.T) {
		t.Parallel()

		db, _ := testutil.InitDB(t)
		publisher := mockgateway.NewMockEventPublisher(gomock.NewController(t))
		publisher.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(errors.New("sns unavailable")).Times(1)
		logCore, logs := observer.New(zap.ErrorLevel)

		relay := newRelay(t, db, publisher, 10, zap.New(logCore))
		if n, err := relay.RelayBatch(context.Background()); err != nil || n != 1 {
			t.Fatalf("RelayBatch() = %d, %v, want 1, nil", n, err)
		}

		row := outboxRow(t, db, fixtureOutboxDue)
		if row.PublishedAt != nil {
			t.Error("failed event marked published")
		}
		if row.Attempts != 1 || row.LastError != "sns unavailable" {
			t.Errorf("attempts = %d, last_error = %q, want 1, %q", row.Attempts, row.LastError, "sns unavailable")
		}
		// First failure: retry after 1s, measured on the database clock.
		if due := secondsUntilNextAttempt(t, db, fixtureOutboxDue); due <= 0 || due > 1 {
			t.Errorf("retry due in %.3fs, want within the 1s backoff", due)
		}

		entries := logs.FilterMessage("failed to publish outbox event").All()
		if len(entries) != 1 {
			t.Fatalf("logged %d publish failures, want 1", len(entries))
		}
		fields := entries[0].ContextMap()
		if fields["event_id"] != fixtureOutboxDue.String() || fields["attempts"] != int64(1) || fields["error"] != "sns unavailable" {
			t.Errorf("publish failure log fields = %v, want event_id, attempts and error", fields)
		}

		// The retry is not due yet, so the next batch skips it (the mock allows 1 call).
		if n, err := relay.RelayBatch(context.Background()); err != nil || n != 0 {
			t.Errorf("second RelayBatch() = %d, %v, want 0, nil", n, err)
		}
	})

	t.Run("publishes at most batch size events", func(t *testing.T) {
		t.Parallel()

		db, _ := testutil.InitDB(t)
		addToOutbox(t, db, newUserEvent(event.UserUpdatedEvent, fixtureAlice))
		publisher := mockgateway.NewMockEventPublisher(gomock.NewController(t))
		publisher.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(nil).Times(1)

		if n, err := newRelay(t, db, publisher, 1, zaptest.NewLogger(t)).RelayBatch(context.Background()); err != nil || n != 1 {
			t.Errorf("RelayBatch() = %d, %v, want 1, nil", n, err)
		}
	})

	t.Run("a publish cut short by shutdown is not counted as a failure", func(t *testing.T) {
		t.Parallel()

		db, _ := testutil.InitDB(t)
		ctx, cancel := context.WithCancel(context.Background())
		publisher := mockgateway.NewMockEventPublisher(gomock.NewController(t))
		publisher.EXPECT().
			Publish(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, _ *event.DomainEvent) error {
				cancel()
				return ctx.Err()
			}).
			Times(1)

		if n, err := newRelay(t, db, publisher, 10, zaptest.NewLogger(t)).RelayBatch(ctx); err != nil || n != 0 {
			t.Fatalf("RelayBatch() = %d, %v, want 0, nil", n, err)
		}
		if row := outboxRow(t, db, fixtureOutboxDue); row.Attempts != 0 || row.PublishedAt != nil {
			t.Errorf("cancelled event recorded: attempts = %d, published_at = %v", row.Attempts, row.PublishedAt)
		}
	})

	t.Run("shutdown keeps what was published and leaves the rest pending", func(t *testing.T) {
		t.Parallel()

		db, _ := testutil.InitDB(t)
		later := newUserEvent(event.UserUpdatedEvent, fixtureAlice)
		addToOutbox(t, db, later)
		ctx, cancel := context.WithCancel(context.Background())
		publisher := mockgateway.NewMockEventPublisher(gomock.NewController(t))
		publisher.EXPECT().
			Publish(gomock.Any(), gomock.Any()).
			DoAndReturn(func(context.Context, *event.DomainEvent) error {
				cancel() // shutdown arrives after the first publish succeeded
				return nil
			}).
			Times(1)

		if n, err := newRelay(t, db, publisher, 10, zaptest.NewLogger(t)).RelayBatch(ctx); err != nil || n != 1 {
			t.Fatalf("RelayBatch() = %d, %v, want 1, nil", n, err)
		}
		if outboxRow(t, db, fixtureOutboxDue).PublishedAt == nil {
			t.Error("event published before shutdown not marked published; it would be published again")
		}
		if outboxRow(t, db, later.ID).PublishedAt != nil {
			t.Error("event after shutdown marked published without being published")
		}
	})

	t.Run("stands by while another relay is working", func(t *testing.T) {
		t.Parallel()

		db, _ := testutil.InitDB(t)
		ctrl := gomock.NewController(t)
		inPublish, release := make(chan struct{}), make(chan struct{})
		working := mockgateway.NewMockEventPublisher(ctrl)
		working.EXPECT().
			Publish(gomock.Any(), gomock.Any()).
			DoAndReturn(func(context.Context, *event.DomainEvent) error {
				close(inPublish)
				<-release
				return nil
			}).
			Times(1)
		standby := mockgateway.NewMockEventPublisher(ctrl) // no calls expected

		done := make(chan error, 1)
		go func() {
			_, err := newRelay(t, db, working, 10, zaptest.NewLogger(t)).RelayBatch(context.Background())
			done <- err
		}()
		<-inPublish

		n, err := newRelay(t, db, standby, 10, zaptest.NewLogger(t)).RelayBatch(context.Background())
		close(release)
		if err != nil || n != 0 {
			t.Errorf("standby RelayBatch() = %d, %v, want 0, nil", n, err)
		}
		if err := <-done; err != nil {
			t.Errorf("working RelayBatch() error = %v", err)
		}
	})
}

// Two relays polling at once (e.g. two server replicas) never publish the
// same event twice.
func Test_OutboxRelay_concurrentRelaysPublishEachEventOnce(t *testing.T) {
	t.Parallel()

	db, _ := testutil.InitDB(t)
	events := make([]*event.DomainEvent, 20)
	for i := range events {
		events[i] = newUserEvent(event.UserUpdatedEvent, fixtureAlice)
	}
	addToOutbox(t, db, events...)
	wantPublished := len(events) + 1 // plus the due fixture row

	var (
		mu        sync.Mutex
		published = map[uuid.UUID]int{}
	)
	publisher := mockgateway.NewMockEventPublisher(gomock.NewController(t))
	publisher.EXPECT().
		Publish(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, evt *event.DomainEvent) error {
			mu.Lock()
			defer mu.Unlock()
			published[evt.ID]++
			return nil
		}).
		Times(wantPublished)

	var wg sync.WaitGroup
	for range 2 {
		relay := newRelay(t, db, publisher, 5, zaptest.NewLogger(t))
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				n, err := relay.RelayBatch(context.Background())
				if err != nil {
					t.Errorf("RelayBatch() error = %v", err)
					return
				}
				if n == 0 {
					return
				}
			}
		}()
	}
	wg.Wait()

	if len(published) != wantPublished {
		t.Errorf("published %d distinct events, want %d", len(published), wantPublished)
	}
	for id, n := range published {
		if n != 1 {
			t.Errorf("event %s published %d times, want once", id, n)
		}
	}
	if _, ok := published[fixtureOutboxPublished]; ok {
		t.Error("already published fixture event was published again")
	}
}

func Test_OutboxRelay_Run(t *testing.T) {
	t.Parallel()

	db, _ := testutil.InitDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	publisher := mockgateway.NewMockEventPublisher(gomock.NewController(t))
	publisher.EXPECT().
		Publish(gomock.Any(), gomock.Any()).
		DoAndReturn(func(context.Context, *event.DomainEvent) error {
			cancel() // stop the relay once it has published the due event
			return nil
		}).
		Times(1)

	if err := newRelay(t, db, publisher, 10, zaptest.NewLogger(t)).Run(ctx); !errors.Is(err, context.Canceled) {
		t.Errorf("Run() error = %v, want context.Canceled", err)
	}
	if outboxRow(t, db, fixtureOutboxDue).PublishedAt == nil {
		t.Error("event published before shutdown not marked published")
	}
}

// secondsUntilNextAttempt returns how long until the event is due, measured
// on the database clock (the one the relay uses).
func secondsUntilNextAttempt(t *testing.T, db *gorm.DB, id uuid.UUID) float64 {
	t.Helper()

	var secs float64
	err := db.Raw("SELECT EXTRACT(EPOCH FROM next_attempt_at - now())::float8 FROM outbox_events WHERE id = ?", id).Scan(&secs).Error
	if err != nil {
		t.Fatalf("read next attempt of %s: %v", id, err)
	}
	return secs
}

// assertPublishedAsSaved checks that the relay publishes the CloudEvent the
// service would have published directly: same attributes and payload.
func assertPublishedAsSaved(t *testing.T, saved, published *event.DomainEvent) {
	t.Helper()

	want, err := saved.ToCloudEvent()
	if err != nil {
		t.Fatalf("saved ToCloudEvent() error = %v", err)
	}
	got, err := published.ToCloudEvent()
	if err != nil {
		t.Fatalf("published ToCloudEvent() error = %v", err)
	}

	type attrs struct {
		ID, Type, Source, Subject, DataContentType string
		Time                                       time.Time
	}
	wantAttrs := attrs{want.ID(), want.Type(), want.Source(), want.Subject(), want.DataContentType(), want.Time()}
	gotAttrs := attrs{got.ID(), got.Type(), got.Source(), got.Subject(), got.DataContentType(), got.Time()}
	if diff := cmp.Diff(wantAttrs, gotAttrs, equalToMillisecond); diff != "" {
		t.Errorf("published CloudEvent attributes mismatch (-want +got):\n%s", diff)
	}
	assertJSONEqual(t, string(want.Data()), got.Data())
}

// assertJSONEqual compares two JSON documents ignoring formatting; Postgres
// jsonb does not keep the original spacing.
func assertJSONEqual(t *testing.T, want string, got []byte) {
	t.Helper()

	var w, g any
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		t.Fatalf("decode want JSON: %v", err)
	}
	if err := json.Unmarshal(got, &g); err != nil {
		t.Fatalf("decode got JSON %q: %v", got, err)
	}
	if diff := cmp.Diff(w, g); diff != "" {
		t.Errorf("JSON mismatch (-want +got):\n%s", diff)
	}
}
