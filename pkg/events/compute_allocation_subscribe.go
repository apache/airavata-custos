package events

import (
	"log/slog"

	"github.com/apache/airavata-custos/pkg/models"
)

// ComputeAllocationHandler handles compute allocation lifecycle events with a typed payload.
type ComputeAllocationHandler func(allocation models.ComputeAllocation)

// SubscribeComputeAllocationCreated registers a typed handler invoked whenever a
// compute_allocation::create event is published. Events with payloads that are
// not a models.ComputeAllocation (or *models.ComputeAllocation) are dropped
// with a warning log.
func (b *Bus) SubscribeComputeAllocationCreated(handler ComputeAllocationHandler) {
	b.subscribeComputeAllocation(ComputeAllocationCreateEvent, handler)
}

// SubscribeComputeAllocationUpdated registers a typed handler invoked whenever a
// compute_allocation::update event is published.
func (b *Bus) SubscribeComputeAllocationUpdated(handler ComputeAllocationHandler) {
	b.subscribeComputeAllocation(ComputeAllocationUpdateEvent, handler)
}

// SubscribeComputeAllocationDeleted registers a typed handler invoked whenever a
// compute_allocation::delete event is published.
func (b *Bus) SubscribeComputeAllocationDeleted(handler ComputeAllocationHandler) {
	b.subscribeComputeAllocation(ComputeAllocationDeleteEvent, handler)
}

func (b *Bus) subscribeComputeAllocation(topic EventType, handler ComputeAllocationHandler) {
	b.Subscribe(topic, func(event Event, value interface{}) {
		switch a := value.(type) {
		case models.ComputeAllocation:
			handler(a)
		case *models.ComputeAllocation:
			if a != nil {
				handler(*a)
			}
		default:
			slog.Warn("compute allocation event payload has unexpected type",
				"type", event.Type,
				"got", value,
			)
		}
	})
}
