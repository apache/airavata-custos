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
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/codes"

	"github.com/apache/airavata-custos/internal/audit"
	"github.com/apache/airavata-custos/internal/db"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
)

// ErrPermanent marks an error a retry cannot fix, so the delivery fails at once.
var ErrPermanent = errors.New("permanent delivery failure")

const (
	pollInterval   = 10 * time.Second
	batchSize      = 50
	handlerTimeout = 2 * time.Minute
	maxAttempts    = 10
	retryBase      = 30 * time.Second
	retryMax       = time.Hour
)

// Run is started once, after the connectors have loaded. It first drops the
// subscriptions the loaded connectors no longer handle, then delivers due
// events until ctx is canceled.
func (b *Bus) Run(ctx context.Context) {
	b.pruneSubscriptionsOnStart(ctx)

	slog.Info("event delivery worker started", "interval", pollInterval)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	b.deliverDue(ctx)
	for {
		select {
		case <-ctx.Done():
			slog.Info("event delivery worker stopping")
			return
		case <-ticker.C:
			b.deliverDue(ctx)
		}
	}
}

// pruneSubscriptionsOnStart removes the subscriptions a loaded subscriber no
// longer has a handler for, which happens when a handler function is dropped
// from the connector's code and the system restarts. A subscriber that did
// not load (could be due to an issue) is skipped, so its events keep queuing.
func (b *Bus) pruneSubscriptionsOnStart(ctx context.Context) {
	loaded := make(map[string]struct{})
	for _, name := range b.loadedSubscribers() {
		loaded[name] = struct{}{}
	}

	stored, err := b.store.ListSubscriptions(ctx)
	if err != nil {
		slog.Error("event subscriptions not listed", "error", err)
		return
	}
	for _, s := range stored {
		// No handler at all means the connector did not load correctly, for a
		// broken config or because it is not part of this deployment. Leave it
		// as it is without deleting anything, so its events keep queuing.
		if _, ok := loaded[s.Subscriber]; !ok {
			continue
		}

		// The connector loaded and still handles this topic.
		if _, ok := b.handlerFor(s.Subscriber, s.EventType); ok {
			continue
		}

		// The connector loaded but no longer handles this topic, so the
		// handler was removed from its code. Drop the subscription.
		if err := b.store.DeleteSubscription(ctx, s.Subscriber, s.EventType); err != nil {
			slog.Error("stale event subscription not removed", "subscriber", s.Subscriber, "topic", s.EventType, "error", err)
			continue
		}
		b.mu.Lock()
		delete(b.subscriptions[s.EventType], s.Subscriber)
		b.mu.Unlock()
		slog.Info("stale event subscription removed", "subscriber", s.Subscriber, "topic", s.EventType)
	}
}

// deliverDue runs one batch of due deliveries for the subscribers that have a
// handler here. Rows for the others are left as they are.
func (b *Bus) deliverDue(ctx context.Context) {
	loaded := b.loadedSubscribers()
	if len(loaded) == 0 {
		return
	}
	rows, err := b.store.FindDueDeliveries(ctx, time.Now().UTC(), loaded, batchSize)
	if err != nil {
		slog.Error("due event deliveries not listed", "error", err)
		return
	}
	for _, row := range rows {
		if ctx.Err() != nil {
			return
		}
		b.deliver(ctx, row)
	}
}

func (b *Bus) loadedSubscribers() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	seen := make(map[string]struct{})
	var names []string
	for _, handlers := range b.topicHandlers {
		for _, h := range handlers {
			if _, ok := seen[h.subscriber]; !ok {
				seen[h.subscriber] = struct{}{}
				names = append(names, h.subscriber)
			}
		}
	}
	return names
}

// deliver calls the handler for one row and records the outcome.
func (b *Bus) deliver(ctx context.Context, row models.PendingDelivery) {
	handler, ok := b.handlerFor(row.Subscriber, row.Event.EventType)
	if !ok {
		return
	}

	// Build the handler's context. It carries the trace id and source of the
	// request that published the event, so audit rows join that trace. It is
	// detached from the worker's context, so shutdown does not cut the
	// handler off, and it has a timeout, so a handler that never returns
	// cannot stop the worker from moving on. `WithoutCancel` has to come first,
	// a timeout added straight on `ctx` (instead of `WithoutCancel`) would still be canceled by shutdown.
	handlerCtx := context.WithoutCancel(ctx)
	handlerCtx = tracing.ContextWithTraceID(handlerCtx, row.Event.TraceID)
	handlerCtx = audit.WithSource(handlerCtx, row.Event.Source)
	handlerCtx, cancel := context.WithTimeout(handlerCtx, handlerTimeout)
	defer cancel()
	handlerCtx, span := tracing.Start(handlerCtx, "bus.deliver:"+row.Event.EventType)
	defer span.End()

	event := Event{Type: EventType(row.Event.EventType), Payload: row.Event.Payload}
	err := callHandler(handlerCtx, handler, event, row.Event.Payload)
	attempts := row.Attempts + 1
	now := time.Now().UTC()

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	markErr := db.TxFn(ctx, b.db, func(tx *sql.Tx) error {
		switch {
		case err == nil:
			return b.store.MarkDeliverySucceeded(ctx, tx, row.ID, attempts, now)
		case errors.Is(err, ErrPermanent) || attempts >= maxAttempts:
			slog.ErrorContext(handlerCtx, "event delivery failed, no more retries", "subscriber", row.Subscriber, "topic", row.Event.EventType, "attempts", attempts, "error", err)
			return b.store.MarkDeliveryFailed(ctx, tx, row.ID, attempts, err.Error(), now)
		default:
			next := now.Add(NextRetryDelay(attempts, retryBase, retryMax))
			slog.WarnContext(handlerCtx, "event delivery failed, will retry", "subscriber", row.Subscriber, "topic", row.Event.EventType, "attempt", attempts, "next_run_at", next, "error", err)
			return b.store.MarkDeliveryRetry(ctx, tx, row.ID, attempts, next, err.Error())
		}
	})
	// The row stays pending, so the next tick runs it again. Handlers are
	// idempotent, so a repeat after a success is harmless.
	if markErr != nil {
		slog.Error("event delivery outcome not recorded", "delivery_id", row.ID, "error", markErr)
	}
}
