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
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/codes"

	"github.com/apache/airavata-custos/internal/audit"
	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
)

// New creates a bus with the existing event subscriptions loaded.
func New(ctx context.Context, database *sqlx.DB) (*Bus, error) {
	b := &Bus{
		db:            database,
		store:         store.NewEventDeliveryStore(database),
		auditEvents:   store.NewAuditEventStore(database),
		topicHandlers: make(map[string][]handler),
		subscriptions: make(map[string]map[string]struct{}),
	}
	rows, err := b.store.ListSubscriptions(ctx)
	if err != nil {
		return nil, fmt.Errorf("load event subscriptions: %w", err)
	}
	for _, r := range rows {
		if b.subscriptions[r.EventType] == nil {
			b.subscriptions[r.EventType] = make(map[string]struct{})
		}
		b.subscriptions[r.EventType][r.Subscriber] = struct{}{}
	}
	return b, nil
}

// Subscribe registers a handler under the subscriber's name and saves the subscription.
func (b *Bus) Subscribe(subscriber string, topic EventType, fn EventSubscriberFunc) {
	b.mu.Lock()
	b.topicHandlers[string(topic)] = append(b.topicHandlers[string(topic)], handler{subscriber: subscriber, fn: fn})
	if b.subscriptions[string(topic)] == nil {
		b.subscriptions[string(topic)] = make(map[string]struct{})
	}
	b.subscriptions[string(topic)][subscriber] = struct{}{}
	b.mu.Unlock()

	// Runs at startup, so there is no request context to pass.
	if err := b.store.SaveSubscription(context.Background(), subscriber, string(topic)); err != nil {
		slog.Error("event subscription not saved", "subscriber", subscriber, "topic", topic, "error", err)
	}
}

// Unsubscribe drops every saved subscription of a subscriber. The loader
// calls it for a disabled connector, so no new deliveries are written for it.
func (b *Bus) Unsubscribe(ctx context.Context, subscriber string) error {
	if err := b.store.DeleteSubscriptions(ctx, subscriber); err != nil {
		return fmt.Errorf("delete event subscriptions: %w", err)
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, names := range b.subscriptions {
		delete(names, subscriber)
	}
	return nil
}

// Subscription is a saved subscription and whether its subscriber is loaded.
// Not loaded means the connector did not start, so its deliveries stay pending.
type Subscription struct {
	models.EventSubscription
	Loaded bool `json:"loaded" binding:"required"`
}

// ListSubscriptions returns every saved subscription with its loaded flag.
func (b *Bus) ListSubscriptions(ctx context.Context) ([]Subscription, error) {
	rows, err := b.store.ListSubscriptions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Subscription, 0, len(rows))
	for _, r := range rows {
		_, loaded := b.handlerFor(r.Subscriber, r.EventType)
		out = append(out, Subscription{EventSubscription: r, Loaded: loaded})
	}
	return out, nil
}

// subscribeTyped wraps a handler that takes the payload as T. The payload
// comes as JSON from a delivery row, or as T or *T from PublishSync.
func subscribeTyped[T any](b *Bus, subscriber string, topic EventType, handler func(context.Context, T) error) {
	b.Subscribe(subscriber, topic, func(ctx context.Context, event Event, value interface{}) error {
		switch v := value.(type) {
		case json.RawMessage:
			var t T
			if err := json.Unmarshal(v, &t); err != nil {
				return fmt.Errorf("%w: decode payload: %v", ErrPermanent, err)
			}
			return handler(ctx, t)
		case T:
			return handler(ctx, v)
		case *T:
			if v != nil {
				return handler(ctx, *v)
			}
		default:
			slog.WarnContext(ctx, "event payload has unexpected type", "type", event.Type, "got", value)
		}
		return nil
	})
}

