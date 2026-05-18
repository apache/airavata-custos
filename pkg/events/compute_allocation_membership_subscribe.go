package events

import (
	"log/slog"

	"github.com/apache/airavata-custos/pkg/models"
)

// ComputeAllocationMembershipHandler handles compute allocation membership lifecycle events with a typed payload.
type ComputeAllocationMembershipHandler func(membership models.ComputeAllocationMembership)

// SubscribeComputeAllocationMembershipCreated registers a typed handler invoked whenever a
// compute_allocation_membership::create event is published. Events with payloads that are
// not a models.ComputeAllocationMembership (or *models.ComputeAllocationMembership) are
// dropped with a warning log.
func (b *Bus) SubscribeComputeAllocationMembershipCreated(handler ComputeAllocationMembershipHandler) {
	b.subscribeComputeAllocationMembership(ComputeAllocationMembershipCreateEvent, handler)
}

// SubscribeComputeAllocationMembershipUpdated registers a typed handler invoked whenever a
// compute_allocation_membership::update event is published.
func (b *Bus) SubscribeComputeAllocationMembershipUpdated(handler ComputeAllocationMembershipHandler) {
	b.subscribeComputeAllocationMembership(ComputeAllocationMembershipUpdateEvent, handler)
}

// SubscribeComputeAllocationMembershipDeleted registers a typed handler invoked whenever a
// compute_allocation_membership::delete event is published.
func (b *Bus) SubscribeComputeAllocationMembershipDeleted(handler ComputeAllocationMembershipHandler) {
	b.subscribeComputeAllocationMembership(ComputeAllocationMembershipDeleteEvent, handler)
}

func (b *Bus) subscribeComputeAllocationMembership(topic EventType, handler ComputeAllocationMembershipHandler) {
	b.Subscribe(topic, func(event Event, value interface{}) {
		switch m := value.(type) {
		case models.ComputeAllocationMembership:
			handler(m)
		case *models.ComputeAllocationMembership:
			if m != nil {
				handler(*m)
			}
		default:
			slog.Warn("compute allocation membership event payload has unexpected type",
				"type", event.Type,
				"got", value,
			)
		}
	})
}
