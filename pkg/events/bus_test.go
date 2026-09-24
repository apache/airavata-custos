//go:build integration

// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package events

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/internal/db"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
)

const subscriberTest = "test-subscriber"

func clusterUser() *models.ComputeClusterUser {
	return &models.ComputeClusterUser{ID: "cu-1", UserID: "u-1", ComputeClusterID: "c-1", LocalUsername: "jdoe"}
}

func onlyDelivery(t *testing.T, bus *Bus, status models.EventDeliveryStatus) models.PendingDelivery {
	t.Helper()
	rows, err := bus.store.ListDeliveriesByStatus(context.Background(), status, 10)
	if err != nil {
		t.Fatalf("list deliveries: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 %s delivery, got %d", status, len(rows))
	}
	return rows[0]
}

func countEvents(t *testing.T, database *sqlx.DB) int {
	t.Helper()
	var n int
	if err := database.Get(&n, "SELECT COUNT(*) FROM events"); err != nil {
		t.Fatalf("count events: %v", err)
	}
	return n
}

// Make sure a published event reaches the subscriber's handler as the model
// it was published as, under the publisher's trace, and the event delivery is then marked succeeded.
func TestPublishIsDeliveredToSubscriber(t *testing.T) {
	database := setupTestDB(t)
	bus := newBus(t, database)

	var got models.ComputeClusterUser
	var gotTrace string
	bus.SubscribeComputeClusterUserCreated(subscriberTest, func(ctx context.Context, cu models.ComputeClusterUser) error {
		got = cu
		gotTrace, _ = tracing.IDsFromContext(ctx)
		return nil
	})

	ctx, span := tracing.Start(context.Background(), "test.publish")
	wantTrace, _ := tracing.IDsFromContext(ctx)
	if err := bus.Publish(ctx, ComputeClusterUserCreateEvent, clusterUser()); err != nil {
		t.Fatalf("publish: %v", err)
	}
	span.End()
	bus.deliverDue(context.Background())

	if got.ID != "cu-1" || got.LocalUsername != "jdoe" {
		t.Fatalf("handler got %+v", got)
	}
	if gotTrace != wantTrace {
		t.Fatalf("handler trace %q, want the publisher's %q", gotTrace, wantTrace)
	}
	if d := onlyDelivery(t, bus, models.EventDeliverySucceeded); d.Attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", d.Attempts)
	}
}

// Make sure nothing is stored for an event nobody subscribes to.
func TestPublishWritesNothingWithoutSubscription(t *testing.T) {
	database := setupTestDB(t)
	bus := newBus(t, database)

	if err := bus.Publish(context.Background(), UserCreateEvent, &models.User{ID: "u-1"}); err != nil {
		t.Fatalf("publish: %v", err)
	}
	if n := countEvents(t, database); n != 0 {
		t.Fatalf("expected no stored events, got %d", n)
	}
}

// Make sure a handler that fails, by error or by panic, is tried again after
// the backoff and not before.
func TestFailedDeliveryIsRetriedAfterBackoff(t *testing.T) {
	for name, handler := range map[string]ComputeClusterUserHandler{
		"error": func(context.Context, models.ComputeClusterUser) error { return errors.New("registry 503") },
		"panic": func(context.Context, models.ComputeClusterUser) error { panic("boom") },
	} {
		t.Run(name, func(t *testing.T) {
			database := setupTestDB(t)
			bus := newBus(t, database)
			calls := 0
			bus.SubscribeComputeClusterUserCreated(subscriberTest, func(ctx context.Context, cu models.ComputeClusterUser) error {
				calls++
				return handler(ctx, cu)
			})

			if err := bus.Publish(context.Background(), ComputeClusterUserCreateEvent, clusterUser()); err != nil {
				t.Fatalf("publish: %v", err)
			}
			// The second pass runs before the backoff has passed, so it must skip the row.
			bus.deliverDue(context.Background())
			bus.deliverDue(context.Background())

			if calls != 1 {
				t.Fatalf("expected one call before the backoff has passed, got %d", calls)
			}
			d := onlyDelivery(t, bus, models.EventDeliveryPending)
			if d.Attempts != 1 || d.LastError == nil {
				t.Fatalf("expected 1 failed attempt with an error, got attempts=%d error=%v", d.Attempts, d.LastError)
			}
			// A little under retryBase, since the test itself takes some time.
			if wait := time.Until(d.NextRunAt); wait < retryBase-5*time.Second || wait > retryBase {
				t.Fatalf("expected the next run about %v away, got %v", retryBase, wait)
			}
		})
	}
}

