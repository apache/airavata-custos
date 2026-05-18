package events

import (
	"log/slog"

	"github.com/apache/airavata-custos/pkg/models"
)

// OrganizationHandler handles organization lifecycle events with a typed payload.
type OrganizationHandler func(organization models.Organization)

// SubscribeOrganizationCreated registers a typed handler invoked whenever an
// organization::create event is published. Events with payloads that are not a
// models.Organization (or *models.Organization) are dropped with a warning log.
func (b *Bus) SubscribeOrganizationCreated(handler OrganizationHandler) {
	b.subscribeOrganization(OrganizationCreateEvent, handler)
}

// SubscribeOrganizationUpdated registers a typed handler invoked whenever an
// organization::update event is published.
func (b *Bus) SubscribeOrganizationUpdated(handler OrganizationHandler) {
	b.subscribeOrganization(OrganizationUpdateEvent, handler)
}

// SubscribeOrganizationDeleted registers a typed handler invoked whenever an
// organization::delete event is published.
func (b *Bus) SubscribeOrganizationDeleted(handler OrganizationHandler) {
	b.subscribeOrganization(OrganizationDeleteEvent, handler)
}

func (b *Bus) subscribeOrganization(topic EventType, handler OrganizationHandler) {
	b.Subscribe(topic, func(event Event, value interface{}) {
		switch o := value.(type) {
		case models.Organization:
			handler(o)
		case *models.Organization:
			if o != nil {
				handler(*o)
			}
		default:
			slog.Warn("organization event payload has unexpected type",
				"type", event.Type,
				"got", value,
			)
		}
	})
}
