package events

import (
	"log/slog"

	"github.com/apache/airavata-custos/pkg/models"
)

// ExternalIdentityHandler handles external identity lifecycle events with a typed payload.
type ExternalIdentityHandler func(ext models.ExternalIdentity)

// SubscribeExternalIdentityCreated registers a typed handler invoked whenever an
// external_identity::create event is published.
func (b *Bus) SubscribeExternalIdentityCreated(handler ExternalIdentityHandler) {
	b.subscribeExternalIdentity(ExternalIdentityCreateEvent, handler)
}

// SubscribeExternalIdentityUpdated registers a typed handler invoked whenever an
// external_identity::update event is published.
func (b *Bus) SubscribeExternalIdentityUpdated(handler ExternalIdentityHandler) {
	b.subscribeExternalIdentity(ExternalIdentityUpdateEvent, handler)
}

// SubscribeExternalIdentityDeleted registers a typed handler invoked whenever an
// external_identity::delete event is published.
func (b *Bus) SubscribeExternalIdentityDeleted(handler ExternalIdentityHandler) {
	b.subscribeExternalIdentity(ExternalIdentityDeleteEvent, handler)
}

func (b *Bus) subscribeExternalIdentity(topic EventType, handler ExternalIdentityHandler) {
	b.Subscribe(topic, func(event Event, value interface{}) {
		switch e := value.(type) {
		case models.ExternalIdentity:
			handler(e)
		case *models.ExternalIdentity:
			if e != nil {
				handler(*e)
			}
		default:
			slog.Warn("external identity event payload has unexpected type",
				"type", event.Type,
				"got", value,
			)
		}
	})
}