// Make sure a delivery is marked failed once it reaches the attempt cap.
func TestDeliveryFailsAtAttemptCap(t *testing.T) {
	database := setupTestDB(t)
	bus := newBus(t, database)
	bus.SubscribeComputeClusterUserCreated(subscriberTest, func(context.Context, models.ComputeClusterUser) error {
		return errors.New("registry 503")
	})

	if err := bus.Publish(context.Background(), ComputeClusterUserCreateEvent, clusterUser()); err != nil {
		t.Fatalf("publish: %v", err)
	}
	// Jump the row to its last allowed attempt instead of failing it nine times.
	pending := onlyDelivery(t, bus, models.EventDeliveryPending)
	if err := db.TxFn(context.Background(), database, func(tx *sql.Tx) error {
		return bus.store.MarkDeliveryRetry(context.Background(), tx, pending.ID, maxAttempts-1, time.Now().Add(-time.Minute), "still down")
	}); err != nil {
		t.Fatalf("mark retry: %v", err)
	}
	bus.deliverDue(context.Background())

	if d := onlyDelivery(t, bus, models.EventDeliveryFailed); d.Attempts != maxAttempts {
		t.Fatalf("expected %d attempts, got %d", maxAttempts, d.Attempts)
	}
}

// Make sure an error wrapped in ErrPermanent fails the delivery at once, since a retry cannot fix it.
func TestPermanentErrorFailsWithoutRetry(t *testing.T) {
	database := setupTestDB(t)
	bus := newBus(t, database)
	bus.SubscribeComputeClusterUserCreated(subscriberTest, func(context.Context, models.ComputeClusterUser) error {
		return fmt.Errorf("%w: bad row", ErrPermanent)
	})

	if err := bus.Publish(context.Background(), ComputeClusterUserCreateEvent, clusterUser()); err != nil {
		t.Fatalf("publish: %v", err)
	}
	bus.deliverDue(context.Background())

	if d := onlyDelivery(t, bus, models.EventDeliveryFailed); d.Attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", d.Attempts)
	}
}

// Make sure an event published before its subscriber loads still gets a
// delivery from the saved subscription, waits untouched, and is delivered
// once the subscriber registers.
func TestEventPublishedBeforeSubscriberLoadsIsDelivered(t *testing.T) {
	database := setupTestDB(t)
	earlier := newBus(t, database)
	earlier.SubscribeComputeClusterUserCreated(subscriberTest, func(context.Context, models.ComputeClusterUser) error { return nil })

	// A fresh process: the subscription is saved, the handler is not loaded yet.
	bus := newBus(t, database)
	if err := bus.Publish(context.Background(), ComputeClusterUserCreateEvent, clusterUser()); err != nil {
		t.Fatalf("publish: %v", err)
	}
	bus.deliverDue(context.Background())
	if d := onlyDelivery(t, bus, models.EventDeliveryPending); d.Attempts != 0 {
		t.Fatalf("expected the row to wait untouched, got %d attempts", d.Attempts)
	}

	delivered := false
	bus.SubscribeComputeClusterUserCreated(subscriberTest, func(context.Context, models.ComputeClusterUser) error {
		delivered = true
		return nil
	})
	bus.deliverDue(context.Background())
	if !delivered {
		t.Fatal("expected the waiting event to be delivered once the subscriber registered")
	}
}

