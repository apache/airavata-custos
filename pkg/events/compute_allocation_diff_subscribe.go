package events

import (
	"context"
	"log/slog"

	"github.com/apache/airavata-custos/pkg/models"
)

// ComputeAllocationDiffHandler handles compute allocation diff lifecycle events with a typed payload.
type ComputeAllocationDiffHandler func(ctx context.Context, diff models.ComputeAllocationDiff)

// SubscribeComputeAllocationDiffCreated registers a typed handler invoked whenever a
// compute_allocation_diff::create event is published. Events with payloads that are
// not a models.ComputeAllocationDiff (or *models.ComputeAllocationDiff) are dropped
// with a warning log.
func (b *Bus) SubscribeComputeAllocationDiffCreated(handler ComputeAllocationDiffHandler) {
	b.subscribeComputeAllocationDiff(ComputeAllocationDiffCreateEvent, handler)
}

// SubscribeComputeAllocationDiffUpdated registers a typed handler invoked whenever a
// compute_allocation_diff::update event is published.
func (b *Bus) SubscribeComputeAllocationDiffUpdated(handler ComputeAllocationDiffHandler) {
	b.subscribeComputeAllocationDiff(ComputeAllocationDiffUpdateEvent, handler)
}

// SubscribeComputeAllocationDiffDeleted registers a typed handler invoked whenever a
// compute_allocation_diff::delete event is published.
func (b *Bus) SubscribeComputeAllocationDiffDeleted(handler ComputeAllocationDiffHandler) {
	b.subscribeComputeAllocationDiff(ComputeAllocationDiffDeleteEvent, handler)
}

func (b *Bus) subscribeComputeAllocationDiff(topic EventType, handler ComputeAllocationDiffHandler) {
	b.Subscribe(topic, func(ctx context.Context, event Event, value interface{}) {
		switch d := value.(type) {
		case models.ComputeAllocationDiff:
			handler(ctx, d)
		case *models.ComputeAllocationDiff:
			if d != nil {
				handler(ctx, *d)
			}
		default:
			slog.Warn("compute allocation diff event payload has unexpected type",
				"type", event.Type,
				"got", value,
			)
		}
	})
}
