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
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type ctxKey string

const testCtxKey ctxKey = "trace-id"

const topicTest EventType = "test::topic"

func TestPublishSyncPropagatesContext(t *testing.T) {
	bus := New()

	got := make(chan string, 1)
	bus.Subscribe(topicTest, func(ctx context.Context, _ Event, _ interface{}) {
		v, _ := ctx.Value(testCtxKey).(string)
		got <- v
	})

	ctx := context.WithValue(context.Background(), testCtxKey, "abc123")
	bus.PublishSync(ctx, topicTest, "payload")

	select {
	case v := <-got:
		if v != "abc123" {
			t.Fatalf("PublishSync did not propagate ctx value, got %q", v)
		}
	case <-time.After(time.Second):
		t.Fatalf("subscriber never ran")
	}
}

func TestPublishAsyncPropagatesContext(t *testing.T) {
	bus := New()

	got := make(chan string, 1)
	bus.Subscribe(topicTest, func(ctx context.Context, _ Event, _ interface{}) {
		v, _ := ctx.Value(testCtxKey).(string)
		got <- v
	})

	ctx := context.WithValue(context.Background(), testCtxKey, "def456")
	bus.Publish(ctx, topicTest, "payload")

	select {
	case v := <-got:
		if v != "def456" {
			t.Fatalf("Publish did not propagate ctx value, got %q", v)
		}
	case <-time.After(time.Second):
		t.Fatalf("subscriber never ran")
	}
}

func TestPublishAsyncDetachesCancellation(t *testing.T) {
	bus := New()

	started := make(chan struct{})
	done := make(chan error, 1)
	bus.Subscribe(topicTest, func(ctx context.Context, _ Event, _ interface{}) {
		close(started)
		select {
		case <-ctx.Done():
			done <- ctx.Err()
		case <-time.After(200 * time.Millisecond):
			done <- nil
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	bus.Publish(ctx, topicTest, "payload")

	<-started
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected subscriber ctx to be detached from cancellation, got err=%v", err)
		}
	case <-time.After(time.Second):
		t.Fatalf("subscriber never finished")
	}
}

func TestPublishSyncPanicPropagatesToCaller(t *testing.T) {
	bus := New()

	bus.Subscribe(topicTest, func(context.Context, Event, interface{}) {
		panic("boom")
	})

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		bus.PublishSync(context.Background(), topicTest, nil)
	}()

	if recovered == nil {
		t.Fatalf("expected sync publish to surface subscriber panic to caller")
	}
}

func TestPublishAsyncSubscriberPanicDoesNotKillOthers(t *testing.T) {
	bus := New()

	var ran atomic.Int32
	var wg sync.WaitGroup
	wg.Add(2)
	bus.Subscribe(topicTest, func(context.Context, Event, interface{}) {
		defer wg.Done()
		ran.Add(1)
		panic("boom")
	})
	bus.Subscribe(topicTest, func(context.Context, Event, interface{}) {
		defer wg.Done()
		ran.Add(1)
	})

	bus.Publish(context.Background(), topicTest, nil)

	finished := make(chan struct{})
	go func() {
		wg.Wait()
		close(finished)
	}()

	select {
	case <-finished:
	case <-time.After(time.Second):
	}

	if got := ran.Load(); got != 2 {
		t.Fatalf("expected both subscribers to run, got %d", got)
	}
}

type recordingProcessor struct {
	mu    sync.Mutex
	spans []sdktrace.ReadOnlySpan
}

func (p *recordingProcessor) OnStart(context.Context, sdktrace.ReadWriteSpan) {}
func (p *recordingProcessor) OnEnd(s sdktrace.ReadOnlySpan) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.spans = append(p.spans, s)
}
func (p *recordingProcessor) Shutdown(context.Context) error   { return nil }
func (p *recordingProcessor) ForceFlush(context.Context) error { return nil }

func (p *recordingProcessor) findByName(name string) sdktrace.ReadOnlySpan {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, s := range p.spans {
		if s.Name() == name {
			return s
		}
	}
	return nil
}

func TestSafeDispatchSetsSpanErrorOnPanic(t *testing.T) {
	rec := &recordingProcessor{}
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec))
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { otel.SetTracerProvider(prev) })

	bus := New()
	bus.Subscribe(topicTest, func(context.Context, Event, interface{}) {
		panic("boom")
	})

	bus.Publish(context.Background(), topicTest, nil)

	deadline := time.Now().Add(time.Second)
	var sub sdktrace.ReadOnlySpan
	for time.Now().Before(deadline) {
		_ = tp.ForceFlush(context.Background())
		sub = rec.findByName("bus.subscribe:" + string(topicTest))
		if sub != nil {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	if sub == nil {
		var names []string
		for _, s := range rec.spans {
			names = append(names, s.Name())
		}
		t.Fatalf("did not see bus.subscribe span; got names=%v", names)
	}
	if got := sub.Status().Code; got != codes.Error {
		t.Fatalf("expected bus.subscribe span status=Error, got %v", got)
	}
}