// Publish stores the event with one delivery per saved subscriber, and the
// worker delivers it later. It writes the event and its deliveries in the caller's
// transaction, so everything is committed or rolled back together.
// Nothing is written when nobody subscribes.
func (b *Bus) Publish(ctx context.Context, tx *sql.Tx, topic EventType, payload any) error {
	subscribers := b.subscribersOf(string(topic))
	if len(subscribers) == 0 {
		return nil
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode event payload: %w", err)
	}
	traceID, spanID, err := tracing.IDsFromContext(ctx)
	if err != nil {
		return fmt.Errorf("publish %s: %w", topic, err)
	}
	source := audit.SourceFromContext(ctx)
	// FIXME - remove defaulting to 'core' every entry point should put its source
	if source == "" {
		source = "core"
	}
	now := time.Now().UTC()
	event := &models.Event{
		ID:        uuid.NewString(),
		EventType: string(topic),
		Payload:   body,
		Source:    source,
		TraceID:   traceID,
		SpanID:    spanID,
		CreatedAt: now,
	}

	if err := b.store.CreateEvent(ctx, tx, event); err != nil {
		return fmt.Errorf("store event: %w", err)
	}
	// Create a delivery for each subscriber with a PENDING status
	for _, name := range subscribers {
		d := &models.EventDelivery{
			ID:         uuid.NewString(),
			EventID:    event.ID,
			Subscriber: name,
			Status:     models.EventDeliveryPending,
			NextRunAt:  now,
			CreatedAt:  now,
		}
		if err := b.store.CreateDelivery(ctx, tx, d); err != nil {
			return fmt.Errorf("store event delivery: %w", err)
		}
	}
	return nil
}

func (b *Bus) subscribersOf(topic string) []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	names := make([]string, 0, len(b.subscriptions[topic]))
	for name := range b.subscriptions[topic] {
		names = append(names, name)
	}
	return names
}

// handlerFor returns the subscriber's handler for a topic, or false if the
// subscriber is not loaded.
func (b *Bus) handlerFor(subscriber, topic string) (EventSubscriberFunc, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, h := range b.topicHandlers[topic] {
		if h.subscriber == subscriber {
			return h.fn, true
		}
	}
	return nil, false
}

// callHandler runs a handler and turns a panic into an error, so a panic
// counts as a failed attempt like any other.
func callHandler(ctx context.Context, h EventSubscriberFunc, event Event, payload any) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("subscriber panic: %v", r)
			slog.ErrorContext(ctx, "event subscriber panicked", "topic", event.Type, "panic", r, "stack", string(debug.Stack()))
		}
	}()
	return h(ctx, event, payload)
}

// PublishSync calls the handlers in the caller's goroutine and stores nothing.
// Use it when the order matters or the caller has to wait. It stops at the
// first handler error and returns it.
func (b *Bus) PublishSync(ctx context.Context, topic EventType, payload any) error {
	ctx, span := tracing.Start(ctx, "bus.publish:"+string(topic))
	defer span.End()

	b.mu.RLock()
	handlers := make([]handler, len(b.topicHandlers[string(topic)]))
	copy(handlers, b.topicHandlers[string(topic)])
	b.mu.RUnlock()

	event := Event{Type: topic, Payload: payload}
	for _, h := range handlers {
		if err := dispatchSync(ctx, h, event, payload); err != nil {
			return err
		}
	}
	return nil
}

func dispatchSync(ctx context.Context, h handler, event Event, payload any) error {
	ctx, span := tracing.Start(ctx, "bus.subscribe:"+string(event.Type))
	defer span.End()

	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("subscriber panic: %v", r)
			span.RecordError(err)
			span.SetStatus(codes.Error, "subscriber panic")
			slog.ErrorContext(ctx, "event subscriber panicked",
				"topic", event.Type,
				"subscriber", h.subscriber,
				"panic", r,
				"stack", string(debug.Stack()),
			)
			panic(r)
		}
	}()
	err := h.fn(ctx, event, payload)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}
