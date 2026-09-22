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
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/apache/airavata-custos/internal/tracing"
	"go.opentelemetry.io/otel/codes"
)

func New() *Bus {
	return &Bus{
		subs: make(map[string][]subscription),
	}
}

// Subscribe registers a handler for a given topic under the subscriber's name.
// The handler is called asynchronously (in a new goroutine) each time
// an event is published on that topic.
func (b *Bus) Subscribe(subscriber string, topic EventType, handler EventSubscriberFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs[string(topic)] = append(b.subs[string(topic)], subscription{subscriber: subscriber, handler: handler})
}

// subscribeTyped registers a handler that takes the payload as T. The payload may be published as T or *T.
func subscribeTyped[T any](b *Bus, subscriber string, topic EventType, handler func(context.Context, T) error) {
	b.Subscribe(subscriber, topic, func(ctx context.Context, event Event, value interface{}) error {
		switch v := value.(type) {
		case T:
			return handler(ctx, v)
		case *T:
			if v != nil {
				return handler(ctx, *v)
			}
		default:
			slog.Warn("event payload has unexpected type", "type", event.Type, "got", value)
		}
		return nil
	})
}

// Publish sends an event to all subscribers of the given topic.
// Each handler runs in its own goroutine so publishers never block.
func (b *Bus) Publish(ctx context.Context, topic EventType, payload any) {
	ctx, span := tracing.Start(ctx, "bus.publish:"+string(topic))
	defer span.End()

	b.mu.RLock()
	subs := make([]subscription, len(b.subs[string(topic)]))
	copy(subs, b.subs[string(topic)])
	b.mu.RUnlock()

	event := Event{Type: topic, Payload: payload}
	detached := context.WithoutCancel(ctx)
	for _, s := range subs {
		go safeDispatch(detached, s, event, payload)
	}
}

func safeDispatch(ctx context.Context, s subscription, event Event, payload any) {
	ctx, span := tracing.Start(ctx, "bus.subscribe:"+string(event.Type))
	defer span.End()

	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("subscriber panic: %v", r)
			span.RecordError(err)
			span.SetStatus(codes.Error, "subscriber panic")
			slog.Error("event subscriber panicked",
				"topic", event.Type,
				"subscriber", s.subscriber,
				"panic", r,
				"stack", string(debug.Stack()),
			)
		}
	}()
	if err := s.handler(ctx, event, payload); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "event subscriber failed",
			"topic", event.Type,
			"subscriber", s.subscriber,
			"error", err,
		)
	}
}

// PublishSync is like Publish but calls handlers in the caller's goroutine.
// Useful when you need to guarantee ordering or want backpressure.
// It stops at the first handler error and returns it.
func (b *Bus) PublishSync(ctx context.Context, topic EventType, payload any) error {
	ctx, span := tracing.Start(ctx, "bus.publish:"+string(topic))
	defer span.End()

	b.mu.RLock()
	subs := make([]subscription, len(b.subs[string(topic)]))
	copy(subs, b.subs[string(topic)])
	b.mu.RUnlock()

	event := Event{Type: topic, Payload: payload}
	for _, s := range subs {
		if err := dispatchSync(ctx, s, event, payload); err != nil {
			return err
		}
	}
	return nil
}

func dispatchSync(ctx context.Context, s subscription, event Event, payload any) error {
	ctx, span := tracing.Start(ctx, "bus.subscribe:"+string(event.Type))
	defer span.End()

	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("subscriber panic: %v", r)
			span.RecordError(err)
			span.SetStatus(codes.Error, "subscriber panic")
			slog.Error("event subscriber panicked",
				"topic", event.Type,
				"subscriber", s.subscriber,
				"panic", r,
				"stack", string(debug.Stack()),
			)
			panic(r)
		}
	}()
	err := s.handler(ctx, event, payload)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}
