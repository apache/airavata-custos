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
		subs: make(map[string][]EventSubscriberFunc),
	}
}

// Subscribe registers a handler for a given topic.
// The handler is called asynchronously (in a new goroutine) each time
// an event is published on that topic.
func (b *Bus) Subscribe(topic EventType, handler EventSubscriberFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs[string(topic)] = append(b.subs[string(topic)], handler)
}

// Publish sends an event to all subscribers of the given topic.
// Each handler runs in its own goroutine so publishers never block.
func (b *Bus) Publish(ctx context.Context, topic EventType, payload any) {
	ctx, span := tracing.Start(ctx, "bus.publish:"+string(topic))
	defer span.End()

	b.mu.RLock()
	handlers := make([]EventSubscriberFunc, len(b.subs[string(topic)]))
	copy(handlers, b.subs[string(topic)])
	b.mu.RUnlock()

	event := Event{Type: topic, Payload: payload}
	detached := context.WithoutCancel(ctx)
	for _, h := range handlers {
		go safeDispatch(detached, h, event, payload)
	}
}

func safeDispatch(ctx context.Context, h EventSubscriberFunc, event Event, payload any) {
	ctx, span := tracing.Start(ctx, "bus.subscribe:"+string(event.Type))
	defer span.End()

	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("subscriber panic: %v", r)
			span.RecordError(err)
			span.SetStatus(codes.Error, "subscriber panic")
			slog.Error("event subscriber panicked",
				"topic", event.Type,
				"panic", r,
				"stack", string(debug.Stack()),
			)
		}
	}()
	h(ctx, event, payload)
}

// PublishSync is like Publish but calls handlers in the caller's goroutine.
// Useful when you need to guarantee ordering or want backpressure.
func (b *Bus) PublishSync(ctx context.Context, topic EventType, payload any) {
	ctx, span := tracing.Start(ctx, "bus.publish:"+string(topic))
	defer span.End()

	b.mu.RLock()
	handlers := make([]EventSubscriberFunc, len(b.subs[string(topic)]))
	copy(handlers, b.subs[string(topic)])
	b.mu.RUnlock()

	event := Event{Type: topic, Payload: payload}
	for _, h := range handlers {
		dispatchSync(ctx, h, event, payload)
	}
}

func dispatchSync(ctx context.Context, h EventSubscriberFunc, event Event, payload any) {
	ctx, span := tracing.Start(ctx, "bus.subscribe:"+string(event.Type))
	defer span.End()

	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("subscriber panic: %v", r)
			span.RecordError(err)
			span.SetStatus(codes.Error, "subscriber panic")
			slog.Error("event subscriber panicked",
				"topic", event.Type,
				"panic", r,
				"stack", string(debug.Stack()),
			)
			panic(r)
		}
	}()
	h(ctx, event, payload)
}