// Make sure an unsubscribed subscriber, which is what the loader does for a
// disabled connector, gets no new deliveries.
func TestUnsubscribeStopsNewDeliveries(t *testing.T) {
	database := setupTestDB(t)
	earlier := newBus(t, database)
	earlier.SubscribeComputeClusterUserCreated(subscriberTest, func(context.Context, models.ComputeClusterUser) error { return nil })

	bus := newBus(t, database)
	if err := bus.Unsubscribe(context.Background(), subscriberTest); err != nil {
		t.Fatalf("unsubscribe: %v", err)
	}
	if err := bus.Publish(context.Background(), ComputeClusterUserCreateEvent, clusterUser()); err != nil {
		t.Fatalf("publish: %v", err)
	}
	if n := countEvents(t, database); n != 0 {
		t.Fatalf("expected no stored events after unsubscribe, got %d", n)
	}
}

// Make sure the prune drops a topic a loaded subscriber no longer handles and
// keeps every topic of a subscriber that did not load.
func TestPruneKeepsSubscriptionsOfSubscribersThatDidNotLoad(t *testing.T) {
	database := setupTestDB(t)
	noop := func(context.Context, models.ComputeClusterUser) error { return nil }
	earlier := newBus(t, database)
	earlier.SubscribeComputeClusterUserCreated("loaded", noop)
	earlier.SubscribeComputeClusterUserDeleted("loaded", noop)
	earlier.SubscribeComputeClusterUserCreated("absent", noop)

	// A fresh process where "loaded" dropped the delete topic and "absent" did not start.
	bus := newBus(t, database)
	bus.SubscribeComputeClusterUserCreated("loaded", noop)
	bus.pruneSubscriptionsOnStart(context.Background())

	subs, err := bus.store.ListSubscriptions(context.Background())
	if err != nil {
		t.Fatalf("list subscriptions: %v", err)
	}
	var kept []string
	for _, s := range subs {
		kept = append(kept, s.Subscriber+" "+s.EventType)
	}
	want := []string{"absent " + string(ComputeClusterUserCreateEvent), "loaded " + string(ComputeClusterUserCreateEvent)}
	if !slices.Equal(kept, want) {
		t.Fatalf("kept %v, want %v", kept, want)
	}
}

// Make sure the sync publish stores nothing, stops at the first failing
// handler, and returns its error.
func TestPublishSyncReturnsHandlerError(t *testing.T) {
	database := setupTestDB(t)
	bus := newBus(t, database)

	want := errors.New("boom")
	var secondRan bool
	bus.Subscribe(subscriberTest, ComputeClusterUserCreateEvent, func(context.Context, Event, interface{}) error {
		return want
	})
	bus.Subscribe(subscriberTest, ComputeClusterUserCreateEvent, func(context.Context, Event, interface{}) error {
		secondRan = true
		return nil
	})

	if err := bus.PublishSync(context.Background(), ComputeClusterUserCreateEvent, nil); !errors.Is(err, want) {
		t.Fatalf("expected handler error, got %v", err)
	}
	if secondRan {
		t.Fatal("expected sync publish to stop at the first error")
	}
	if n := countEvents(t, database); n != 0 {
		t.Fatalf("expected sync publish to store nothing, got %d events", n)
	}
}

// Make sure a handler panic reaches the caller of the sync publish. The
// stored path records it as a failed attempt instead.
func TestPublishSyncPanicPropagatesToCaller(t *testing.T) {
	database := setupTestDB(t)
	bus := newBus(t, database)
	bus.Subscribe(subscriberTest, ComputeClusterUserCreateEvent, func(context.Context, Event, interface{}) error {
		panic("boom")
	})

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		_ = bus.PublishSync(context.Background(), ComputeClusterUserCreateEvent, nil)
	}()

	if recovered == nil {
		t.Fatal("expected sync publish to surface subscriber panic to caller")
	}
}
