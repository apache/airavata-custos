package events

import (
	"context"
	"log/slog"

	"github.com/apache/airavata-custos/pkg/models"
)

// ComputeAllocationResourceHandler handles compute allocation resource lifecycle events with a typed payload.
type ComputeAllocationResourceHandler func(ctx context.Context, resource models.ComputeAllocationResource)

// SubscribeComputeAllocationResourceCreated registers a typed handler invoked whenever a
// compute_allocation_resource::create event is published. Events with payloads that are
// not a models.ComputeAllocationResource (or *models.ComputeAllocationResource) are dropped
// with a warning log.
func (b *Bus) SubscribeComputeAllocationResourceCreated(handler ComputeAllocationResourceHandler) {
	b.subscribeComputeAllocationResource(ComputeAllocationResourceCreateEvent, handler)
}

// SubscribeComputeAllocationResourceUpdated registers a typed handler invoked whenever a
// compute_allocation_resource::update event is published.
func (b *Bus) SubscribeComputeAllocationResourceUpdated(handler ComputeAllocationResourceHandler) {
	b.subscribeComputeAllocationResource(ComputeAllocationResourceUpdateEvent, handler)
}

// SubscribeComputeAllocationResourceDeleted registers a typed handler invoked whenever a
// compute_allocation_resource::delete event is published.
func (b *Bus) SubscribeComputeAllocationResourceDeleted(handler ComputeAllocationResourceHandler) {
	b.subscribeComputeAllocationResource(ComputeAllocationResourceDeleteEvent, handler)
}

func (b *Bus) subscribeComputeAllocationResource(topic EventType, handler ComputeAllocationResourceHandler) {
	b.Subscribe(topic, func(ctx context.Context, event Event, value interface{}) {
		switch r := value.(type) {
		case models.ComputeAllocationResource:
			handler(ctx, r)
		case *models.ComputeAllocationResource:
			if r != nil {
				handler(ctx, *r)
			}
		default:
			slog.Warn("compute allocation resource event payload has unexpected type",
				"type", event.Type,
				"got", value,
			)
		}
	})
}
