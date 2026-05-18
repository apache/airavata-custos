package events

import (
	"log/slog"

	"github.com/apache/airavata-custos/pkg/models"
)

// UserDNHandler handles user DN lifecycle events with a typed payload.
type UserDNHandler func(d models.UserDN)

// SubscribeUserDNCreated registers a typed handler invoked whenever a
// user_dn::create event is published.
func (b *Bus) SubscribeUserDNCreated(handler UserDNHandler) {
	b.subscribeUserDN(UserDNCreateEvent, handler)
}

// SubscribeUserDNDeleted registers a typed handler invoked whenever a
// user_dn::delete event is published.
func (b *Bus) SubscribeUserDNDeleted(handler UserDNHandler) {
	b.subscribeUserDN(UserDNDeleteEvent, handler)
}

func (b *Bus) subscribeUserDN(topic EventType, handler UserDNHandler) {
	b.Subscribe(topic, func(event Event, value interface{}) {
		switch d := value.(type) {
		case models.UserDN:
			handler(d)
		case *models.UserDN:
			if d != nil {
				handler(*d)
			}
		default:
			slog.Warn("user dn event payload has unexpected type",
				"type", event.Type,
				"got", value,
			)
		}
	})
}
